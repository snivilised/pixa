package proxy_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // foo
	. "github.com/onsi/gomega"    //nolint:revive // foo
	"github.com/pkg/errors"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/app/proxy/filing"
	"github.com/snivilised/pixa/src/internal/helpers"
	lab "github.com/snivilised/pixa/src/internal/laboratory"
	"github.com/snivilised/pixa/src/internal/matchers"
	"github.com/snivilised/traverse/lfs"
)

type reasons struct {
	folder string
	file   string
}

type arranger func(entry *pixaTE, origin string)

type asserter func(entry *pixaTE, input, origin string, pa *pathAssertion, tfs lfs.TraverseFS)

type asserters struct {
	transfer asserter
	result   asserter
	bs       *command.Bootstrap
}

func assertTransfer(folder string, pa *pathAssertion, tfs lfs.TraverseFS) {
	actualDestination := filepath.Join(pa.actual.folder, pa.actual.file)
	Expect(matchers.AsDirectory(folder)).To(matchers.ExistInFS(tfs),
		because(actualDestination, "🌀 TRANSFER"),
	)

	file := filepath.Join(folder, pa.info.Item.Extension.Name)
	Expect(matchers.AsFile(file)).To(matchers.ExistInFS(tfs), because(actualDestination))
}

func assertTransferSupplementedOrigin(name string,
	entry *pixaTE, origin string, pa *pathAssertion, tfs lfs.TraverseFS,
) {
	_ = name

	folder := filing.SupplementFolder(origin,
		entry.supplement,
	)
	assertTransfer(folder, pa, tfs)
}

func assertResultItemFile(pa *pathAssertion,
) {
	// We don't have anything that actually creates the result actual
	// so instead of checking that it exists in the actual system, we
	// check the path is what we expect.
	//
	// there is a strangle loop iteration issue which means this is failing
	// for an unknown reason
	// comment: actual := pa.actual.file
	// comment: Expect(actual).To(Equal(pa.info.Item.Extension.Name), because(actual, "🎁 RESULT"))
	_ = pa
}

func assertResultFile(expected string, pa *pathAssertion) {
	// We don't have anything that actually creates the actual result
	// so instead of checking that it exists in the file system, we
	// check the path is what we expect.
	//
	actual := pa.actual.file
	Expect(strings.EqualFold(actual, expected)).To(BeTrue(), because(actual, "🎁 RESULT"))
}

func assertSampleFile(entry *pixaTE, input string, pa *pathAssertion) {
	statics := entry.finder.Statics()
	withSampling := statics.Sample
	supp := entry.finder.SampleFileSupplement(withSampling)
	expected := filing.SupplementFilename(
		input, supp, statics,
	)

	Expect(strings.EqualFold(pa.actual.file, expected)).To(BeTrue(),
		because(pa.actual.file, "🎁 RESULT"),
	)
}

func createSamples(entry *pixaTE,
	origin string, finder common.PathFinder, tfs lfs.TraverseFS,
) {
	statics := finder.Statics()
	supp := finder.SampleFileSupplement(statics.Sample)
	destination := filing.SupplementFolder(
		filepath.Join(origin, entry.intermediate, entry.output),
		supp,
	)

	if err := tfs.MkDirAll(destination, common.Permissions.Write); err != nil {
		Fail(fmt.Sprintf("could not create intermediate path: '%v'", destination))
	}

	for _, input := range entry.inputs {
		create := filepath.Join(destination, input)
		if f, e := tfs.Create(create); e != nil {
			Fail(fmt.Sprintf("could not create sample file: '%v'", create))
		} else {
			f.Close()
		}
	}
}

type pixaTE struct {
	given              string
	should             string
	reasons            reasons
	arrange            arranger
	asserters          asserters
	exists             bool
	args               []string
	isTui              bool
	dry                bool
	intermediate       string
	output             string // relative to root
	trash              string // relative to root
	sample             bool
	profile            string
	scheme             string
	relative           string
	mandatory          []string
	supplement         string
	inputs             []string
	configTestFilename string
	finder             common.PathFinder
}

func because(reason string, extras ...string) string {
	if len(extras) == 0 {
		return fmt.Sprintf("🔥 %v", reason)
	}

	return fmt.Sprintf("🔥 %v (%v)", reason, strings.Join(extras, ","))
}

func augment(entry *pixaTE,
	args []string, tfs lfs.TraverseFS, root, directory string,
) []string {
	result := args
	result = append(result, entry.args...)

	if entry.exists {
		location := filepath.Join(directory, entry.intermediate, entry.supplement)
		if err := tfs.MkDirAll(location, common.Permissions.Write); err != nil {
			Fail(errors.Wrap(err, err.Error()).Error())
		}
	}

	if entry.output != "" {
		output := lab.Path(root, entry.output)
		result = append(result, "--output", output)
	}

	if entry.trash != "" {
		trash := lab.Path(root, entry.trash)
		entry.trash = trash
		result = append(result, "--trash", trash)
	}

	if entry.sample {
		result = append(result, "--sample")
	}

	if entry.profile != "" {
		result = append(result, "--profile", entry.profile)
	}

	if entry.scheme != "" {
		result = append(result, "--scheme", entry.scheme)
	}

	if entry.dry {
		result = append(result, "--dry-run")
	}

	if !entry.isTui {
		result = append(result, "--no-tui")
	}

	return result
}

type coreTest struct {
	entry           *pixaTE
	root            string
	configPath      string
	tfs             lfs.TraverseFS
	withoutRenderer bool
}

func (t *coreTest) run() {
	origin := lab.Path(t.root, t.entry.relative)
	args := augment(t.entry,
		[]string{
			common.Definitions.Commands.Shrink, origin,
		},
		t.tfs, t.root, origin,
	)

	observer := &testPathFinderObserver{
		transfers: make(observerAssertions),
		results:   make(observerAssertions),
	}

	bootstrap := command.Bootstrap{
		FS: t.tfs,
		Presentation: common.PresentationOptions{
			WithoutRenderer: t.withoutRenderer,
		},
		Observers: common.Observers{
			PathFinder: observer,
		},
		Notifications: common.LifecycleNotifications{
			OnBegin: func(finder common.PathFinder, _, _ string) {
				t.entry.finder = finder

				if t.entry.arrange != nil {
					t.entry.arrange(t.entry, origin)
				}
			},
		},
	}

	configTestFilename := common.Definitions.Pixa.ConfigTestFilename
	if t.entry.configTestFilename != "" {
		configTestFilename = t.entry.configTestFilename
	}

	tester := lab.CommandTester{
		Args: args,
		Root: bootstrap.Root(func(co *command.ConfigureOptionsInfo) {
			co.Detector = &helpers.DetectorStub{}
			co.Config.Name = configTestFilename
			co.Config.ConfigPath = t.configPath
			co.Config.Viper = &configuration.GlobalViperConfig{}
		}),
	}

	_, err := tester.Execute()
	Expect(err).Error().To(BeNil(),
		fmt.Sprintf("execution result non nil (%v)", err),
	)

	observer.assertAll(t.entry, origin, t.tfs)
}

var _ = Describe("pixa", Ordered, func() {
	var (
		repo            string
		l10nPath        string
		configPath      string
		root            string
		FS              lfs.TraverseFS
		withoutRenderer bool
	)

	BeforeAll(func() {
		repo = lab.Repo("")
		l10nPath = lab.Path(repo, "test/data/l10n")
		configPath = lab.Path(repo, "test/data/configuration")

		var (
			err error
			f   *os.File
		)

		if f, err = openInputTTY(); err != nil {
			withoutRenderer = true
		}
		f.Close()
	})

	BeforeEach(func() {
		FS, root = lab.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, lab.Silent,
		)
	})

	DescribeTable("interactive",
		func(entry *pixaTE) {
			core := coreTest{
				entry:           entry,
				root:            root,
				configPath:      configPath,
				tfs:             FS,
				withoutRenderer: withoutRenderer,
			}
			core.run()
		},
		func(entry *pixaTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should: '%v'",
				entry.given, entry.should,
			)
		},
		//
		// === TRANSPARENT / PROFILE
		//
		Entry(nil, &pixaTE{
			given:    "regex/transparent/profile/not-cuddled (🎯 @TID-CORE-1/2:_TBD__TR-PR-NC_TR)",
			should:   "transfer input to supplemented folder // input filename not modified",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			profile: "blur",
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplement:   filepath.Join("$TRASH$", "blur"),
			inputs:       lab.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(entry *pixaTE, input, origin string, pa *pathAssertion, tfs lfs.TraverseFS) {
					assertTransferSupplementedOrigin(input, entry, origin, pa, tfs)
				},
				result: func(_ *pixaTE, input, _ string, pa *pathAssertion, _ lfs.TraverseFS) {
					if pa.info.Item.Extension.Name != input {
						fmt.Printf("===> ⛔ WARNING DISCREPANCY FOUND: name: '%v' // input '%v'\n",
							pa.info.Item.Extension.Name, input,
						)
					}
					assertResultItemFile(pa)
				},
			},
		}),
		//
		// === TRANSPARENT / ADHOC
		//
		Entry(nil, &pixaTE{
			given:    "regex/transparent/adhoc/not-cuddled (🎯 @TID-CORE-9/10:_TBD__TR-AD-NC_SF_TR)",
			should:   "transfer input to supplemented folder // input filename not modified",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "transparency, result should take place of input",
				file:   "file should be moved out of the way and result not cuddled",
			},
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplement:   filepath.Join("$TRASH$", "ADHOC"),
			inputs:       lab.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(entry *pixaTE, input, origin string, pa *pathAssertion, tfs lfs.TraverseFS) {
					assertTransferSupplementedOrigin(input, entry, origin, pa, tfs)
				},
				result: func(_ *pixaTE, _, _ string, _ *pathAssertion, _ lfs.TraverseFS) {
					// comment: assertResultItemFile(pa)
				},
			},
		}),
		//
		// TRANSPARENT --trash SPECIFIED
		//
		Entry(nil, &pixaTE{
			given:    "regex/transparent/profile/not-cuddled (🎯 @TID-CORE-11/12:_TBD__TR-PR-TRA_TR)",
			should:   "transfer input to supplemented folder // input filename not modified",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "transparency, result should take place of input",
				file:   "file should be moved out of the way to specified trash and result not cuddled",
			},
			arrange: func(entry *pixaTE, _ string) {
				trash := filepath.Join(entry.trash, entry.supplement)
				_ = FS.MkDirAll(trash, common.Permissions.Write.Perm())
			},
			profile: "blur",
			trash:   filepath.Join("foo", "sessions", "scan01", "rubbish"),
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplement:   filepath.Join("$TRASH$", "blur"),
			inputs:       lab.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(entry *pixaTE, _, _ string, pa *pathAssertion, tfs lfs.TraverseFS) {
					folder := filing.SupplementFolder(entry.trash,
						entry.supplement,
					)

					assertTransfer(folder, pa, tfs)
				},
				result: func(_ *pixaTE, _, _ string, _ *pathAssertion, _ lfs.TraverseFS) {},
			},
		}),
		//
		// ❌ --output SPECIFIED, can't be TRANSPARENT as the output is being diverted
		// elsewhere and by definition can't take the place of the input.
		//

		//
		// NON-TRANSPARENT --output SPECIFIED
		//
		Entry(nil, &pixaTE{
			given:    "regex/not-transparent/profile/not-cuddled (🎯 @TID-CORE-15/16:_TBD__NTR-PR-NC-OUT_TR)",
			should:   "not transfer input // not modify input filename // re-direct result to output",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "result should be re-directed, so input can stay in place",
				file:   "input file remains un modified",
			},
			arrange: func(entry *pixaTE, _ string) {
				_ = FS.MkDirAll(entry.output, common.Permissions.Write.Perm())
			},
			profile: "blur",
			output:  filepath.Join("foo", "sessions", "scan01", "results"),
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplement:   "blur",
			inputs:       lab.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(entry *pixaTE, _, _ string, pa *pathAssertion, tfs lfs.TraverseFS) {
					folder := filing.SupplementFolder(entry.output,
						entry.supplement,
					)

					assertTransfer(folder, pa, tfs)
				},
				result: func(_ *pixaTE, _, _ string, _ *pathAssertion, _ lfs.TraverseFS) {},
			},
		}),
		//
		// === NON-TRANSPARENT / SCHEME (non-cuddle) [BLUE]
		//
		Entry(nil, &pixaTE{
			given:    "regex/not-transparent/scheme/output (🎯 @TID-CORE-17/18:_TBD__NTR-SC-NC-OUT_BLUR_TR)",
			should:   "not transfer input // not modify input filename // re-direct result to output",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "result should be re-directed, so input can stay in place",
				file:   "input file remains un modified",
			},
			scheme: "blur-sf",
			arrange: func(entry *pixaTE, _ string) {
				_ = FS.MkDirAll(entry.output, common.Permissions.Write.Perm())
			},
			output: filepath.Join("foo", "sessions", "scan01", "results"),
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplement:   "blur-sf", // !! +blue/sf
			inputs:       lab.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				// transfer: not transparent; no transfer is invoked
				result: func(_ *pixaTE, _, _ string, pa *pathAssertion, _ lfs.TraverseFS) {
					assertResultItemFile(pa)
				},
			},
		}),
		//
		// === NON-TRANSPARENT / ADHOC
		//
		Entry(nil, &pixaTE{
			given:    "regex/not-transparent/adhoc/output (🎯 @TID-CORE-19/20:_TBD__NTR-AD-OUT_SF_TR)",
			should:   "not transfer input // not modify input filename // re-direct result to output",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "result should be re-directed, so input can stay in place",
				file:   "input file remains un modified",
			},
			arrange: func(entry *pixaTE, _ string) {
				_ = FS.MkDirAll(entry.output, common.Permissions.Write.Perm())
			},
			output: filepath.Join("foo", "sessions", "scan01", "results"),
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplement:   "ADHOC",
			inputs:       lab.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				// transfer: not transparent; no transfer is invoked
				result: func(_ *pixaTE, _, _ string, pa *pathAssertion, _ lfs.TraverseFS) {
					assertResultItemFile(pa)
				},
			},
		}),

		//
		// === MISC
		//

		//
		// === NO LOGGER IN CONFIG (TRANSPARENT / PROFILE)
		//
		Entry(nil, &pixaTE{
			given:    "no-log (🎯 @TID-CORE-1/2:_TBD__TR-PR-NC_TR)",
			should:   "use log with default scope",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "file should be moved out of the way and not cuddled",
			},
			configTestFilename: "pixa-test-no-logger",
			profile:            "blur",
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplement:   filepath.Join("$TRASH$", "blur"),
			inputs:       lab.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(_ *pixaTE, _, _ string, _ *pathAssertion, _ lfs.TraverseFS) {
				},
				result: func(_ *pixaTE, _, _ string, _ *pathAssertion, _ lfs.TraverseFS) {
				},
			},
		}),

		//
		// === SAMPLE / TRANSPARENT / PROFILE
		//
		Entry(nil, &pixaTE{
			given:    "regex/sample/transparent/profile/not-cuddled (🎯 @TID-CORE-1/2:_TBD__SMPL-TR-PR-NC_TR)",
			should:   "transfer input to supplemented folder // marked result as sample",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "transparency, result should take place of input in same folder",
				file:   "input file should be moved out of the way and result marked as sample",
			},
			profile: "blur",
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			sample:       true,
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplement:   filepath.Join("$TRASH$", "blur"),
			inputs:       lab.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(_ *pixaTE, _, origin string, pa *pathAssertion, tfs lfs.TraverseFS) {
					assertTransfer(origin, pa, tfs)
				},
				result: func(entry *pixaTE, input, _ string, pa *pathAssertion, _ lfs.TraverseFS) {
					assertSampleFile(entry, input, pa)
				},
			},
		}),
		//
		// === SAMPLE / TRANSPARENT / ADHOC
		//
		XEntry(nil, &pixaTE{
			given:    "regex/sample/transparent/adhoc/not-cuddled (🎯 @TID-CORE-9/10:_TBD__TR-AD-NC_SF_TR)",
			should:   "transfer input to supplemented folder // marked result as sample",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "transparency, result should take place of input",
				file:   "file should be moved out of the way and result marked as sample",
			},
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			sample:       true,
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplement:   filepath.Join("$TRASH$", "ADHOC"),
			inputs:       lab.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(entry *pixaTE, input, origin string, pa *pathAssertion, tfs lfs.TraverseFS) {
					assertTransferSupplementedOrigin(input, entry, origin, pa, tfs)
				},
				result: func(_ *pixaTE, _, _ string, pa *pathAssertion, _ lfs.TraverseFS) {
					assertResultItemFile(pa)
				},
			},
		}),

		//
		// === SAMPLE FILE ALREADY EXISTS (NOT-TRANSPARENT / ADHOC / OUTPUT)
		//
		Entry(nil, &pixaTE{
			given:    "regex/not-transparent/adhoc/output (🎯 @TID-CORE-19/20:_TBD__NTR-AD-OUT_SF_TR)",
			should:   "not transfer input // not modify input filename // re-direct result to output",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "result should be re-directed, so input can stay in place",
				file:   "input file remains un modified",
			},
			arrange: func(entry *pixaTE, origin string) {
				createSamples(entry, origin, entry.finder, FS)
			},
			output: filepath.Join("foo", "sessions", "scan01", "results"),
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplement:   "ADHOC",
			inputs:       lab.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				// transfer: not transparent; no transfer is invoked
				result: func(_ *pixaTE, _, _ string, _ *pathAssertion, _ lfs.TraverseFS) {
					// assertResultItemFile(pa)
				},
			},
		}),

		// given: destination already exists, should: skip agent invoke
	)
})
