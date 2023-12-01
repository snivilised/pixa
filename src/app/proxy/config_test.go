package proxy_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	ci18n "github.com/snivilised/cobrass/src/assistant/i18n"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/i18n"
	"github.com/snivilised/pixa/src/internal/helpers"
	"github.com/snivilised/pixa/src/internal/matchers"
)

func expectValidShrinkCmdInvocation(vfs storage.VirtualFS, entry *configTE) {
	bootstrap := command.Bootstrap{
		Vfs: vfs,
	}

	options := []string{
		entry.comm, entry.file,
		"--dry-run",
		"--mode", "tidy",
		"--profile", entry.profile,
	}

	repo := helpers.Repo(filepath.Join("..", "..", ".."))
	configPath := filepath.Join(repo, "test", "data", "configuration")
	tester := helpers.CommandTester{
		Args: append(options, entry.args...),
		Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
			co.Detector = &helpers.DetectorStub{}
			co.Program = &helpers.ExecutorStub{
				Name: helpers.ProgName,
			}
			co.Config.Name = helpers.PixaConfigTestFilename
			co.Config.ConfigPath = configPath
		}),
	}

	_, _ = tester.Execute()
}

type configTE struct {
	message  string
	comm     string
	file     string
	args     []string
	profile  string
	expected any
	actual   func(entry *configTE) any
	assert   func(entry *configTE, actual any)
}

var _ = Describe("Config", Ordered, func() {
	var (
		repo       string
		l10nPath   string
		configPath string
		config     configuration.ViperConfig
		nfs        storage.VirtualFS
	)

	BeforeAll(func() {
		nfs = storage.UseNativeFS()
		repo = helpers.Repo(filepath.Join("..", "..", ".."))

		l10nPath = helpers.Path(repo, filepath.Join("test", "data", "l10n"))
		Expect(matchers.AsDirectory(l10nPath)).To(matchers.ExistInFS(nfs))

		configPath = filepath.Join(repo, "test", "data", "configuration")
		Expect(matchers.AsDirectory(configPath)).To(matchers.ExistInFS(nfs))
	})

	BeforeEach(func() {
		viper.Reset()
		config = &configuration.GlobalViperConfig{}

		config.SetConfigType(helpers.PixaConfigType)
		config.SetConfigName(helpers.PixaConfigTestFilename)

		config.AddConfigPath(configPath)
		if err := config.ReadInConfig(); err != nil {
			Fail(fmt.Sprintf("🔥 can't read config (err: '%v')", err))
		}

		err := xi18n.Use(func(uo *xi18n.UseOptions) {
			uo.From = xi18n.LoadFrom{
				Path: l10nPath,
				Sources: xi18n.TranslationFiles{
					i18n.PixaSourceID: xi18n.TranslationSource{
						Name: "dummy-cobrass",
					},

					ci18n.CobrassSourceID: xi18n.TranslationSource{
						Name: "dummy-cobrass",
					},
				},
			}
		})

		if err != nil {
			Fail(err.Error())
		}
	})

	DescribeTable("profile",
		func(entry *configTE) {
			if entry.assert == nil {
				actual := entry.actual(entry)
				_ = actual

				Expect(1).To(Equal(1))
				expectValidShrinkCmdInvocation(nfs, entry)
			} else {
				actual := entry.actual(entry)
				entry.assert(entry, actual)
			}
		},
		func(entry *configTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should access profile: '%v'",
				entry.message, entry.profile,
			)
		},

		XEntry(nil, &configTE{
			message:  "adaptive",
			comm:     "shrink",
			file:     "cover.nfr.lana-del-rey.jpg",
			args:     []string{},
			profile:  "adaptive",
			expected: 42,
			actual: func(e *configTE) any {
				return config.Get(e.profile)
			},
		}),
	)
})
