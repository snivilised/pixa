package command_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/internal/helpers"
	"github.com/snivilised/pixa/src/internal/matchers"
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
		nfs        storage.VirtualFS
	)

	BeforeAll(func() {
		nfs = storage.UseNativeFS()
		repo = helpers.Repo(filepath.Join("..", "..", ".."))

		l10nPath = helpers.Path(repo, filepath.Join("test", "data", "l10n"))
		Expect(matchers.AsDirectory(l10nPath)).To(matchers.ExistInFS(nfs))

		configPath = filepath.Join(repo, "test", "data", "configuration")
		Expect(matchers.AsDirectory(configPath)).To(matchers.ExistInFS(nfs))
	})

	Context("given: root defined with magick sub-command", func() {
		It("🧪 should: setup command without error", func() {
			bootstrap := command.Bootstrap{
				Vfs: nfs,
			}
			rootCmd := bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
				co.Detector = &DetectorStub{}
				co.Program = &ExecutorStub{
					Name: "magick",
				}
				co.Config.Name = helpers.PixaConfigTestFilename
				co.Config.ConfigPath = configPath
			})
			Expect(rootCmd).NotTo(BeNil())
		})
	})
})
