package command_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/extendio/xfs/utils"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/internal/helpers"
)

type rootTE struct {
	given       string
	commandLine []string
}

var _ = Describe("RootCmd", Ordered, func() {
	var (
		repo     string
		l10nPath string

		tester helpers.CommandTester
	)

	BeforeAll(func() {
		repo = helpers.Repo("../..")
		l10nPath = helpers.Path(repo, "test/data/l10n")
		Expect(utils.FolderExists(l10nPath)).To(BeTrue())

	})

	BeforeEach(func() {
		bootstrap := command.Bootstrap{}
		tester = helpers.CommandTester{
			Args: []string{"./"},
			Root: bootstrap.Root(func(co *command.ConfigureOptions) {
				co.Detector = &DetectorStub{}
				co.Executor = &ExecutorStub{
					Name: "magick",
				}
			}),
		}
	})

	DescribeTable("dummy magick",
		func(entry *rootTE) {
			bootstrap := command.Bootstrap{}
			tester = helpers.CommandTester{
				Args: entry.commandLine,
				Root: bootstrap.Root(func(co *command.ConfigureOptions) {
					co.Detector = &DetectorStub{}
					co.Executor = &ExecutorStub{
						Name: "magick",
					}
				}),
			}

			_, err := tester.Execute()
			Expect(err).Error().To(BeNil())
		},
		func(entry *rootTE) string {
			return fmt.Sprintf("🧪 given: '%v', should: execute", entry.given)
		},

		Entry(
			nil, &rootTE{
				given:       "just a positional",
				commandLine: []string{"./"},
			},
		),

		Entry(
			nil, &rootTE{
				given:       "a family defined switch (--dry-run)",
				commandLine: []string{"./", "--dry-run"},
			},
		),
	)
})
