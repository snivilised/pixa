package command_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	cmocks "github.com/snivilised/cobrass/src/assistant/mocks"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/mocks"
	"github.com/snivilised/pixa/src/internal/helpers"
	"go.uber.org/mock/gomock"

	"github.com/snivilised/extendio/xfs/storage"
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

func expectValidShrinkCmdInvocation(vfs storage.VirtualFS, entry *shrinkTE, root string,
	config configuration.ViperConfig,
) {
	bootstrap := command.Bootstrap{
		Vfs: vfs,
	}

	directory := helpers.Path(root, entry.directory)
	args := append([]string{helpers.ShrinkCommandName, directory}, []string{
		"--dry-run", "--mode", "tidy",
	}...)

	if entry.outputFlag != "" && entry.outputValue != "" {
		output := helpers.Path(root, entry.outputValue)
		args = append(args, entry.outputFlag, output)
	}

	if entry.trashFlag != "" && entry.trashValue != "" {
		trash := helpers.Path(root, entry.trashValue)
		args = append(args, entry.trashFlag, trash)
	}

	var (
		ctrl               = gomock.NewController(GinkgoT())
		mockViperConfig    = cmocks.NewMockViperConfig(ctrl)
		mockProfilesReader = mocks.NewMockProfilesConfigReader(ctrl)
		mockSchemesReader  = mocks.NewMockSchemesConfigReader(ctrl)
		mockSamplerReader  = mocks.NewMockSamplerConfigReader(ctrl)
		mockAdvancedReader = mocks.NewMockAdvancedConfigReader(ctrl)
	)

	helpers.DoMockReadInConfig(mockViperConfig)
	helpers.DoMockConfigs(config,
		mockProfilesReader, mockSchemesReader, mockSamplerReader, mockAdvancedReader,
	)

	tester := helpers.CommandTester{
		Args: append(args, entry.args...),
		Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
			co.Detector = &DetectorStub{}
			co.Program = &ExecutorStub{
				Name: helpers.ProgName,
			}
			co.Config.Name = helpers.PixaConfigTestFilename
			co.Config.ConfigPath = entry.configPath

			co.Viper = &configuration.GlobalViperConfig{}
			co.Config.Readers = command.ConfigReaders{
				Profiles: mockProfilesReader,
				Schemes:  mockSchemesReader,
				Sampler:  mockSamplerReader,
				Advanced: mockAdvancedReader,
			}
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
		config     configuration.ViperConfig
	)

	BeforeAll(func() {
		repo = helpers.Repo(filepath.Join("..", "..", ".."))
		l10nPath = helpers.Path(repo, "test/data/l10n")
		configPath = helpers.Path(repo, "test/data/configuration")
	})

	BeforeEach(func() {
		vfs, root, config = helpers.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, helpers.Silent,
		)
	})

	DescribeTable("ShrinkCmd",
		func(entry *shrinkTE) {
			entry.directory = BackyardWorldsPlanet9Scan01
			entry.configPath = configPath
			expectValidShrinkCmdInvocation(vfs, entry, root, config)
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

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "vanilla with long form skim",
				args: []string{
					"--strip",
					"--interlace", "plane",
					"--quality", "85",
					"--skim",
				},
			},
		}),

		Entry(nil, &shrinkTE{
			commandTE: commandTE{
				message: "vanilla with short form skim",
				args: []string{
					"--strip",
					"--interlace", "plane",
					"--quality", "85",
					"-K",
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

			expectValidShrinkCmdInvocation(vfs, entry, root, config)
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

			expectValidShrinkCmdInvocation(vfs, entry, root, config)
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

			expectValidShrinkCmdInvocation(vfs, entry, root, config)
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

			expectValidShrinkCmdInvocation(vfs, entry, root, config)
		})
	})
})
