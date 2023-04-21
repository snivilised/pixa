package command_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ci18n "github.com/snivilised/cobrass/src/assistant/i18n"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/i18n"
	"github.com/snivilised/pixa/src/internal/helpers"

	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/extendio/xfs/utils"
)

type commandTE struct {
	message string
	args    []string
}

type shrinkTE struct {
	commandTE
}

func coreShrinkCmdTest(entry *shrinkTE) {
	bootstrap := command.Bootstrap{
		Detector: &DetectorStub{},
	}

	const (
		prog = "shrink"
	)

	options := append([]string{prog}, []string{
		"--preview", "--mode", "tidy",
	}...)

	tester := helpers.CommandTester{
		Args: append(options, entry.args...),
		Root: bootstrap.Root(),
	}

	_, err := tester.Execute()
	Expect(err).Error().To(BeNil(),
		"should pass validation due to all flag being valid",
	)
}

var _ = Describe("ShrinkCmd", Ordered, func() {
	var (
		repo     string
		l10nPath string
	)

	BeforeAll(func() {
		repo = helpers.Repo("../../..")
		l10nPath = helpers.Path(repo, "src/test/data/l10n")
		Expect(utils.FolderExists(l10nPath)).To(BeTrue(),
			fmt.Sprintf("💥 l10Path: '%v' does not exist", l10nPath),
		)
	})

	BeforeEach(func() {
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

	DescribeTable("ShrinkCmd",
		func(entry *shrinkTE) {
			coreShrinkCmdTest(entry)
		},
		func(entry *shrinkTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v'", entry.message)
		},

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "vanilla with long form options",
				args: []string{
					"source.jpg", "--strip", "--interlace", "plane", "--quality", "85", "result.jpg",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "vanilla with short form options",
				args: []string{
					"source.jpg", "-s", "-i", "plane", "-q", "85", "result.jpg",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "blur with long form options",
				args: []string{
					"source.jpg", "result.jpg", "--strip", "--interlace", "plane", "--quality", "85",
					"--gaussian-blur", "0.85",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "blur with short form options",
				args: []string{
					"source.jpg", "result.jpg", "-s", "-i", "plane", "-q", "85",
					"-b", "0.85",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "sampling factor with long form options",
				args: []string{
					"source.jpg", "result.jpg", "--strip", "--interlace", "plane", "--quality", "85",
					"--sampling-factor", "4:2:0",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "sampling factor with short form options",
				args: []string{
					"source.jpg", "result.jpg", "-s", "-i", "Plane", "-q", "85",
					"-f", "4:2:0",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "with long form glob filtering options options",
				args: []string{
					"source.jpg", "result.jpg",
					"--folder-gb", "A*",
					"--files-gb", "*.jpg",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "with short form regex filtering options options",
				args: []string{
					"source.jpg", "result.jpg",
					"--folder-rx", "^A",
					"--files-rx", "\\.jpg$",
				},
			},
		}),
	)

	// NB: these tests are required because state does not work with
	// DescribeTable. (eg l10nPath is not set within a table entry)
	//
	When("with general long form parameters", func() {
		It("🧪 should: execute successfully", func() {
			entry := &shrinkTE{
				commandTE: commandTE{
					message: "with general long form parameters",
					args: []string{
						"source.jpg", "result.jpg",
						"--mirror-path", l10nPath,
					},
				},
			}

			coreShrinkCmdTest(entry)
		})
	})
})
