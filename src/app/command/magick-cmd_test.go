package command_test

import (
	. "github.com/onsi/ginkgo/v2" //nolint:revive // foo
	. "github.com/onsi/gomega"    //nolint:revive // foo

	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/internal/helpers"
	lab "github.com/snivilised/pixa/src/internal/laboratory"
	"github.com/snivilised/traverse/lfs"
)

var _ = Describe("MagickCmd", Ordered, func() {
	var (
		repo       string
		l10nPath   string
		configPath string
		FS         lfs.TraverseFS
	)

	BeforeAll(func() {
		repo = helpers.Repo("")
		l10nPath = lab.Path(repo, "test/data/l10n")
		configPath = lab.Path(repo, "test/data/configuration")
	})

	BeforeEach(func() {
		Expect(lab.UseI18n(l10nPath)).To(Succeed())

		FS, _ = lab.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, helpers.Silent,
		)
	})

	When("specified flags are valid", func() {
		It("🧪 should: execute without error", func() {
			bootstrap := command.Bootstrap{
				FS: FS,
			}
			tester := lab.CommandTester{
				Args: []string{"mag", "--no-tui"},
				Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
					co.Detector = &lab.DetectorStub{}
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
