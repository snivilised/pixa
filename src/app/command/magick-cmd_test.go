package command_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	cmocks "github.com/snivilised/cobrass/src/assistant/mocks"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/mocks"
	"github.com/snivilised/pixa/src/internal/helpers"
)

var _ = Describe("MagickCmd", Ordered, func() {
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
		mockLoggingReader  *mocks.MockLoggingConfigReader
		mockViperConfig    *cmocks.MockViperConfig
	)

	BeforeAll(func() {
		vfs = storage.UseNativeFS()
		repo = helpers.Repo("")
		l10nPath = helpers.Path(repo, "test/data/l10n")
		configPath = helpers.Path(repo, "test/data/configuration")
	})

	BeforeEach(func() {
		xi18n.ResetTx()
		vfs, _, config = helpers.SetupTest(
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
		helpers.DoMockConfigs(config,
			mockProfilesReader,
			mockSchemesReader,
			mockSamplerReader,
			mockAdvancedReader,
			mockLoggingReader,
		)
	})

	When("specified flags are valid", func() {
		It("🧪 should: execute without error", func() {
			bootstrap := command.Bootstrap{
				Vfs: vfs,
			}
			tester := helpers.CommandTester{
				Args: []string{"mag"},
				Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
					co.Detector = &DetectorStub{}
					co.Program = &helpers.ExecutorStub{
						Name: helpers.ProgName,
					}
					co.Config.Name = helpers.PixaConfigTestFilename
					co.Config.ConfigPath = configPath
					co.Config.Viper = &configuration.GlobalViperConfig{}
					co.Config.Readers = command.ConfigReaders{
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
				"should pass validation due to all flag being valid",
			)
		})
	})
})
