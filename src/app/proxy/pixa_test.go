package proxy_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	cmocks "github.com/snivilised/cobrass/src/assistant/mocks"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/extendio/xfs/utils"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/app/proxy/filing"
	"github.com/snivilised/pixa/src/cfg"

	"github.com/snivilised/pixa/src/app/mocks"
	"github.com/snivilised/pixa/src/internal/helpers"
	"github.com/snivilised/pixa/src/internal/matchers"
	"go.uber.org/mock/gomock"
)

const (
	BackyardWorldsPlanet9Scan01 = "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01"
	BackyardWorldsPlanet9Scan02 = "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-02"
	perm                        = 0o766
)

type controllerTE struct {
	given        string
	should       string
	exists       bool
	args         []string
	dry          bool
	withFake     bool
	outputFlag   string
	trashFlag    string
	profile      string
	relative     string
	mandatory    []string
	intermediate string
	supplement   string
	inputs       []string
}

type samplerTE struct {
	controllerTE
	scheme string
}

var _ = Describe("pixa", Ordered, func() {
	var (
		repo               string
		l10nPath           string
		configPath         string
		root               string
		config             configuration.ViperConfig
		vfs                storage.VirtualFS
		ctrl               *gomock.Controller
		mockProfilesReader *mocks.MockProfilesConfigReader
		mockSchemesReader  *mocks.MockSchemesConfigReader
		mockSamplerReader  *mocks.MockSamplerConfigReader
		mockAdvancedReader *mocks.MockAdvancedConfigReader
		mockLoggingReader  *mocks.MockLoggingConfigReader
		mockViperConfig    *cmocks.MockViperConfig
	)

	BeforeAll(func() {
		repo = helpers.Repo("")
		l10nPath = helpers.Path(repo, "test/data/l10n")
		configPath = helpers.Path(repo, "test/data/configuration")
	})

	BeforeEach(func() {
		vfs, root, config = helpers.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, helpers.Silent,
		)

		ctrl = gomock.NewController(GinkgoT())
		mockViperConfig = cmocks.NewMockViperConfig(ctrl)
		mockProfilesReader = mocks.NewMockProfilesConfigReader(ctrl)
		mockSchemesReader = mocks.NewMockSchemesConfigReader(ctrl)
		mockSamplerReader = mocks.NewMockSamplerConfigReader(ctrl)
		mockAdvancedReader = mocks.NewMockAdvancedConfigReader(ctrl)
		mockLoggingReader = mocks.NewMockLoggingConfigReader(ctrl)
		helpers.DoMockReadInConfig(mockViperConfig)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	DescribeTable("run",
		func(entry *samplerTE) {
			helpers.DoMockConfigs(config,
				mockProfilesReader,
				mockSchemesReader,
				mockSamplerReader,
				mockAdvancedReader,
				mockLoggingReader,
			)
			directory := helpers.Path(root, entry.relative)
			options := []string{
				helpers.ShrinkCommandName, directory,
			}

			args := options
			args = append(args, entry.args...)

			if entry.exists {
				location := filepath.Join(directory, entry.intermediate, entry.supplement)
				if err := vfs.MkdirAll(location, perm); err != nil {
					Fail(errors.Wrap(err, err.Error()).Error())
				}
			}

			if entry.outputFlag != "" {
				output := helpers.Path(root, entry.outputFlag)
				args = append(args, "--output", output)
			}
			if entry.trashFlag != "" {
				trash := helpers.Path(root, entry.trashFlag)
				args = append(args, "--trash", trash)
			}
			if entry.dry {
				args = append(args, "--dry-run")
			}

			bootstrap := command.Bootstrap{
				Vfs: vfs,
			}
			tester := helpers.CommandTester{
				Args: args,
				Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
					co.Detector = &helpers.DetectorStub{}
					co.Config.Name = helpers.PixaConfigTestFilename
					co.Config.ConfigPath = configPath
					co.Config.Viper = &configuration.GlobalViperConfig{}
					co.Config.Readers = cfg.ConfigReaders{
						Profiles: mockProfilesReader,
						Schemes:  mockSchemesReader,
						Sampler:  mockSamplerReader,
						Advanced: mockAdvancedReader,
						Logging:  mockLoggingReader,
					}
				}),
			}

			_, err := tester.Execute()
			Expect(err).Error().To(BeNil(),
				fmt.Sprintf("execution result non nil (%v)", err),
			)

			if entry.mandatory != nil && entry.dry {
				dejaVuSupplement := filepath.Join(common.DejaVu, entry.supplement)
				supplement := helpers.Path(entry.intermediate, dejaVuSupplement)

				for _, original := range entry.mandatory {
					originalPath := filepath.Join(supplement, original)

					if entry.withFake {
						withFake := filing.ComposeFake(original, bootstrap.AdvancedCFG.FakeLabel())
						originalPath = filepath.Join(directory, withFake)
					}

					Expect(matchers.AsFile(originalPath)).To(matchers.ExistInFS(vfs))
				}
			}
		},
		func(entry *samplerTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should: '%v'",
				entry.given, entry.should,
			)
		},

		// full run

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full run transparent adhoc, with ex-glob",
				should:   "full run with ex-glob filter, result file takes place of input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "ADHOC/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full run transparent adhoc, with regex",
				should:   "full run with regex filter, result file takes place of input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files-rx", "Backyard-Worlds",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "ADHOC/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full run transparent with profile, with ex-glob",
				should:   "full run with ex-glob filter, result file takes place of input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--profile", "adaptive",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "adaptive/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full run, profile",
				should:   "full run with regex filter using the defined profile",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--strip",
					"--interlace", "plane",
					"--quality", "85",
					"--files-rx", "Backyard-Worlds",
					"--profile", "adaptive",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "adaptive/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full run transparent with scheme with single profile",
				should:   "full run with ex-glob filter, result file takes place of input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--scheme", "singleton",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "singleton/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full run non transparent adhoc",
				should:   "full run with ex-glob filter, input moved to alternative location",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				trashFlag:    "discard",
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "discard",
				supplement:   "ADHOC/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full run non transparent with profile",
				should:   "full run with ex-glob filter, input moved to alternative location",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--profile", "adaptive",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				trashFlag:    "discard",
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "discard",
				supplement:   "adaptive/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "run non transparent scheme single with profile",
				should:   "full run with ex-glob filter, input moved to alternative location",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--scheme", "singleton",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				trashFlag:    "discard",
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "discard",
				supplement:   "singleton/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full run, scheme",
				should:   "full run, all profiles in the scheme",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--strip",
					"--interlace", "plane",
					"--quality", "85",
					"--scheme", "blur-sf",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "blur-sf/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full run transparent adhoc and target already exists",
				should:   "full run with ex-glob filter, result file takes place of input",
				exists:   true,
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "ADHOC/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "directory contains files with same name different extensions",
				should:   "create journal file include file extension",
				relative: BackyardWorldsPlanet9Scan02,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan02,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-02",
				supplement:   "ADHOC/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan02,
			},
		}),

		// sample run

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "run transparent adhoc",
				should:   "sample(first) with ex-glob filter, result file takes place of input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--no-files", "4",
					"--files", "*Backyard-Worlds*",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First4,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "ADHOC/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First4,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "run transparent with profile",
				should:   "sample(first) with ex-glob filter, result file takes place of input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--no-files", "4",
					"--files", "*Backyard-Worlds*",
					"--profile", "adaptive",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First4,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "adaptive/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First4,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "run(last) transparent with profile",
				should:   "sample(last) with ex-glob filter using the defined profile",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--last",
					"--no-files", "4",
					"--files", "*Backyard-Worlds*",
					"--profile", "adaptive",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01Last4,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "adaptive/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01Last4,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "profile without no-files in args",
				should:   "sample(first) with ex-glob filter, using no-files from config",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--files", "*Backyard-Worlds*",
					"--profile", "adaptive",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First2,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "adaptive/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First2,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "profile",
				should:   "sample with regex filter using the defined profile",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--no-files", "4",
					"--strip",
					"--interlace", "plane",
					"--quality", "85",
					"--files-rx", "Backyard-Worlds",
					"--profile", "adaptive",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First4,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "adaptive/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First4,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "run transparent with scheme with single profile",
				should:   "sample(first) with ex-glob filter, result file takes place of input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--no-files", "4",
					"--files", "*Backyard-Worlds*",
					"--scheme", "singleton",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First4,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "singleton/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First4,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "run non transparent adhoc",
				should:   "sample(first) with ex-glob filter, input moved to alternative location",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--no-files", "4",
					"--files", "*Backyard-Worlds*",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				trashFlag:    "discard",
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First4,
				intermediate: "discard",
				supplement:   "ADHOC/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First4,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "run non transparent with profile",
				should:   "sample(first) with ex-glob filter, input moved to alternative location",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--no-files", "4",
					"--files", "*Backyard-Worlds*",
					"--profile", "adaptive",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				trashFlag:    "discard",
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First4,
				intermediate: "discard",
				supplement:   "adaptive/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First4,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "run non transparent scheme single with profile",
				should:   "sample(first) with ex-glob filter, input moved to alternative location",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--no-files", "4",
					"--files", "*Backyard-Worlds*",
					"--scheme", "singleton",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				trashFlag:    "discard",
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First4,
				intermediate: "discard",
				supplement:   "singleton/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First4,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "scheme",
				should:   "sample all profiles in the scheme",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--no-files", "4",
					"--files", "*Backyard-Worlds*",
					"--strip",
					"--interlace", "plane",
					"--quality", "85",
					"--scheme", "blur-sf",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "blur-sf/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First4,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "run transparent adhoc and target already exists",
				should:   "sample(first) with ex-glob filter, result file takes place of input",
				exists:   true,
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--no-files", "4",
					"--files", "*Backyard-Worlds*",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First4,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "ADHOC/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First4,
			},
		}),

		// dry run

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full dry run transparent adhoc, using glob",
				should:   "full run with ex-glob filter, without moving input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				dry:          true,
				withFake:     true,
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "ADHOC/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full dry run transparent adhoc, using regex",
				should:   "full run with regex filter, without moving input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files-rx", "Backyard-Worlds",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				dry:          true,
				withFake:     true,
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "ADHOC/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),

		Entry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "full dry run transparent with profile, using glob",
				should:   "full run with ex-glob filter, without moving input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--profile", "adaptive",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				dry:          true,
				withFake:     true,
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "adaptive/TRASH",
				inputs:       helpers.BackyardWorldsPlanet9Scan01First6,
			},
		}),
	)
})

var _ = Describe("end to end", Ordered, func() {
	Context("REAL", func() {
		XIt("should: tinkle the ivories", func() {
			// pixa shrink ~/dev/test/pics --profile blur --sample --no-files 1 --files "screen*" --dry-run
			args := []string{
				"shrink",
				"/Users/plastikfan/dev/test/pics",
				"--profile", "blur",
				// "--sample",
				// "--no-files", "1",
				"--files", "wonky*",
				"--dry-run",
			}
			configPath := utils.ResolvePath("~/snivilised/pixa")
			bootstrap := command.Bootstrap{
				Vfs: storage.UseNativeFS(),
			}
			tester := helpers.CommandTester{
				Args: args,
				Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
					co.Detector = &helpers.DetectorStub{}
					co.Config.Name = "pixa"
					co.Config.ConfigPath = configPath
				}),
			}

			_, err := tester.Execute()
			Expect(err).Error().To(BeNil(),
				fmt.Sprintf("execution result non nil (%v)", err),
			)
		})
	})
})
