package command_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/internal/helpers"
	"github.com/snivilised/pixa/src/internal/matchers"
)

type rootTE struct {
	given       string
	commandLine []string
}

var _ = Describe("RootCmd", Ordered, func() {
	var (
		repo       string
		l10nPath   string
		configPath string
		nfs        storage.VirtualFS
		tester     helpers.CommandTester
	)

	BeforeAll(func() {
		nfs = storage.UseNativeFS()
		repo = helpers.Repo(filepath.Join("..", "..", ".."))

		l10nPath = helpers.Path(repo, filepath.Join("test", "data", "l10n"))
		Expect(matchers.AsDirectory(l10nPath)).To(matchers.ExistInFS(nfs))

		configPath = filepath.Join(repo, "test", "data", "configuration")
		Expect(matchers.AsDirectory(configPath)).To(matchers.ExistInFS(nfs))

		if err := helpers.UseI18n(l10nPath); err != nil {
			Fail(err.Error())
		}
	})

	BeforeEach(func() {
		bootstrap := command.Bootstrap{
			Vfs: nfs,
		}
		tester = helpers.CommandTester{
			Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
				co.Detector = &DetectorStub{}
				co.Program = &ExecutorStub{
					Name: "magick",
				}
				co.Config.Name = helpers.PixaConfigTestFilename
				co.Config.ConfigPath = configPath
			}),
		}
	})

	DescribeTable("dummy magick",
		func(entry *rootTE) {
			tester.Args = entry.commandLine
			_, err := tester.Execute()
			Expect(err).Error().To(BeNil())
		},
		func(entry *rootTE) string {
			return fmt.Sprintf("🧪 given: '%v', should: execute", entry.given)
		},

		XEntry(
			nil, &rootTE{
				given:       "just a positional",
				commandLine: []string{"./"},
			},
		),

		XEntry(
			nil, &rootTE{
				given:       "a family defined switch (--dry-run)",
				commandLine: []string{"./", "--dry-run"},
			},
		),
	)
})
