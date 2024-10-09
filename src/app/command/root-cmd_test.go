package command_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // foo
	. "github.com/onsi/gomega"    //nolint:revive // foo

	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/internal/helpers"
	lab "github.com/snivilised/pixa/src/internal/laboratory"
	"github.com/snivilised/traverse/lfs"
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
		FS         lfs.TraverseFS

		tester lab.CommandTester
	)

	BeforeAll(func() {
		Expect(lab.UseI18n(l10nPath)).To(Succeed())

		repo = helpers.Repo("")
		l10nPath = lab.Path(repo, "test/data/l10n")
		configPath = lab.Path(repo, "test/data/configuration")
	})

	BeforeEach(func() {
		FS, _ = lab.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, helpers.Silent,
		)

		bootstrap := command.Bootstrap{
			FS: FS,
		}
		tester = lab.CommandTester{
			Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
				co.Detector = &lab.DetectorStub{}
				co.Config.Name = common.Definitions.Pixa.ConfigTestFilename
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
				commandLine: []string{"./", "--no-tui"},
			},
		),

		XEntry(
			nil, &rootTE{
				given:       "a family defined switch (--dry-run)",
				commandLine: []string{"./", "--dry-run", "--no-tui"},
			},
		),
	)
})
