package proxy_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/proxy"
	"github.com/snivilised/pixa/src/internal/helpers"
	"github.com/snivilised/pixa/src/internal/matchers"
)

var _ = Describe("PathFinder", Ordered, func() {
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
		if err := helpers.UseI18n(l10nPath); err != nil {
			Fail(err.Error())
		}
	})
	// the PathFinder should not be aware of profile/sample, it only
	// know about paths. So it knows about:
	// - output
	// - item-path (full-path to current file item)
	// - input: parent dir of item
	// - current
	// - output: is the parent dir of the output, it may be the same as input
	// - subpath (this incorporates profile & sample)
	// - trash

	// - segments: "profile-name", "sample-scheme"
	//
	// * output (if output is "" => LOCALISED-MODE else CENTRALISED-MODE):
	// - localised (trash-dir=pwd+"__trash")
	// - centralised (trash-dir=pwd)

	Context("foo", func() {
		It("should:", func() {
			_ = proxy.PathFinder{}
			Expect(1).To(Equal(1))
		})
	})
})
