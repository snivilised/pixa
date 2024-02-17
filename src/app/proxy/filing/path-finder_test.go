package filing_test

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/cfg"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/app/proxy/filing"
)

type supplements struct {
	folder string
	file   string
}

type reasons = supplements

type asserter func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE)

type pfTE struct {
	given          string
	should         string
	reasons        supplements
	scheme         string
	profile        string
	supplements    supplements
	output         string
	trash          string
	cuddle         bool
	dry            bool
	actionTransfer bool
	assert         asserter
}

func because(reason string, extras ...string) string {
	if len(extras) == 0 {
		return fmt.Sprintf("🔥 %v", reason)
	}

	return fmt.Sprintf("🔥 %v (%v)", reason, strings.Join(extras, ","))
}

var _ = Describe("PathFinder", Ordered, func() {
	var (
		advanced *cfg.MsAdvancedConfig
		schemes  *cfg.MsSchemesConfig
	)

	BeforeAll(func() {
		schemes = &cfg.MsSchemesConfig{
			"blur-sf": &cfg.MsSchemeConfig{
				ProfilesData: []string{"blur", "sf"},
			},
			"adaptive-sf": &cfg.MsSchemeConfig{
				ProfilesData: []string{"adaptive", "sf"},
			},
			"adaptive-blur": &cfg.MsSchemeConfig{
				ProfilesData: []string{"adaptive", "blur"},
			},
			"singleton": &cfg.MsSchemeConfig{
				ProfilesData: []string{"adaptive"},
			},
		}

		advanced = &cfg.MsAdvancedConfig{
			Abort: false,
			LabelsCFG: cfg.MsLabelsConfig{
				Adhoc:      "ADHOC",
				Journal:    "journal",
				Legacy:     ".LEGACY",
				Trash:      "TRASH",
				Fake:       ".FAKE",
				Supplement: "SUPP",
			},
			ExtensionsCFG: cfg.MsExtensionsConfig{
				FileSuffixes:  "jpg,jpeg,png",
				TransformsCSV: "lower",
				Remap: map[string]string{
					"jpeg": "jpg",
				},
			},
			ExecutableCFG: cfg.MsExecutableConfig{
				ProgramName:      "dummy",
				Timeout:          "1s",
				NoProgramRetries: 3,
			},
		}
	})

	DescribeTable("core",
		func(entry *pfTE) {
			finder := filing.NewFinder(&filing.NewFinderInfo{
				Advanced:   advanced,
				Schemes:    schemes,
				Scheme:     entry.scheme,
				OutputPath: entry.output,
				TrashPath:  entry.trash,
			})

			origin := filepath.Join("foo", "sessions", "scan01")
			pi := &common.PathInfo{
				Item: &nav.TraverseItem{
					Path: origin,
					Extension: nav.ExtendedItem{
						Name: "01_Backyard-Worlds-Planet-9_s01.jpg",
					},
				},
				Origin:  origin,
				Profile: entry.profile,
				Scheme:  entry.scheme,
				Cuddle:  entry.cuddle,
				Output:  entry.output,
				Trash:   entry.trash,
			}

			var (
				folder, file string
			)

			if entry.actionTransfer {
				folder, file = finder.Transfer(pi)
			} else {
				folder, file = finder.Result(pi)
			}
			entry.assert(folder, file, pi, finder.Statics(), entry)
		},
		func(entry *pfTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should: '%v'",
				entry.given, entry.should,
			)
		},

		//
		// === TRANSPARENT / PROFILE
		//

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/profile/non-cuddled (🎯 @TID-CORE-1_TR-PR-NC_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			profile: "blur",
			supplements: supplements{
				folder: filepath.Join("$TRASH$", "blur"),
			},
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplements.folder)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: transparent/profile (🎯 @TID-CORE-2_TR-PR-NC_R)",
			should: "not modify folder // not modify filename",
			reasons: reasons{
				folder: "transparency, result should take place of input",
				file:   "file should be moved out of the way and not cuddled",
			},
			profile: "blur",
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/profile/cuddled (🎯 @TID-CORE-3_TR-PR-CU_T)",
			should: "not modify folder // file decorated with supplement",
			reasons: reasons{
				folder: "not modify folder to enable cuddle",
				file:   "cuddled file needs to be disambiguated from the input",
			},
			profile: "blur",
			supplements: supplements{
				file: fmt.Sprintf("%v.%v", "$TRASH$", "blur"),
			},
			actionTransfer: true,
			cuddle:         true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				supplemented := filing.SupplementFilename(
					pi.Item.Extension.Name, entry.supplements.file, statics,
				)
				Expect(file).To(Equal(supplemented), because(entry.reasons.file, file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: transparent/profile/cuddled (🎯 @TID-CORE-4_TR-PR-CU_R)",
			should: "not modify folder // not modify filename",
			reasons: reasons{
				folder: "not modify folder to enable cuddle",
				file:   "cuddled result file needs to replace the input",
			},
			profile: "blur",
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// === TRANSPARENT / SCHEME (non-cuddle) [BLUE]

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/scheme/non-cuddled (🎯 @TID-CORE-5_TR-SC-NC_BLUR_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			scheme:  "blur-sf",
			profile: "blur",
			supplements: supplements{
				folder: filepath.Join("$TRASH$", "blur-sf", "blur"),
			},
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplements.folder)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// ...

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/scheme/non-cuddled (🎯 @TID-CORE-6_TR-SC-NC_SF_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			scheme:  "blur-sf",
			profile: "sf",
			supplements: supplements{
				folder: filepath.Join("$TRASH$", "blur-sf", "sf"),
			},
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplements.folder)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// === TRANSPARENT / SCHEME (cuddle) [GREEN]

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/scheme/cuddled (🎯 @TID-CORE-7_TR-SC-CU_BLUR_T)",
			should: "not modify folder // file decorated with supplement",
			reasons: reasons{
				folder: "not modify folder to enable cuddle",
				file:   "cuddled file needs to be disambiguated from the input",
			},
			scheme:  "blur-sf",
			profile: "blur",
			supplements: supplements{
				folder: "",
				file:   "$TRASH$.blur-sf.blur",
			},
			actionTransfer: true,
			cuddle:         true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				supplemented := filing.SupplementFilename(
					pi.Item.Extension.Name, entry.supplements.file, statics,
				)
				Expect(file).To(Equal(supplemented), because(entry.reasons.file, file))
			},
		}),

		// ...

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/scheme/cuddled (🎯 @TID-CORE-8_TR-SC-CU_SF_T)",
			should: "not modify folder // file decorated with supplement",
			reasons: reasons{
				folder: "not modify folder to enable cuddle",
				file:   "cuddled file needs to be disambiguated from the input",
			},
			scheme:  "blur-sf",
			profile: "sf",
			supplements: supplements{
				folder: "",
				file:   "$TRASH$.blur-sf.sf",
			},
			actionTransfer: true,
			cuddle:         true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				supplemented := filing.SupplementFilename(
					pi.Item.Extension.Name, entry.supplements.file, statics,
				)
				Expect(file).To(Equal(supplemented), because(entry.reasons.file))
			},
		}),

		//
		// === TRANSPARENT / ADHOC
		//
		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/adhoc/non-cuddled (🎯 @TID-CORE-9_TR-AD-NC_SF_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			supplements: supplements{
				folder: filepath.Join("$TRASH$", "ADHOC"),
			},
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplements.folder)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: transparent/adhoc (🎯 @TID-CORE-10_TR-AD_R)",
			should: "not modify folder // not modify filename",
			reasons: reasons{
				folder: "transparency, result should take place of input",
				file:   "file should be moved out of the way and not cuddled",
			},
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		//
		// TRANSPARENT --trash SPECIFIED
		//
		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/profile/trash (🎯 @TID-CORE-11_TR-PR-TRA_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			profile: "blur",
			trash:   filepath.Join("foo", "sessions", "scan01", "rubbish"),
			supplements: supplements{
				folder: filepath.Join("$TRASH$", "blur"),
			},
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(entry.trash, entry.supplements.folder)), because(entry.reasons.folder, folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: transparent/profile/trash (🎯 @TID-CORE-12_TR-PR-TRA_R)",
			should: "not modify folder // not modify filename",
			reasons: reasons{
				folder: "transparency, result should take place of input",
				file:   "file should be moved out of the",
			},
			profile: "blur",
			trash:   filepath.Join("foo", "sessions", "scan01", "rubbish"),
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// === TRANSPARENT / SCHEME (non-cuddle) [BLUE]

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/scheme/trash (🎯 @TID-CORE-13_TR-SC-TRA_BLUR_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way",
			},
			scheme:  "blur-sf",
			profile: "blur",
			trash:   filepath.Join("foo", "sessions", "scan01", "rubbish"),
			supplements: supplements{
				folder: filepath.Join("rubbish", "$TRASH$", "blur-sf", "blur"),
			},
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplements.folder)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// ...

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/scheme/trash (🎯 @TID-CORE-14_TR-SC-TRA_SF_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way",
			},
			scheme:  "blur-sf",
			profile: "sf",
			trash:   filepath.Join("foo", "sessions", "scan01", "rubbish"),
			supplements: supplements{
				folder: filepath.Join("rubbish", "$TRASH$", "blur-sf", "sf"),
			},
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplements.folder)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		//
		// NON-TRANSPARENT --output SPECIFIED
		//

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: profile/output (🎯 @TID-CORE-15_NT-PR-OUT_T)",
			should: "return empty folder and file",
			reasons: reasons{
				folder: "no transfer required",
				file:   "input file left alone",
			},
			profile: "blur",
			output:  filepath.Join("foo", "sessions", "scan01", "results"),
			supplements: supplements{
				folder: filepath.Join("$TRASH$", "blur"),
			},
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(BeEmpty(), because(entry.reasons.folder, folder))
				Expect(file).To(BeEmpty(), because(entry.reasons.file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: profile/output (🎯 @TID-CORE-16_NT-PR-OUT_R)",
			should: "redirect result to output // supplement folder // not modify filename",
			reasons: reasons{
				folder: "result should be send to supplemented output folder",
				file:   "filename only needs to match input filename because the folder is supplemented",
			},
			profile: "blur",
			output:  filepath.Join("foo", "sessions", "scan01", "results"),
			supplements: supplements{
				folder: "blur",
			},
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				supplemented := filing.SupplementFolder(
					entry.output, entry.supplements.folder,
				)
				Expect(folder).To(Equal(supplemented), because(entry.reasons.folder, folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// === NON-TRANSPARENT / SCHEME (non-cuddle) [BLUE]

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: scheme/output (🎯 @TID-CORE-17_NT-SC-OUT_BLUR_T)",
			should: "return empty folder and file",
			reasons: reasons{
				folder: "no transfer required",
				file:   "input file left alone",
			},
			scheme:  "blur-sf",
			profile: "blur",
			output:  filepath.Join("foo", "sessions", "scan01", "results"),
			supplements: supplements{
				folder: filepath.Join("rubbish", "$TRASH$", "blur-sf", "blur"),
			},
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(BeEmpty(), because(entry.reasons.folder))
				Expect(file).To(BeEmpty(), because(entry.reasons.file))
			},
		}),

		// ...

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: scheme/output (🎯 @TID-CORE-18_NT-SC-OUT_SF_T)",
			should: "return empty folder and file",
			reasons: reasons{
				folder: "no transfer required",
				file:   "input file left alone",
			},
			scheme:  "blur-sf",
			profile: "sf",
			output:  filepath.Join("foo", "sessions", "scan01", "results"),
			supplements: supplements{
				folder: filepath.Join("rubbish", "$TRASH$", "blur-sf", "sf"),
			},
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(BeEmpty(), because(entry.reasons.folder))
				Expect(file).To(BeEmpty(), because(entry.reasons.file))
			},
		}),

		//
		// === NON-TRANSPARENT / ADHOC
		//
		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: adhoc/output (🎯 @TID-CORE-19_NT-AD-OUT_SF_T)",
			should: "return empty folder and file",
			reasons: reasons{
				folder: "no transfer required",
				file:   "input file left alone",
			},
			output:         filepath.Join("foo", "sessions", "scan01", "results"),
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(BeEmpty(), because(entry.reasons.folder))
				Expect(file).To(BeEmpty(), because(entry.reasons.file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: adhoc/output (🎯 @TID-CORE-20_NT-AD-OUT_SF_R)",
			should: "redirect result to output // supplement folder // not modify filename",
			reasons: reasons{
				folder: "result should be send to supplemented output folder",
				file:   "filename only needs to match input filename because the folder is supplemented",
			},
			output: filepath.Join("foo", "sessions", "scan01", "results"),
			supplements: supplements{
				folder: "ADHOC",
			},
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				supplemented := filing.SupplementFolder(
					entry.output, entry.supplements.folder,
				)
				Expect(folder).To(Equal(supplemented), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),
	)
})
