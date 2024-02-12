package proxy_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/extendio/xfs/utils"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/app/proxy/filing"

	"github.com/snivilised/pixa/src/internal/helpers"
	"github.com/snivilised/pixa/src/internal/matchers"
)

const (
	BackyardWorldsPlanet9Scan01 = "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01"
	BackyardWorldsPlanet9Scan02 = "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-02"
)

type controllerTE struct {
	given        string
	should       string
	exists       bool
	args         []string
	isTui        bool
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

func augment(entry *samplerTE,
	args []string, vfs storage.VirtualFS, root, directory string,
) []string {
	result := args
	result = append(result, entry.args...)

	if entry.exists {
		location := filepath.Join(directory, entry.intermediate, entry.supplement)
		if err := vfs.MkdirAll(location, common.Permissions.Write); err != nil {
			Fail(errors.Wrap(err, err.Error()).Error())
		}
	}

	if entry.outputFlag != "" {
		output := helpers.Path(root, entry.outputFlag)
		result = append(result, "--output", output)
	}

	if entry.trashFlag != "" {
		trash := helpers.Path(root, entry.trashFlag)
		result = append(result, "--trash", trash)
	}

	if entry.dry {
		result = append(result, "--dry-run")
	}

	return result
}

func assertInFs(entry *samplerTE, bs *command.Bootstrap, directory string) {
	vfs := bs.Vfs

	if entry.mandatory != nil && entry.dry {
		dejaVuSupplement := filepath.Join(common.Definitions.Filing.DejaVu, entry.supplement)
		supplement := helpers.Path(entry.intermediate, dejaVuSupplement)

		for _, original := range entry.mandatory {
			originalPath := filepath.Join(supplement, original)

			if entry.withFake {
				withFake := filing.ComposeFake(original, bs.Configs.Advanced.FakeLabel())
				originalPath = filepath.Join(directory, withFake)
			}

			Expect(matchers.AsFile(originalPath)).To(matchers.ExistInFS(vfs))
		}
	}
}

var _ = Describe("pixa", Ordered, func() {
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

	DescribeTable("interactive",
		func(entry *samplerTE) {
			directory := helpers.Path(root, entry.relative)
			args := augment(entry,
				[]string{
					common.Definitions.Commands.Shrink, directory,
				},
				vfs, root, directory,
			)

			bootstrap := command.Bootstrap{
				Vfs: vfs,
				Presentation: common.PresentationOptions{
					WithoutRenderer: true,
				},
			}

			if !entry.isTui {
				args = append(args, "--no-tui")
			}

			tester := helpers.CommandTester{
				Args: args,
				Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
					co.Detector = &helpers.DetectorStub{}
					co.Config.Name = common.Definitions.Pixa.ConfigTestFilename
					co.Config.ConfigPath = configPath
					co.Config.Viper = &configuration.GlobalViperConfig{}
				}),
			}

			_, err := tester.Execute()
			Expect(err).Error().To(BeNil(),
				fmt.Sprintf("execution result non nil (%v)", err),
			)

			assertInFs(entry, &bootstrap, directory)
		},
		func(entry *samplerTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should: '%v'",
				entry.given, entry.should,
			)
		},

		// linear-ui
		//
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

		// textual-ui
		// https://dev.to/pomdtr/how-to-debug-bubble-tea-applications-in-visual-studio-code-50jp
		//
		// full run
		//
		// 📚 debugging bubbletea
		// !! use the predefined "Attach to dlv" launch task. It requires the task
		// defined as "Run headless dlv". This starts the dlv debugger in headless mode.
		// Note that because the args are hardcoded into the dlv task, if you need to
		// debug pixa with different args, then the dlv task needs to be modified.
		//
		// to attach to dlv debugger manually, start dlv like this:
		// dlv debug --headless --listen=:2345 ./src/app/main/ -- shrink /Users/plastikfan/dev/test --profile blur --files "wonky*" --dry-run
		// the args come after --
		// the launch.json does not support args for an attach request, args are only
		// appropriate for launch
		//
		// Beware, if you start dlv manually, you will need to define a new launch entry
		// that does not depend on the "Run headless dlv" as that attempts to start dlv
		// automatically.
		//
		XEntry(nil, &samplerTE{
			controllerTE: controllerTE{
				given:    "bubbletea tui, full run transparent adhoc, with ex-glob",
				should:   "full run with ex-glob filter, result file takes place of input",
				relative: BackyardWorldsPlanet9Scan01,
				args: []string{
					"--files", "*Backyard-Worlds*",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				isTui:        true,
				mandatory:    helpers.BackyardWorldsPlanet9Scan01First6,
				intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
				supplement:   "ADHOC/TRASH",
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
					co.Config.Name = common.Definitions.Pixa.AppName
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
