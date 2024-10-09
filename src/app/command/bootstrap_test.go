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

var _ = Describe("Bootstrap", Ordered, func() {
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
		FS, _ = lab.SetupTest("nasa-scientist-index.xml", configPath, l10nPath, lab.Silent)
	})

	Context("given: root defined with magick sub-command", func() {
		It("🧪 should: setup command without error", func() {
			bootstrap := command.Bootstrap{
				FS: FS,
			}
			rootCmd := bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
				co.Detector = &lab.DetectorStub{}
				co.Config.Name = common.Definitions.Pixa.ConfigTestFilename
				co.Config.ConfigPath = configPath
				co.Config.Viper = &configuration.GlobalViperConfig{}
			})

			Expect(rootCmd).NotTo(BeNil())
		})
	})
})
