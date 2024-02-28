package command_test

import (
	. "github.com/onsi/ginkgo/v2" //nolint:revive // foo
	. "github.com/onsi/gomega"    //nolint:revive // foo
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/internal/helpers"
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
		repo       string
		l10nPath   string
		configPath string
		vfs        storage.VirtualFS
	)

	BeforeAll(func() {
		repo = helpers.Repo("")
		l10nPath = helpers.Path(repo, "test/data/l10n")
		configPath = helpers.Path(repo, "test/data/configuration")
	})

	BeforeEach(func() {
		vfs, _ = helpers.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, helpers.Silent,
		)
	})

	Context("given: root defined with magick sub-command", func() {
		It("🧪 should: setup command without error", func() {
			bootstrap := command.Bootstrap{
				Vfs: vfs,
			}
			rootCmd := bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
				co.Detector = &DetectorStub{}
				co.Config.Name = common.Definitions.Pixa.ConfigTestFilename
				co.Config.ConfigPath = configPath
				co.Config.Viper = &configuration.GlobalViperConfig{}
			})

			Expect(rootCmd).NotTo(BeNil())
		})
	})
})
