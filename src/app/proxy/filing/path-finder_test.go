package filing_test

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/cfg"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/app/proxy/filing"
)

type reasons struct {
	folder string
	file   string
}

type asserter func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE)

type pfTE struct {
	given          string
	should         string
	reasons        reasons
	scheme         string
	profile        string
	supplement     string
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
			Abort: true,
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
			scheme, _ := schemes.Scheme(entry.scheme)
			noProfiles := lo.TernaryF(scheme == nil,
				func() uint {
					return 1
				},
				func() uint {
					return uint(len(scheme.Profiles()))
				},
			)
			finder := filing.NewFinder(&filing.NewFinderInfo{
				Advanced:   advanced,
				Schemes:    schemes,
				Scheme:     entry.scheme,
				OutputPath: entry.output,
				TrashPath:  entry.trash,
				Arity:      noProfiles,
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
			given:  "🌀 TRANSFER: transparent/profile/not-cuddled (🎯 @TID-CORE-1_TR-PR-NC_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			profile:        "blur",
			supplement:     filepath.Join("$TRASH$", "blur"),
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplement)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: transparent/profile/not-cuddled (🎯 @TID-CORE-2_TR-PR-NC_R)",
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
			// If we say transparent and cuddled, then that implies
			// - inputs and outputs should be in the same folder
			// - transparent means the output take the place of the input
			// - input has to be moved out of the way
			// - but which directory? By default we attempt to be transparent, so
			// the directory should be origin. This would be changed by the presence
			// of either --trash(input), --output(result).
			// - input renamed with supplement

			given:  "🌀 TRANSFER: transparent/profile/cuddled (🎯 @TID-CORE-3_TR-PR-CU_T)",
			should: "return origin folder // file decorated with supplement",
			reasons: reasons{
				folder: "origin folder to enable cuddle",
				file:   "input needs to be supplemented so result can be cuddled",
			},
			profile:        "blur",
			supplement:     fmt.Sprintf("%v.%v", "$TRASH$", "blur"),
			actionTransfer: true,
			cuddle:         true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				// Expect(folder).To(BeEmpty(), because(entry.reasons.folder))
				supplemented := filing.SupplementFilename(
					pi.Item.Extension.Name, entry.supplement, statics,
				)
				Expect(file).To(Equal(supplemented), because(entry.reasons.file, file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: transparent/profile/cuddled (🎯 @TID-CORE-4_TR-PR-CU_R)",
			should: "not modify folder // not modify filename",
			reasons: reasons{
				folder: "not modify folder to enable cuddle",
				file:   "cuddled un-supplemented result file needs to replace the input",
			},
			profile: "blur",
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// === TRANSPARENT / SCHEME (non-cuddle) [BLUE]

		Entry(nil, &pfTE{
			// NOT-TRANSPARENT:
			//
			// scheme not compatible with transparency, so this may not be a valid test
			// since arity > 1, how can we achieve transparency? we can't, because only
			// 1 result can take the place of the input.
			// This is not transparent, but the case may still be valid. SInce this
			// is not transparent, we either need:
			// - --trash: to transfer the input to this explicit location
			// - --output: to create new file as the output of operation
			// if neither are specified, the the input stays as is and the results are
			// decorated with the supplement, all origin.
			given:  "🌀 TRANSFER: not-transparent/scheme/not-cuddled (🎯 @TID-CORE-5_NTR-SC-NC_BLUR_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			scheme:         "blur-sf",
			profile:        "blur",
			supplement:     filepath.Join("$TRASH$", "blur-sf", "blur"),
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplement)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// ...

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: not-transparent/scheme/not-cuddled (🎯 @TID-CORE-6_NTR-SC-NC_SF_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			scheme:         "blur-sf",
			profile:        "sf",
			supplement:     filepath.Join("$TRASH$", "blur-sf", "sf"),
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplement)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// === TRANSPARENT / SCHEME (cuddle) [GREEN]

		Entry(nil, &pfTE{
			// !!! input file should stay unmodified
			// the results should be supplemented
			given:  "🌀 TRANSFER: not-transparent/scheme/cuddled (🎯 @TID-CORE-7_NTR-SC-CU_BLUR_T)",
			should: "return origin folder // file decorated with supplement",
			reasons: reasons{
				folder: "return origin folder to enable cuddle",
				file:   "cuddled file needs to be disambiguated from the input",
			},
			scheme:         "blur-sf",
			profile:        "blur",
			supplement:     "$TRASH$.blur-sf.blur",
			actionTransfer: true,
			cuddle:         true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				supplemented := filing.SupplementFilename(
					pi.Item.Extension.Name, entry.supplement, statics,
				)
				Expect(file).To(Equal(supplemented), because(entry.reasons.file, file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: not-transparent/scheme/cuddled (🎯 @TID-CORE-(7:_TBD_)_NTR-SC-CU_BLUR_T)",
			should: "return empty folder // file decorated with supplement",
			reasons: reasons{
				folder: "return empty folder to enable cuddle",
				file:   "cuddled file needs to be disambiguated from the input",
			},
			scheme:     "blur-sf",
			profile:    "blur",
			supplement: "blur-sf.blur",
			cuddle:     true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				supplemented := filing.SupplementFilename(
					pi.Item.Extension.Name, entry.supplement, statics,
				)
				Expect(file).To(Equal(supplemented), because(entry.reasons.file, file))
			},
		}),

		// ...

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: not-transparent/scheme/cuddled (🎯 @TID-CORE-8_NTR-SC-CU_SF_T)",
			should: "return origin folder // file decorated with supplement",
			reasons: reasons{
				folder: "return origin folder to enable cuddle",
				file:   "cuddled file needs to be disambiguated from the input",
			},
			scheme:         "blur-sf",
			profile:        "sf",
			supplement:     "$TRASH$.blur-sf.sf",
			actionTransfer: true,
			cuddle:         true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				supplemented := filing.SupplementFilename(
					pi.Item.Extension.Name, entry.supplement, statics,
				)
				Expect(file).To(Equal(supplemented), because(entry.reasons.file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: not-transparent/scheme/cuddled (🎯 @TID-CORE-(8:_TBD_)_NTR-SC-CU_SF_R)",
			should: "return origin folder // file decorated with supplement",
			reasons: reasons{
				folder: "return origin folder to enable cuddle",
				file:   "cuddled file needs to be disambiguated from the input",
			},
			scheme:     "blur-sf",
			profile:    "sf",
			supplement: "blur-sf.sf",
			cuddle:     true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(pi.Origin), because(entry.reasons.folder))
				supplemented := filing.SupplementFilename(
					pi.Item.Extension.Name, entry.supplement, statics,
				)
				Expect(file).To(Equal(supplemented), because(entry.reasons.file))
			},
		}),

		//
		// === TRANSPARENT / ADHOC
		//
		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: transparent/adhoc/not-cuddled (🎯 @TID-CORE-9_TR-AD-NC_SF_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			supplement:     filepath.Join("$TRASH$", "ADHOC"),
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplement)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: transparent/adhoc/not-cuddled (🎯 @TID-CORE-10_TR-AD-NC_SF_R)",
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
			profile:        "blur",
			trash:          filepath.Join("foo", "sessions", "scan01", "rubbish"),
			supplement:     filepath.Join("$TRASH$", "blur"),
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(entry.trash, entry.supplement)), because(entry.reasons.folder, folder))
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
			// NOT-TRANSPARENT
			//
			given:  "🌀 TRANSFER: not-transparent/scheme/trash (🎯 @TID-CORE-13_NTR-SC-TRA_BLUR_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way",
			},
			scheme:         "blur-sf",
			profile:        "blur",
			trash:          filepath.Join("foo", "sessions", "scan01", "rubbish"),
			supplement:     filepath.Join("rubbish", "$TRASH$", "blur-sf", "blur"),
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplement)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// ...

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: not-transparent/scheme/trash (🎯 @TID-CORE-14_NTR-SC-TRA_SF_T)",
			should: "redirect input to supplemented folder // filename not modified",
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way",
			},
			scheme:         "blur-sf",
			profile:        "sf",
			trash:          filepath.Join("foo", "sessions", "scan01", "rubbish"),
			supplement:     filepath.Join("rubbish", "$TRASH$", "blur-sf", "sf"),
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(Equal(filepath.Join(pi.Origin, entry.supplement)), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		//
		// NON-TRANSPARENT --output SPECIFIED
		//

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: not-transparent/profile/output (🎯 @TID-CORE-15_NTR-PR-NC-OUT_T)",
			should: "return empty folder and file",
			reasons: reasons{
				folder: "no transfer required",
				file:   "input file left alone",
			},
			profile:        "blur",
			output:         filepath.Join("foo", "sessions", "scan01", "results"),
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(BeEmpty(), because(entry.reasons.folder, folder))
				Expect(file).To(BeEmpty(), because(entry.reasons.file))
			},
		}),

		Entry(nil, &pfTE{
			given:  "🎁 RESULT: not-transparent/profile/output (🎯 @TID-CORE-16_NTR-PR-NC-OUT_R)",
			should: "redirect result to output // supplement folder // not modify filename",
			reasons: reasons{
				folder: "result should be send to supplemented output folder",
				file:   "filename only needs to match input filename because the folder is supplemented",
			},
			profile:    "blur",
			output:     filepath.Join("foo", "sessions", "scan01", "results"),
			supplement: "blur",
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				supplemented := filing.SupplementFolder(
					entry.output, entry.supplement,
				)
				Expect(folder).To(Equal(supplemented), because(entry.reasons.folder, folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),

		// === NON-TRANSPARENT / SCHEME (non-cuddle) [BLUE]

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: not-transparent/scheme/output (🎯 @TID-CORE-17_NTR-SC-NC-OUT_BLUR_T)",
			should: "return empty folder and file",
			reasons: reasons{
				folder: "no transfer required",
				file:   "input file left alone",
			},
			scheme:         "blur-sf",
			profile:        "blur",
			output:         filepath.Join("foo", "sessions", "scan01", "results"),
			actionTransfer: true,
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				Expect(folder).To(BeEmpty(), because(entry.reasons.folder))
				Expect(file).To(BeEmpty(), because(entry.reasons.file))
			},
		}),

		// ...

		Entry(nil, &pfTE{
			given:  "🌀 TRANSFER: not-transparent/scheme/output (🎯 @TID-CORE-18_NTR-SC-NC-OUT_SF_T)",
			should: "return empty folder and file",
			reasons: reasons{
				folder: "no transfer required",
				file:   "input file left alone",
			},
			scheme:         "blur-sf",
			profile:        "sf",
			output:         filepath.Join("foo", "sessions", "scan01", "results"),
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
			given:  "🌀 TRANSFER: not-transparent/adhoc/output (🎯 @TID-CORE-19_NTR-AD-OUT_SF_T)",
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
			given:  "🎁 RESULT: not-transparent/adhoc/output (🎯 @TID-CORE-20_NTR-AD-OUT_SF_R)",
			should: "redirect result to output // supplement folder // not modify filename",
			reasons: reasons{
				folder: "result should be send to supplemented output folder",
				file:   "filename only needs to match input filename because the folder is supplemented",
			},
			output:     filepath.Join("foo", "sessions", "scan01", "results"),
			supplement: "ADHOC",
			assert: func(folder, file string, pi *common.PathInfo, statics *common.StaticInfo, entry *pfTE) {
				supplemented := filing.SupplementFolder(
					entry.output, entry.supplement,
				)
				Expect(folder).To(Equal(supplemented), because(entry.reasons.folder))
				Expect(file).To(Equal(pi.Item.Extension.Name), because(entry.reasons.file))
			},
		}),
	)
})
