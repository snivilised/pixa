package command_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/internal/helpers"
	"github.com/snivilised/pixa/src/internal/matchers"
)

var _ = Describe("MagickCmd", Ordered, func() {
	var (
		repo     string
		l10nPath string
		nfs      storage.VirtualFS
	)

	BeforeAll(func() {
		nfs = storage.UseNativeFS()
		repo = helpers.Repo(filepath.Join("..", "..", ".."))
		l10nPath = helpers.Path(repo, filepath.Join("test", "data", "l10n"))
		Expect(matchers.AsDirectory(l10nPath)).To(matchers.ExistInFS(nfs))
	})

	BeforeEach(func() {
		xi18n.ResetTx()

		if err := helpers.UseI18n(l10nPath); err != nil {
			Fail(err.Error())
		}
	})

	When("specified flags are valid", func() {
		It("🧪 should: execute without error", func() {
			bootstrap := command.Bootstrap{
				Vfs: nfs,
			}
			tester := helpers.CommandTester{
				Args: []string{"mag"},
				Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
					co.Detector = &DetectorStub{}
				}),
			}
			_, err := tester.Execute()
			Expect(err).Error().To(BeNil(),
				"should pass validation due to all flag being valid",
			)
		})
	})
})
