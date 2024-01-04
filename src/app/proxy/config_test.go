package proxy_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	cmocks "github.com/snivilised/cobrass/src/assistant/mocks"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/mocks"
	"github.com/snivilised/pixa/src/internal/helpers"
)

func expectValidShrinkCmdInvocation(vfs storage.VirtualFS, entry *configTE,
	config configuration.ViperConfig,
) {
	bootstrap := command.Bootstrap{
		Vfs: vfs,
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

	options := []string{
		entry.comm, entry.file,
		"--dry-run",
		"--mode", "tidy",
		"--profile", entry.profile,
	}

	repo := helpers.Repo("")
	configPath := filepath.Join(repo, "test", "data", "configuration")
	tester := helpers.CommandTester{
		Args: append(options, entry.args...),
		Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
			co.Detector = &helpers.DetectorStub{}
			co.Program = &helpers.ExecutorStub{
				Name: helpers.ProgName,
			}
			co.Config.Name = helpers.PixaConfigTestFilename
			co.Config.ConfigPath = configPath
			co.Viper = &configuration.GlobalViperConfig{}
			co.Config.Readers = command.ConfigReaders{
				Profiles: mockProfilesReader,
				Schemes:  mockSchemesReader,
				Sampler:  mockSamplerReader,
				Advanced: mockAdvancedReader,
			}
		}),
	}

	_, _ = tester.Execute()
}

type configTE struct {
	message  string
	comm     string
	file     string
	args     []string
	profile  string
	expected any
	actual   func(entry *configTE) any
	assert   func(entry *configTE, actual any)
}

var _ = Describe("Config", Ordered, func() {
	var (
		repo               string
		l10nPath           string
		configPath         string
		config             configuration.ViperConfig
		vfs                storage.VirtualFS
		ctrl               *gomock.Controller
		mockProfilesReader *mocks.MockProfilesConfigReader
		mockSchemesReader  *mocks.MockSchemesConfigReader
		mockSamplerReader  *mocks.MockSamplerConfigReader
		mockAdvancedReader *mocks.MockAdvancedConfigReader
		mockViperConfig    *cmocks.MockViperConfig
	)

	BeforeAll(func() {
		repo = helpers.Repo("")
		l10nPath = helpers.Path(repo, "test/data/l10n")
		configPath = helpers.Path(repo, "test/data/configuration")
	})

	BeforeEach(func() {
		vfs, _, config = helpers.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, helpers.Silent,
		)

		ctrl = gomock.NewController(GinkgoT())
		mockViperConfig = cmocks.NewMockViperConfig(ctrl)
		mockProfilesReader = mocks.NewMockProfilesConfigReader(ctrl)
		mockSchemesReader = mocks.NewMockSchemesConfigReader(ctrl)
		mockSamplerReader = mocks.NewMockSamplerConfigReader(ctrl)
		mockAdvancedReader = mocks.NewMockAdvancedConfigReader(ctrl)
		helpers.DoMockReadInConfig(mockViperConfig)
		helpers.DoMockConfigs(config,
			mockProfilesReader, mockSchemesReader, mockSamplerReader, mockAdvancedReader,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	DescribeTable("profile",
		func(entry *configTE) {
			if entry.assert == nil {
				actual := entry.actual(entry)
				_ = actual

				Expect(1).To(Equal(1))
				expectValidShrinkCmdInvocation(vfs, entry, config)
			} else {
				actual := entry.actual(entry)
				entry.assert(entry, actual)
			}
		},
		func(entry *configTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should access profile: '%v'",
				entry.message, entry.profile,
			)
		},

		XEntry(nil, &configTE{
			message:  "adaptive",
			comm:     "shrink",
			file:     "cover.nfr.lana-del-rey.jpg",
			args:     []string{},
			profile:  "adaptive",
			expected: 42,
			actual: func(e *configTE) any {
				return config.Get(e.profile)
			},
		}),
	)
})
