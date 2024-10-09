package command_test

import (
	. "github.com/onsi/ginkgo/v2" //nolint:revive // foo
	. "github.com/onsi/gomega"    //nolint:revive // foo

	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/li18ngo"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/internal/helpers"
)

var _ = Describe("MagickCmd", Ordered, func() {
	var (
		repo       string
		l10nPath   string
		configPath string
		vfs        storage.VirtualFS
	)

	BeforeAll(func() {
		vfs = storage.UseNativeFS()
		repo = helpers.Repo("")
		l10nPath = helpers.Path(repo, "test/data/l10n")
		configPath = helpers.Path(repo, "test/data/configuration")
	})

	BeforeEach(func() {
		Expect(li18ngo.Use()).To(Succeed())
		vfs, _ = helpers.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, helpers.Silent,
		)
	})

	When("specified flags are valid", func() {
		It("🧪 should: execute without error", func() {
			bootstrap := command.Bootstrap{
				Vfs: vfs,
			}
			tester := helpers.CommandTester{
				Args: []string{"mag", "--no-tui"},
				Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
					co.Detector = &DetectorStub{}
					co.Config.Name = common.Definitions.Pixa.ConfigTestFilename
					co.Config.ConfigPath = configPath
					co.Config.Viper = &configuration.GlobalViperConfig{}
				}),
			}
			_, err := tester.Execute()
			Expect(err).Error().To(BeNil(),
				"should pass validation due to all flag being valid",
			)
		})
	})
})
