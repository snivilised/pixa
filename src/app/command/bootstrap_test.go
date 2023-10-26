package command_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/extendio/xfs/utils"
	"github.com/snivilised/pixa/src/app/command"
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
		repo     string
		l10nPath string
	)

	BeforeAll(func() {
		repo = helpers.Repo("../..")
		l10nPath = helpers.Path(repo, "test/data/l10n")
		Expect(utils.FolderExists(l10nPath)).To(BeTrue())
	})

	Context("given: root defined with magick sub-command", func() {
		It("🧪 should: setup command without error", func() {
			bootstrap := command.Bootstrap{}
			rootCmd := bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
				co.Detector = &DetectorStub{}
				co.Executor = &ExecutorStub{
					Name: "magick",
				}
				co.Config.Name = "pixa-test"
				co.Config.ConfigPath = "../../test/data/configuration"
			})
			Expect(rootCmd).NotTo(BeNil())
		})
	})
})
