package proxy_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	cmocks "github.com/snivilised/cobrass/src/assistant/mocks"
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/proxy"

	"github.com/snivilised/pixa/src/app/mocks"
	"github.com/snivilised/pixa/src/internal/helpers"
	"github.com/snivilised/pixa/src/internal/matchers"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"
)

const (
	silent                      = true
	verbose                     = false
	faydeaudeau                 = os.FileMode(0o777)
	beezledub                   = os.FileMode(0o666)
	backyardWorldsPlanet9Scan01 = "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01"
)

var (
	_ proxy.ProfilesConfig       = proxy.MsProfilesConfig{}
	_ proxy.SamplerConfig        = &proxy.MsSamplerConfig{}
	_ proxy.ProfilesConfigReader = &proxy.MsProfilesConfigReader{}
	_ proxy.SamplerConfigReader  = &proxy.MsSamplerConfigReader{}

	backyardWorldsPlanet9Scan01First2 []string
	backyardWorldsPlanet9Scan01First4 []string
	backyardWorldsPlanet9Scan01First6 []string

	profilesConfigData proxy.ProfilesConfigMap
	samplerConfigData  *proxy.MsSamplerConfig
)

func init() {
	backyardWorldsPlanet9Scan01First2 = []string{
		"01_Backyard-Worlds-Planet-9_s01.jpg",
		"02_Backyard-Worlds-Planet-9_s01.jpg",
	}

	backyardWorldsPlanet9Scan01First4 = backyardWorldsPlanet9Scan01First2
	backyardWorldsPlanet9Scan01First4 = append(
		backyardWorldsPlanet9Scan01First4,
		[]string{
			"03_Backyard-Worlds-Planet-9_s01.jpg",
			"04_Backyard-Worlds-Planet-9_s01.jpg",
		}...,
	)

	backyardWorldsPlanet9Scan01First6 = backyardWorldsPlanet9Scan01First4
	backyardWorldsPlanet9Scan01First6 = append(
		backyardWorldsPlanet9Scan01First6,
		[]string{
			"05_Backyard-Worlds-Planet-9_s01.jpg",
			"06_Backyard-Worlds-Planet-9_s01.jpg",
		}...,
	)

	profilesConfigData = proxy.ProfilesConfigMap{
		"blur": clif.ChangedFlagsMap{
			"strip":         "true",
			"interlace":     "plane",
			"gaussian-blur": "0.05",
		},
		"sf": clif.ChangedFlagsMap{
			"dry-run":         "true",
			"strip":           "true",
			"interlace":       "plane",
			"sampling-factor": "4:2:0",
		},
		"adaptive": clif.ChangedFlagsMap{
			"strip":           "true",
			"interlace":       "plane",
			"gaussian-blur":   "0.25",
			"adaptive-resize": "60",
		},
	}

	samplerConfigData = &proxy.MsSamplerConfig{
		Files:   2,
		Folders: 1,
		Schemes: proxy.MsSamplerSchemesConfig{
			"blur-sf": proxy.MsSchemeConfig{
				Profiles: []string{"blur", "sf"},
			},
			"adaptive-sf": proxy.MsSchemeConfig{
				Profiles: []string{"adaptive", "sf"},
			},
			"adaptive-blur": proxy.MsSchemeConfig{
				Profiles: []string{"adaptive", "blur"},
			},
		},
	}
}

func doMockProfilesConfigsWith(
	data proxy.ProfilesConfigMap,
	config configuration.ViperConfig,
	reader *mocks.MockProfilesConfigReader,
) {
	reader.EXPECT().Read(config).DoAndReturn(
		func(viper configuration.ViperConfig) (proxy.ProfilesConfig, error) {
			stub := &proxy.MsProfilesConfig{
				Profiles: data,
			}

			return stub, nil
		},
	).AnyTimes()
}

func doMockSamplerConfigWith(
	data *proxy.MsSamplerConfig,
	config configuration.ViperConfig,
	reader *mocks.MockSamplerConfigReader,
) {
	reader.EXPECT().Read(config).DoAndReturn(
		func(viper configuration.ViperConfig) (proxy.SamplerConfig, error) {
			stub := data

			return stub, nil
		},
	).AnyTimes()
}

func doMockConfigs(
	config configuration.ViperConfig,
	profilesReader *mocks.MockProfilesConfigReader,
	samplerReader *mocks.MockSamplerConfigReader,
) {
	doMockProfilesConfigsWith(profilesConfigData, config, profilesReader)
	doMockSamplerConfigWith(samplerConfigData, config, samplerReader)
}

func doMockViper(config *cmocks.MockViperConfig) {
	config.EXPECT().ReadInConfig().DoAndReturn(
		func() error {
			return nil
		},
	).AnyTimes()
}

func resetFS(index string, silent bool) (vfs storage.VirtualFS, root string) {
	vfs = storage.UseMemFS()
	root = helpers.Scientist(vfs, index, silent)
	// ??? Expect(matchers.AsDirectory(root)).To(matchers.ExistInFS(vfs))

	return vfs, root
}

type runnerTE struct {
	given    string
	should   string
	args     []string
	profile  string
	relative string
	expected []string
}

type samplerTE struct {
	runnerTE
	scheme string
}

var _ = Describe("SamplerRunner", Ordered, func() {
	var (
		repo               string
		l10nPath           string
		configPath         string
		root               string
		config             configuration.ViperConfig
		vfs                storage.VirtualFS
		ctrl               *gomock.Controller
		mockProfilesReader *mocks.MockProfilesConfigReader
		mockSamplerReader  *mocks.MockSamplerConfigReader
		mockViperConfig    *cmocks.MockViperConfig
	)

	BeforeAll(func() {
		repo = helpers.Repo(filepath.Join("..", "..", ".."))
		l10nPath = helpers.Path(repo, filepath.Join("test", "data", "l10n"))
		configPath = filepath.Join(repo, "test", "data", "configuration")
	})

	BeforeEach(func() {
		viper.Reset()
		vfs, root = resetFS("nasa-scientist-index.xml", silent)

		ctrl = gomock.NewController(GinkgoT())
		mockViperConfig = cmocks.NewMockViperConfig(ctrl)
		mockProfilesReader = mocks.NewMockProfilesConfigReader(ctrl)
		mockSamplerReader = mocks.NewMockSamplerConfigReader(ctrl)
		doMockViper(mockViperConfig)

		// create a dummy config file in vfs
		//
		_ = vfs.MkdirAll(configPath, beezledub)
		if _, err := vfs.Create(filepath.Join(configPath, helpers.PixaConfigTestFilename)); err != nil {
			Fail(fmt.Sprintf("🔥 can't create dummy config (err: '%v')", err))
		}

		Expect(matchers.AsDirectory(configPath)).To(matchers.ExistInFS(vfs))

		config = &configuration.GlobalViperConfig{}

		config.SetConfigType(helpers.PixaConfigType)
		config.SetConfigName(helpers.PixaConfigTestFilename)
		config.AddConfigPath(configPath)

		if err := config.ReadInConfig(); err != nil {
			Fail(fmt.Sprintf("🔥 can't read config (err: '%v')", err))
		}

		if err := helpers.UseI18n(l10nPath); err != nil {
			Fail(err.Error())
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	DescribeTable("sampler",
		func(entry *samplerTE) {
			doMockConfigs(config, mockProfilesReader, mockSamplerReader)

			directory := helpers.Path(root, entry.relative)
			options := []string{
				helpers.ShrinkCommandName, directory,
				"--dry-run",
				"--mode", "tidy",
			}
			args := options
			args = append(args, entry.args...)

			bootstrap := command.Bootstrap{
				Vfs: vfs,
			}
			tester := helpers.CommandTester{
				Args: args,
				Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
					co.Detector = &helpers.DetectorStub{}
					co.Program = &helpers.ExecutorStub{
						Name: helpers.ProgName,
					}
					co.Config.Name = helpers.PixaConfigTestFilename
					co.Config.ConfigPath = configPath
					co.Viper = &configuration.GlobalViperConfig{}
					co.Config.Readers = proxy.ConfigReaders{
						Profiles: mockProfilesReader,
						Sampler:  mockSamplerReader,
					}
				}),
			}

			_, err := tester.Execute()
			Expect(err).Error().To(BeNil(),
				fmt.Sprintf("execution result non nil (%v)", err),
			)

			// eventually, we should assert on files created in the virtual
			// file system, using entry.expected
			//
		},
		func(entry *samplerTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should: '%v'",
				entry.given, entry.should,
			)
		},

		Entry(nil, &samplerTE{
			runnerTE: runnerTE{
				given:    "profile",
				should:   "sample(first) with glob filter using the defined profile",
				relative: backyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--no-files", "4",
					"--files-gb", "*Backyard Worlds*",
					"--profile", "adaptive",
					"--gaussian-blur", "0.51",
					"--interlace", "line",
				},
				expected: backyardWorldsPlanet9Scan01First4,
			},
		}),

		Entry(nil, &samplerTE{
			runnerTE: runnerTE{
				given:    "profile",
				should:   "sample(last) with glob filter using the defined profile",
				relative: backyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--last",
					"--no-files", "4",
					"--files-gb", "*Energy-Explorers*",
					"--profile", "adaptive",
				},
				expected: backyardWorldsPlanet9Scan01First4,
			},
		}),

		Entry(nil, &samplerTE{
			runnerTE: runnerTE{
				given:    "profile without no-files in args",
				should:   "sample(first) with glob filter, using no-files from config",
				relative: backyardWorldsPlanet9Scan01,
				args: []string{
					"--sample",
					"--files-gb", "*Energy-Explorers*",
					"--profile", "adaptive",
				},
				expected: backyardWorldsPlanet9Scan01First2,
			},
		}),

		XEntry(nil, &samplerTE{
			runnerTE: runnerTE{
				given:    "profile",
				should:   "sample with regex filter using the defined profile",
				relative: backyardWorldsPlanet9Scan01,
				args: []string{
					"--strip", "--interlace", "plane", "--quality", "85", "--profile", "adaptive",
				},
			},
		}),

		// override config with explicitly defined args on command line
		// ie they should be present in args as opposed to relying on presence
		// in config. Here we are testing that command line overrides config
		// ...

		// ===

		Entry(nil, &samplerTE{
			runnerTE: runnerTE{
				given:    "scheme",
				should:   "sample all profiles in the scheme",
				relative: backyardWorldsPlanet9Scan01,
				args: []string{
					"--strip", "--interlace", "plane", "--quality", "85", "--scheme", "blur-sf",
				},
				expected: backyardWorldsPlanet9Scan01First6,
			},
		}),
	)
})
