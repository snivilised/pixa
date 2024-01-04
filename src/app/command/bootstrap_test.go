package command_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	cmocks "github.com/snivilised/cobrass/src/assistant/mocks"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/mocks"
	"github.com/snivilised/pixa/src/internal/helpers"
	"go.uber.org/mock/gomock"
	"golang.org/x/text/language"
)

type DetectorStub struct {
}

type ExecutorStub struct {
	Name string
}

func (e *ExecutorStub) ProgName() string {
	return e.Name
}

func (e *ExecutorStub) Look() (string, error) {
	return "", nil
}

func (e *ExecutorStub) Execute(_ ...string) error {
	return nil
}

func (j *DetectorStub) Scan() language.Tag {
	return language.BritishEnglish
}

var _ = Describe("Bootstrap", Ordered, func() {

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

	Context("given: root defined with magick sub-command", func() {
		It("🧪 should: setup command without error", func() {
			bootstrap := command.Bootstrap{
				Vfs: vfs,
			}
			rootCmd := bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
				co.Detector = &DetectorStub{}
				co.Program = &ExecutorStub{
					Name: "magick",
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
			})

			Expect(rootCmd).NotTo(BeNil())
		})
	})
})
