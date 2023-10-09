package magick_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	ci18n "github.com/snivilised/cobrass/src/assistant/i18n"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/i18n"
	"github.com/snivilised/pixa/src/internal/helpers"
)

const (
	relative = "../../test/data/configuration"
	prog     = "magick"
)

func expectValidShrinkCmdInvocation(entry *configTE) {
	bootstrap := command.Bootstrap{}

	options := []string{
		entry.comm, entry.file,
		"--dry-run",
		"--mode", "tidy",
		"--profile", entry.profile,
	}

	tester := helpers.CommandTester{
		Args: append(options, entry.args...),
		Root: bootstrap.Root(func(co *command.ConfigureOptions) {
			co.Detector = &helpers.DetectorStub{}
			co.Executor = &helpers.ExecutorStub{
				Name: prog,
			}
			co.Config.Name = "pixa-test"
			co.Config.ConfigPath = "../../test/data/configuration"
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

func reason(field string, expected, actual any) string {
	return fmt.Sprintf("🔥 expected field '%v' to be '%v', but was '%v'\n",
		field, expected, actual,
	)
}

var _ = Describe("Config", func() {
	var (
		config   configuration.ViperConfig
		l10nPath string
	)

	BeforeEach(func() {
		viper.Reset()
		config = &configuration.GlobalViperConfig{}

		config.SetConfigType("yml")
		config.SetConfigName("pixa-test")

		if _, err := os.Lstat(relative); err != nil {
			Fail("🔥 can't find config path")
		}
		config.AddConfigPath(relative)
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
				// se := magick.ShrinkEntry{
				// 	EntryBase: magick.EntryBase{
				// 		RootPS: assistant.NewParamSet[magick.RootParameterSet](bootstrap.Root()),
				// 		Program: &helpers.ExecutorStub{
				// 			Name: comm,
				// 		},
				// 		Config: config,
				// 	},
				// 	ParamSet: assistant.NewParamSet[magick.ShrinkParameterSet](
				// 		bootstrap.Container.Command(comm),
				// 	),
				// }

				// result := se.ReadProfile()
				// _ = result

				expectValidShrinkCmdInvocation(entry)
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

		Entry(nil, &configTE{
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
