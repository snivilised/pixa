package command_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/internal/helpers"
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
		vfs        storage.VirtualFS

		tester helpers.CommandTester
	)

	BeforeAll(func() {
		repo = helpers.Repo(filepath.Join("..", "..", ".."))
		l10nPath = helpers.Path(repo, "test/data/l10n")
		configPath = helpers.Path(repo, "test/data/configuration")
	})

	BeforeEach(func() {
		vfs, _, _ = helpers.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, helpers.Silent,
		)

		bootstrap := command.Bootstrap{
			Vfs: vfs,
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
