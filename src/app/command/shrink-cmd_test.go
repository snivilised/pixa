package command_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/pixa/src/app/cfg"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/internal/helpers"

	"github.com/snivilised/extendio/xfs/storage"
)

var (
	_ common.ProfilesConfig = &cfg.MsProfilesConfig{}
	_ common.SamplerConfig  = &cfg.MsSamplerConfig{}
)

const (
	BackyardWorldsPlanet9Scan01 = "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01"
)

type commandTE struct {
	message     string
	args        []string
	trashFlag   string
	trashValue  string
	outputFlag  string
	outputValue string
	configPath  string
}

type shrinkTE struct {
	commandTE
	directory string
}

func expectValidShrinkCmdInvocation(vfs storage.VirtualFS, entry *shrinkTE, root string) {
	bootstrap := command.Bootstrap{
		Vfs: vfs,
	}

	directory := helpers.Path(root, entry.directory)
	args := append([]string{common.Definitions.Commands.Shrink, directory}, []string{
		"--dry-run",
	}...)

	if entry.outputFlag != "" && entry.outputValue != "" {
		output := helpers.Path(root, entry.outputValue)
		args = append(args, entry.outputFlag, output)
	}

	if entry.trashFlag != "" && entry.trashValue != "" {
		trash := helpers.Path(root, entry.trashValue)
		args = append(args, entry.trashFlag, trash)
	}

	tester := helpers.CommandTester{
		Args: append(args, entry.args...),
		Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
			co.Detector = &DetectorStub{}
			co.Config.Name = common.Definitions.Pixa.ConfigTestFilename
			co.Config.ConfigPath = entry.configPath
			co.Config.Viper = &configuration.GlobalViperConfig{}
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
		root       string
		vfs        storage.VirtualFS
	)

	BeforeAll(func() {
		repo = helpers.Repo("")
		l10nPath = helpers.Path(repo, "test/data/l10n")
		configPath = helpers.Path(repo, "test/data/configuration")
	})

	BeforeEach(func() {
		vfs, root = helpers.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, helpers.Silent,
		)
	})

	DescribeTable("ShrinkCmd",
		func(entry *shrinkTE) {
			entry.directory = BackyardWorldsPlanet9Scan01
			entry.configPath = configPath
			expectValidShrinkCmdInvocation(vfs, entry, root)
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
					"--files", "*.jpg",
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

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "vanilla with long form no-recurse",
				args: []string{
					"--strip",
					"--interlace", "plane",
					"--quality", "85",
					"--no-recurse",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "vanilla with short form no-recurse",
				args: []string{
					"--strip",
					"--interlace", "plane",
					"--quality", "85",
					"-N",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "vanilla with depth",
				args: []string{
					"--strip",
					"--interlace", "plane",
					"--quality", "85",
					"--depth", "1",
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
				directory: BackyardWorldsPlanet9Scan01,
				commandTE: commandTE{
					message:     "with general long form parameters",
					args:        []string{},
					outputFlag:  "--output",
					outputValue: "output",
					configPath:  configPath,
				},
			}

			expectValidShrinkCmdInvocation(vfs, entry, root)
		})

		It("🧪 should: execute successfully", func() {
			entry := &shrinkTE{
				directory: BackyardWorldsPlanet9Scan01,
				commandTE: commandTE{
					message:    "with general long form parameters",
					args:       []string{},
					trashFlag:  "--trash",
					trashValue: "discard",
					configPath: configPath,
				},
			}

			expectValidShrinkCmdInvocation(vfs, entry, root)
		})
	})

	When("with general short form parameters", func() {
		It("🧪 should: execute successfully", func() {
			entry := &shrinkTE{
				directory: BackyardWorldsPlanet9Scan01,
				commandTE: commandTE{
					message:     "with general short form parameters",
					args:        []string{},
					outputFlag:  "-o",
					outputValue: "output",
					configPath:  configPath,
				},
			}

			expectValidShrinkCmdInvocation(vfs, entry, root)
		})

		It("🧪 should: execute successfully", func() {
			entry := &shrinkTE{
				directory: BackyardWorldsPlanet9Scan01,
				commandTE: commandTE{
					message:    "with general short form parameters",
					args:       []string{},
					trashFlag:  "-t",
					trashValue: "discard",
					configPath: configPath,
				},
			}

			expectValidShrinkCmdInvocation(vfs, entry, root)
		})
	})
})
