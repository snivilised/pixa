package command_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/internal/helpers"
	"github.com/snivilised/pixa/src/internal/matchers"

	"github.com/snivilised/extendio/xfs/storage"
)

type commandTE struct {
	message    string
	args       []string
	configPath string
}

type shrinkTE struct {
	commandTE
	directory string
}

func expectValidShrinkCmdInvocation(vfs storage.VirtualFS, entry *shrinkTE) {
	bootstrap := command.Bootstrap{
		Vfs: vfs,
	}

	// we also prepend the directory name to the command line
	//
	options := append([]string{helpers.ShrinkCommandName, entry.directory}, []string{
		"--dry-run", "--mode", "tidy",
	}...)

	tester := helpers.CommandTester{
		Args: append(options, entry.args...),
		Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
			co.Detector = &DetectorStub{}
			co.Program = &ExecutorStub{
				Name: helpers.ProgName,
			}
			co.Config.Name = helpers.PixaConfigTestFilename
			co.Config.ConfigPath = entry.configPath
		}),
	}

	_, err := tester.Execute()
	Expect(err).Error().To(BeNil(),
		"should pass validation due to all flag being valid",
	)
}

var _ = Describe("ShrinkCmd", Ordered, func() {
	var (
		repo       string
		l10nPath   string
		configPath string
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
		if err := helpers.UseI18n(l10nPath); err != nil {
			Fail(err.Error())
		}
	})

	DescribeTable("ShrinkCmd",
		func(entry *shrinkTE) {
			// set directory here, because during discovery phase of unit test ,
			// l10nPath is not set, so we can't set it inside the Entry
			//
			entry.directory = l10nPath
			entry.configPath = configPath
			expectValidShrinkCmdInvocation(nfs, entry)
		},
		func(entry *shrinkTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v'", entry.message)
		},

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "vanilla with long form options",
				args: []string{
					"--strip", "--interlace", "plane", "--quality", "85", "--cpu",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "vanilla with short form options",
				args: []string{
					"-s", "-i", "plane", "-q", "85",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "gaussian-blur with long form options",
				args: []string{
					"--strip", "--interlace", "plane", "--quality", "85",
					"--gaussian-blur", "0.85",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "blur with short form options",
				args: []string{
					"-s", "-i", "plane", "-q", "85",
					"-b", "0.85",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "sampling factor with long form options",
				args: []string{
					"--strip", "--interlace", "plane", "--quality", "85",
					"--sampling-factor", "4:2:0",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "sampling factor with short form options",
				args: []string{
					"-s", "-i", "Plane", "-q", "85",
					"-f", "4:2:0",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "with long form glob filtering options options",
				args: []string{
					// "--folders-gb", "A*",
					"--files-gb", "*.jpg",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "with short form regex filtering options options",
				args: []string{
					// "--folders-rx", "^A",
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
				directory: l10nPath,
				commandTE: commandTE{
					message: "with general long form parameters",
					args: []string{
						"--output", l10nPath,
					},
					configPath: configPath,
				},
			}

			expectValidShrinkCmdInvocation(nfs, entry)
		})

		It("🧪 should: execute successfully", func() {
			entry := &shrinkTE{
				directory: l10nPath,
				commandTE: commandTE{
					message: "with general long form parameters",
					args: []string{
						"--trash", l10nPath,
					},
					configPath: configPath,
				},
			}

			expectValidShrinkCmdInvocation(nfs, entry)
		})
	})

	When("with general short form parameters", func() {
		It("🧪 should: execute successfully", func() {
			entry := &shrinkTE{
				directory: l10nPath,
				commandTE: commandTE{
					message: "with general short form parameters",
					args: []string{
						"-o", l10nPath,
					},
					configPath: configPath,
				},
			}

			expectValidShrinkCmdInvocation(nfs, entry)
		})

		It("🧪 should: execute successfully", func() {
			entry := &shrinkTE{
				directory: l10nPath,
				commandTE: commandTE{
					message: "with general short form parameters",
					args: []string{
						"-t", l10nPath,
					},
					configPath: configPath,
				},
			}

			expectValidShrinkCmdInvocation(nfs, entry)
		})
	})
})
