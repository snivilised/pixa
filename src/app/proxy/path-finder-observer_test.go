package proxy_test

import (
	"fmt"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/app/proxy/filing"
	"github.com/snivilised/pixa/src/internal/matchers"

	. "github.com/onsi/gomega"
)

type pathAssertion struct {
	file   string
	folder string
}

type observerAssertions map[string]*pathAssertion

type testPathFinderObserver struct {
	target common.PathFinder

	transfers observerAssertions
	results   observerAssertions
}

// ✨ BEFORE A NON TRANSPARENT RUN, ANY SAMPLE FILES FOUND IN THE SAME DIR AS THE ORIGIN
// SHOULD BE MOVED TO THE --output LOCATION. THE REASON FOR THIS IS THE USER
// MAY PERFORM A TRANSPARENT SAMPLE RUN, FOLLOWED BY A NON TRANSPARENT FULL RUN. IN THIS
// CASE, THE SAMPLE FILES WILL BE IN THE WRONG LOCATION AND SHOULD BE MOVED TO THE --output

// --output(RESULT) specifies where the result goes to
// --trash(INPUT) specified where the input goes to; also, means not-transparent
// usually, the user would specify either the --output or --input; not both,
// but not prohibited, but not in the user's interest
// sample should follow the output, so that when full run occurs, the sample is easily found,
// and the user can perform direct comparison to the input. the sample is always decorated by
// the supplement, which has an additional '.$sample' label, eg
// 01_Backyard-Worlds-Planet-9_s01.jpg -> 01_Backyard-Worlds-Planet-9_s01.blur-sf.blur.$sample.jpg
// where blur-sf.blur.$sample is the full supplement (scheme=blur-sf;profile=blur;sample=.$sample)

// combinations out of dry-run/scheme(ar>1)/transparency | input -> ? | output ->
// /null = no move
// //R = output, either --output or --trash
//
// - = NO-MOVE
// X = DOES-NOT-EXIST
// $OR = ORIGIN
// $N = ITEM-NAME
// $S = SUPPLEMENT
//
// WE NEED ANOTHER OPTION, --cuddle. When specified, it tries to keep the output files
// close to the input to enable easier comparison for the user. The user is more likely
// to user --cuddle when performing a sample. When cuddled, the output file name is
// decorated with the supplement; this will disambiguate the output from the input. When
// not cuddled
//
// TRANSPARENCY NOT COMPATIBLE WITH SCHEME(ar>1); CAST IN STONE
// WHEN TRANSPARENT, THE SUPPLEMENT MUST APPLY TO FOLDER AND FILE TO AVOID NAME CLASH
//
// We only need to decorate the filename, when cuddle is active, because the input file
// is in the same location as the output file. When cuddle is not active, then the folder
// is decorated with the directory supplement, therefore no class arises, so there is
// no need for a file supplement. The supplement function should be passed the cuddle
// flag so that it can encode this logic internally.
//
//
//	TRANSP			| SCHEME	|	DRY			| SAMPLE	| CUDDLE |TRANSFER(--trash)									| RESULT(--output)
// -------------|---------|---------|---------|--------|----------------------------------|-----------------------------------------
//  YES					| NO			| NO			| NO			| NO     | - //OR/$N.$S											| //OR/$N
//	YES					| YES			|	YES			| NO			| NO     | input > /null									  | output > /null

func (o *testPathFinderObserver) assert(entry *samplerTE,
	origin string, vfs storage.VirtualFS,
) {
	if entry.supplements.file != "do not enter" {
		return
	}

	first := lo.Keys(o.transfers)[0]
	fmt.Printf("\n 📂 FOLDER: '%v'\n", o.transfers[first].folder)

	for _, v := range o.transfers {
		if o.TransparentInput() {
			// input should just be renamed with supplement in the origin folder
			//
			statics := o.target.Statics()

			originalPath := filepath.Join(origin,
				filing.SupplementFilename(v.file, entry.supplements.file, statics),
			)
			Expect(matchers.AsFile(originalPath)).To(matchers.ExistInFS(vfs))
		}
	}

	for _, v := range o.results {
		if o.TransparentInput() { // ?? CHECK-VALID FOR OUTPUT?
			// result should take the place of input
			//
			// directory
			// intermediate
			// name
			//
			originalPath := filepath.Join(origin, v.file)
			Expect(matchers.AsFile(originalPath)).To(matchers.ExistInFS(vfs))
		}
	}
}

func (o *testPathFinderObserver) Transfer(info *common.PathInfo) (folder, file string) {
	folder, file = o.target.Transfer(info)
	o.transfers[info.Item.Extension.Name] = &pathAssertion{
		folder: folder,
		file:   file,
	}

	return folder, file
}

func (o *testPathFinderObserver) Result(info *common.PathInfo) (folder, file string) {
	folder, file = o.target.Result(info)
	o.results[info.Item.Extension.Name] = &pathAssertion{
		folder: folder,
		file:   file,
	}

	return folder, file
}

func (o *testPathFinderObserver) TransparentInput() bool {
	return o.target.TransparentInput()
}

func (o *testPathFinderObserver) JournalFullPath(item *nav.TraverseItem) string {
	return o.target.JournalFullPath(item)
}

func (o *testPathFinderObserver) Statics() *common.StaticInfo {
	return o.target.Statics()
}

func (o *testPathFinderObserver) Scheme() string {
	return o.target.Scheme()
}

func (o *testPathFinderObserver) Observe(t common.PathFinder) common.PathFinder {
	o.target = t

	return o
}
