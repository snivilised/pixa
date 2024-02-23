package proxy_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/command"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/app/proxy/filing"

	"github.com/snivilised/pixa/src/internal/helpers"
	"github.com/snivilised/pixa/src/internal/matchers"
)

type reasons struct {
	folder string
	file   string
}

type arrange func(entry *pixaTE, origin string, vfs storage.VirtualFS)

type asserter func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS)

type asserters struct {
	transfer asserter
	result   asserter
	bs       *command.Bootstrap
}

func assertTransfer(folder string, pa *pathAssertion, vfs storage.VirtualFS) {
	actualDestination := filepath.Join(pa.actual.folder, pa.actual.file)
	Expect(matchers.AsDirectory(folder)).To(matchers.ExistInFS(vfs),
		because(actualDestination, "🌀 TRANSFER"),
	)

	file := filepath.Join(folder, pa.info.Item.Extension.Name)
	Expect(matchers.AsFile(file)).To(matchers.ExistInFS(vfs), because(actualDestination))
}

func assertTransferSupplementedOrigin(name string,
	entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS,
) {
	_ = name

	folder := filing.SupplementFolder(origin,
		entry.supplements.folder,
	)
	assertTransfer(folder, pa, vfs)
}

func assertResultItemFile(name string,
	entry *pixaTE, origin string, pa *pathAssertion,
) {
	_ = name
	_ = entry
	_ = origin
	// We don't have anything that actually creates the result file
	// so instead of checking that it exists in the file system, we
	// check the path is what we expect.
	//
	file := pa.info.Item.Path
	Expect(file).To(Equal(pa.info.Item.Path), because(file, "🎁 RESULT"))
}

type pixaTE struct {
	given              string
	should             string
	reasons            reasons
	arranger           arrange
	asserters          asserters
	exists             bool
	args               []string
	isTui              bool
	dry                bool
	intermediate       string
	output             string
	trash              string
	profile            string
	scheme             string
	relative           string
	mandatory          []string
	supplements        supplements
	inputs             []string
	configTestFilename string
}

func because(reason string, extras ...string) string {
	if len(extras) == 0 {
		return fmt.Sprintf("🔥 %v", reason)
	}

	return fmt.Sprintf("🔥 %v (%v)", reason, strings.Join(extras, ","))
}

func augment(entry *pixaTE,
	args []string, vfs storage.VirtualFS, root, directory string,
) []string {
	result := args
	result = append(result, entry.args...)

	if entry.exists {
		location := filepath.Join(directory, entry.intermediate, entry.supplements.folder)
		if err := vfs.MkdirAll(location, common.Permissions.Write); err != nil {
			Fail(errors.Wrap(err, err.Error()).Error())
		}
	}

	if entry.output != "" {
		output := helpers.Path(root, entry.output)
		result = append(result, "--output", output)
	}

	if entry.trash != "" {
		result = append(result, "--trash", entry.trash)
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
	vfs             storage.VirtualFS
	withoutRenderer bool
}

func (t *coreTest) run() {
	origin := helpers.Path(t.root, t.entry.relative)

	if t.entry.arranger != nil {
		t.entry.arranger(t.entry, origin, t.vfs)
	}

	args := augment(t.entry,
		[]string{
			common.Definitions.Commands.Shrink, origin,
		},
		t.vfs, t.root, origin,
	)

	observer := &testPathFinderObserver{
		transfers: make(observerAssertions),
		results:   make(observerAssertions),
	}

	bootstrap := command.Bootstrap{
		Vfs: t.vfs,
		Presentation: common.PresentationOptions{
			WithoutRenderer: t.withoutRenderer,
		},
		Observers: common.Observers{
			PathFinder: observer,
		},
	}

	configTestFilename := common.Definitions.Pixa.ConfigTestFilename
	if t.entry.configTestFilename != "" {
		configTestFilename = t.entry.configTestFilename
	}

	tester := helpers.CommandTester{
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

	observer.assertAll(t.entry, origin, t.vfs)
}

var _ = Describe("pixa", Ordered, func() {
	var (
		repo            string
		l10nPath        string
		configPath      string
		root            string
		vfs             storage.VirtualFS
		withoutRenderer bool
	)

	BeforeAll(func() {
		repo = helpers.Repo("")
		l10nPath = helpers.Path(repo, "test/data/l10n")
		configPath = helpers.Path(repo, "test/data/configuration")

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
		vfs, root = helpers.SetupTest(
			"nasa-scientist-index.xml", configPath, l10nPath, helpers.Silent,
		)
	})

	DescribeTable("interactive",
		func(entry *pixaTE) {
			core := coreTest{
				entry:           entry,
				root:            root,
				configPath:      configPath,
				vfs:             vfs,
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
			supplements: supplements{
				file:   "$TRASH$.blur",
				folder: filepath.Join("$TRASH$", "blur"),
			},
			inputs: helpers.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {
					assertTransferSupplementedOrigin(name, entry, origin, pa, vfs)
				},
				result: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {
					assertResultItemFile(name, entry, origin, pa)
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
			supplements: supplements{
				file:   "$TRASH$.ADHOC",
				folder: filepath.Join("$TRASH$", "ADHOC"),
			},
			inputs: helpers.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {
					assertTransferSupplementedOrigin(name, entry, origin, pa, vfs)
				},
				result: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {
					assertResultItemFile(name, entry, origin, pa)
				},
			},
		}),
		//
		// TRANSPARENT --trash SPECIFIED
		//
		Entry(nil, &pixaTE{
			given:    "regex/transparent/adhoc/not-cuddled (🎯 @TID-CORE-11/12:_TBD__TR-PR-TRA_TR)",
			should:   "transfer input to supplemented folder // input filename not modified",
			relative: BackyardWorldsPlanet9Scan01,
			reasons: reasons{
				folder: "transparency, result should take place of input",
				file:   "file should be moved out of the way to specified trash and result not cuddled",
			},
			arranger: func(entry *pixaTE, origin string, vfs storage.VirtualFS) {
				p := filepath.Join(origin, entry.trash)
				entry.trash = p
				_ = vfs.MkdirAll(p, common.Permissions.Write.Perm())
			},
			profile: "blur",
			trash:   "rubbish",
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplements: supplements{
				file:   "$TRASH$.ADHOC",
				folder: filepath.Join("$TRASH$", "blur"),
			},
			inputs: helpers.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {
					folder := filing.SupplementFolder(entry.trash,
						entry.supplements.folder,
					)

					assertTransfer(folder, pa, vfs)
				},
				result: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {},
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
			arranger: func(entry *pixaTE, origin string, vfs storage.VirtualFS) {
				_ = vfs.MkdirAll(entry.output, common.Permissions.Write.Perm())
			},
			profile: "blur",
			output:  filepath.Join("foo", "sessions", "scan01", "results"),
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplements: supplements{
				folder: "blur",
			},
			inputs: helpers.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {
					folder := filing.SupplementFolder(entry.output,
						entry.supplements.folder,
					)

					assertTransfer(folder, pa, vfs)
				},
				result: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {},
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
			arranger: func(entry *pixaTE, origin string, vfs storage.VirtualFS) {
				_ = vfs.MkdirAll(entry.output, common.Permissions.Write.Perm())
			},
			output: filepath.Join("foo", "sessions", "scan01", "results"),
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplements: supplements{
				folder: "blur-sf", // !! +blue/sf
			},
			inputs: helpers.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				// transfer: not transparent; no transfer is invoked
				result: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {
					assertResultItemFile(name, entry, origin, pa)
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
			arranger: func(entry *pixaTE, origin string, vfs storage.VirtualFS) {
				_ = vfs.MkdirAll(entry.output, common.Permissions.Write.Perm())
			},
			output: filepath.Join("foo", "sessions", "scan01", "results"),
			args: []string{
				"--files-rx", "Backyard-Worlds",
				"--gaussian-blur", "0.51",
				"--interlace", "line",
			},
			intermediate: "nasa/exo/Backyard Worlds - Planet 9/sessions/scan-01",
			supplements: supplements{
				folder: "ADHOC",
			},
			inputs: helpers.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				// transfer: not transparent; no transfer is invoked
				result: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {
					assertResultItemFile(name, entry, origin, pa)
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
			supplements: supplements{
				file:   "$TRASH$.blur",
				folder: filepath.Join("$TRASH$", "blur"),
			},
			inputs: helpers.BackyardWorldsPlanet9Scan01First6,
			asserters: asserters{
				transfer: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {
				},
				result: func(name string, entry *pixaTE, origin string, pa *pathAssertion, vfs storage.VirtualFS) {
				},
			},
		}),
	)
})
