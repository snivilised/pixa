package filing

import (
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/cfg"
)

func NewFinder(
	inputs *common.ShrinkCommandInputs,
	advancedCFG cfg.AdvancedConfig,
	schemesCFG cfg.SchemesConfig,
) common.PathFinder {
	extensions := advancedCFG.Extensions()
	finder := &PathFinder{
		scheme:          inputs.Root.ProfileFam.Native.Scheme,
		ExplicitProfile: inputs.Root.ProfileFam.Native.Profile,
		Arity:           1,
		statics: &common.StaticInfo{
			Adhoc:  advancedCFG.AdhocLabel(),
			Legacy: advancedCFG.LegacyLabel(),
			Trash:  advancedCFG.TrashLabel(),
		},
		Ext: &ExtensionTransformation{
			Transformers: strings.Split(extensions.Transforms(), ","),
			Remap:        extensions.Map(),
		},
	}

	if finder.scheme != "" {
		schemeCFG, _ := schemesCFG.Scheme(finder.scheme)
		finder.Arity = len(schemeCFG.Profiles())
	}

	if inputs.ParamSet.Native.OutputPath != "" {
		finder.Output = inputs.ParamSet.Native.OutputPath
	} else {
		finder.transparentInput = true
	}

	if inputs.ParamSet.Native.TrashPath != "" {
		finder.Trash = inputs.ParamSet.Native.TrashPath
	}

	journal := advancedCFG.JournalLabel()

	if !strings.HasSuffix(journal, common.JournalExtension) {
		journal += common.JournalExtension
	}

	if !strings.HasPrefix(journal, common.JournalTag) {
		journal = common.JournalTag + journal
	}

	withoutExt := strings.TrimSuffix(journal, common.JournalExtension)
	core := strings.TrimPrefix(withoutExt, common.JournalTag)

	finder.statics.Meta = common.JournalMetaInfo{
		Core:       core,
		Journal:    journal,
		WithoutExt: withoutExt,
		Extension:  common.JournalExtension,
		Tag:        common.JournalTag,
	}

	return finder
}

type pfPath uint

const (
	pfPathUndefined pfPath = iota
	pfPathInputTransferFolder
	pfPathTxInputDestinationFolder
	pfPathInputDestinationFileOriginalExt
	pfPathResultFolder
	pfPathResultFile
)

const (
	inlineDestinationTempl = ""
)

type (
	templateSegments      []string
	pfTemplatesCollection map[pfPath]templateSegments
	pfFieldValues         map[string]string
)

var (
	pfTemplates pfTemplatesCollection
)

/*
📚 FIELD DICTIONARY:
- ADHOC: (static): tag that indicates no profile or scheme is active
- INPUT-DESTINATION: the path where the input file is moved to
- ITEM-FULL-NAME: the original item.Name, which includes the original extension
- OUTPUT-ROOT: --output flag
- ITEM-SUB-PATH: item.Extension.SubPath
- RESULT-NAME: the path of the result file
- SUPPLEMENT: ${{ADHOC}} | <scheme?>/<profile> --> created dynamically
- TRASH-LABEL: (static) input file tag marked for deletion
- DEJA-VU: a label that is meant to defend against accidental inclusion
by recursion, caused by a subsequent run in the same location. eg, a user
might run a sample shrink on ~/library/pics. This would create ~/library/pics/profile/TRASH.
The user may re-run on the same path, but this second run would see that old
trash path and attempt to process those files, by accident. To fix this,
~/library/pics/profile/TRASH would be replaced by ~/library/pics/<DEJA-VU>/profile/TRASH
where <DEJA-VU> would be skipped by he navigation process.
*/

func init() {
	pfTemplates = pfTemplatesCollection{
		pfPathInputTransferFolder: templateSegments{
			"${{INPUT-DESTINATION}}",
			"${{ITEM-SUB-PATH}}",
			"${{DEJA-VU}}",
			"${{SUPPLEMENT}}",
			"${{TRASH-LABEL}}",
		},
		pfPathTxInputDestinationFolder: templateSegments{
			"${{OUTPUT-ROOT}}",
		},
		pfPathInputDestinationFileOriginalExt: templateSegments{
			"${{ITEM-FULL-NAME}}",
		},
		pfPathResultFolder: templateSegments{
			"${{OUTPUT-ROOT}}",
			"${{ITEM-SUB-PATH}}",
			"${{SUPPLEMENT}}",
		},
		pfPathResultFile: templateSegments{
			"${{RESULT-NAME}}",
		},
	}
}

// expand returns a string as a result of joining the segments
func (tc pfTemplatesCollection) expand(segments ...string) string {
	return filepath.Join(segments...)
}

// evaluate returns a string representing a file system path from a
// template string containing place-holders and field values.
//
// Make sure that the keys of the values passed in match the segments.
// If they differ, then the result will contain unresolved segments (ie,
// 1 or more segments that are not evaluated and still contain the
// template placeholder.)
func (tc pfTemplatesCollection) evaluate(
	values pfFieldValues,
	segments ...string,
) string {
	// There is a very subtle but important point to note about the evaluate
	// method, in particular the parameters being passed in. It might seem
	// to the reader that the segments being passed in are redundant, because
	// they could be derived from the keys of the values map. However, map
	// entries do not have a guaranteed iteration order. Only arrays are
	// guaranteed to remain in the same order in which they were created. This
	// is the purpose of the segments parameter; it dictates the order in which
	// the segments of a path are evaluated. We can't even use the OrderedKeys
	// map, because entries are sorted lexically, which is not what we want.
	//
	const (
		quantity = 1
	)

	sourceTemplate := filepath.Join(segments...)
	result := lo.Reduce(segments, func(acc, field string, _ int) string {
		return strings.Replace(acc, field, values[field], quantity)
	},
		sourceTemplate,
	)

	return filepath.Clean(result)
}

type ExtensionTransformation struct {
	Transformers []string
	Remap        map[string]string
}

// PathFinder provides the common paths required, but its the controller that know
// the specific paths based around this common framework
type PathFinder struct {
	scheme          string
	ExplicitProfile string
	// Origin is the parent of the item (item.Parent)
	//
	Origin string

	// Output is the output as indicated by --output. If not set, then it is
	// derived:
	// - sampling: (inline) -- item.parent; => item.parent/SHRINK/<supplement>
	// - full: (inline) -- item.parent
	Output           string
	Trash            string
	Arity            int
	Ext              *ExtensionTransformation
	transparentInput bool
	statics          *common.StaticInfo
}

func (f *PathFinder) JournalFullPath(item *nav.TraverseItem) string {
	file := f.statics.JournalLocation(
		item.Extension.Name, item.Extension.Parent,
	)

	// ---> fmt.Printf("🔥🔥🔥 JOURNAL-FILE: '%v'\n", file)

	return file
}

func (f *PathFinder) Statics() *common.StaticInfo {
	return f.statics
}

func (f *PathFinder) Scheme() string {
	return f.scheme
}

// Transfer creates a path for the input; should return empty
// string for the folder, if no move is required (ie non transparent).
// The FileManager will only call this function when the input
// is not transparent. When the --Trash option is present, it will
// determine the destination path for the input.
func (f *PathFinder) Transfer(info *common.PathInfo) (folder, file string) {
	to := lo.TernaryF(f.Trash != "",
		func() string {
			return f.Trash
		},
		func() string {
			return info.Origin
		},
	)

	folder = func() string {
		segments := pfTemplates[pfPathInputTransferFolder]

		return pfTemplates.evaluate(pfFieldValues{
			"${{INPUT-DESTINATION}}": to,
			"${{ITEM-SUB-PATH}}":     info.Item.Extension.SubPath,
			"${{DEJA-VU}}":           DejaVu,
			"${{SUPPLEMENT}}":        f.supplement(),
			"${{TRASH-LABEL}}":       f.statics.Trash,
		}, segments...)
	}()

	file = func() string {
		segments := pfTemplates[pfPathInputDestinationFileOriginalExt]

		return pfTemplates.evaluate(pfFieldValues{
			"${{ITEM-FULL-NAME}}": info.Item.Extension.Name,
		}, segments...)
	}()

	return folder, file
}

func (f *PathFinder) mutateExtension(file string) string {
	extension := filepath.Ext(file)
	withoutDot := extension[1:]

	if _, found := f.Ext.Remap[withoutDot]; found {
		extension = "." + f.Ext.Remap[withoutDot]
	}

	if len(f.Ext.Transformers) > 0 {
		base := FilenameWithoutExtension(file)

		for _, transform := range f.Ext.Transformers {
			switch transform {
			case "lower":
				extension = strings.ToLower(extension)
			case "upper":
				extension = strings.ToUpper(extension)
			}
		}

		file = base + extension
	}

	return file
}

// Result creates a path for each result so should be called by the
// execution step
func (f *PathFinder) Result(info *common.PathInfo) (folder, file string) {
	to := lo.TernaryF(f.Output != "",
		func() string {
			return f.Output
		},
		func() string {
			return info.Origin
		},
	)

	folder = func() string {
		segments := pfTemplates[pfPathInputTransferFolder]

		return lo.TernaryF(f.transparentInput && f.Arity == 1,
			func() string {
				// The result file has to be in the same folder
				// as the input
				//
				segments = pfTemplates[pfPathTxInputDestinationFolder]

				return pfTemplates.evaluate(pfFieldValues{
					"${{OUTPUT-ROOT}}": info.Origin,
				}, segments...)
			},
			func() string {
				// If there is no scheme or profile, then the user is
				// only relying flags on the command line, ie running adhoc
				// so the result path should include an adhoc label. Otherwise,
				// the result should reflect the supplementary path.
				//
				return pfTemplates.evaluate(pfFieldValues{
					"${{OUTPUT-ROOT}}":   to,
					"${{SUPPLEMENT}}":    f.supplement(),
					"${{ITEM-SUB-PATH}}": info.Item.Extension.SubPath,
				}, segments...)
			},
		)
	}()

	file = func() string {
		// The file name just matches the input file name. The folder name
		// provides the context.
		//
		segments := pfTemplates[pfPathResultFile]

		return pfTemplates.evaluate(pfFieldValues{
			"${{RESULT-NAME}}": info.Item.Extension.Name,
		}, segments...)
	}()

	return folder, f.mutateExtension(file)
}

func (f *PathFinder) supplement() string {
	return lo.TernaryF(f.scheme == "" && f.ExplicitProfile == "",
		func() string {
			adhocLabel := f.statics.Adhoc
			return adhocLabel
		},
		func() string {
			return filepath.Join(f.scheme, f.ExplicitProfile)
		},
	)
}

func (f *PathFinder) TransparentInput() bool {
	return f.transparentInput
}
