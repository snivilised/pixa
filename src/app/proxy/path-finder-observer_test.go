package proxy_test

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/traverse/lfs"
)

type splitPath struct {
	file       string
	folder     string
	sampleFile string
}
type pathAssertion struct {
	actual splitPath
	info   *common.PathInfo
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
// assertL

func (o *testPathFinderObserver) assertAll(entry *pixaTE,
	origin string, tfs lfs.TraverseFS,
) {
	if len(o.transfers) > 0 {
		first := lo.Keys(o.transfers)[0]
		fmt.Printf("\n 📂 TRANSFER FOLDER: '%v'\n", o.transfers[first].actual.folder)

		if !entry.dry && entry.asserters.transfer != nil {
			for input, assertion := range o.transfers {
				entry.asserters.transfer(entry, input, origin, assertion, tfs)
			}
		}
	}

	if len(o.results) > 0 && entry.asserters.result != nil {
		first := lo.Keys(o.results)[0]
		fmt.Printf("\n 📂 RESULT FOLDER: '%v'\n", o.results[first].actual.folder)

		if !entry.dry {
			for input, assertion := range o.results {
				// for loop iteration bug here, assertion is wrong
				//
				entry.asserters.result(entry, input, origin, assertion, tfs)
			}
		}
	}
}

func (o *testPathFinderObserver) Transfer(info *common.PathInfo) (folder, file string) {
	folder, file = o.target.Transfer(info)

	o.transfers[info.Item.Extension.Name] = &pathAssertion{
		actual: splitPath{
			folder: folder,
			file:   file,
		},
		info: info,
	}

	return folder, file
}

func (o *testPathFinderObserver) Result(info *common.PathInfo) (folder, file string) {
	folder, file = o.target.Result(info) // info.Item is wrong
	statics := o.Statics()
	o.results[info.Item.Extension.Name] = &pathAssertion{
		actual: splitPath{
			folder:     folder,
			file:       file,
			sampleFile: o.FileSupplement(info.Profile, statics.Sample), // !!! SampleFileSupplement
		},
		info: info,
	}

	return folder, file
}

func (o *testPathFinderObserver) FolderSupplement(profile string) string {
	return o.target.FolderSupplement(profile)
}

func (o *testPathFinderObserver) FileSupplement(profile, withSampling string) string {
	return o.target.FileSupplement(profile, withSampling)
}

func (o *testPathFinderObserver) SampleFileSupplement(withSampling string) string {
	return o.target.SampleFileSupplement(withSampling)
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
