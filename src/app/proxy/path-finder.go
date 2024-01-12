package proxy

import (
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/snivilised/extendio/xfs/nav"
)

type pfPath uint

const (
	pfPathUndefined pfPath = iota
	pfPathInputDestinationFolder
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
*/

func init() {
	pfTemplates = pfTemplatesCollection{
		pfPathInputDestinationFolder: templateSegments{
			"${{INPUT-DESTINATION}}",
			"${{ITEM-SUB-PATH}}",
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

type staticInfo struct {
	adhoc   string
	journal string
	legacy  string
	trash   string
}

type extensionTransformation struct {
	transformers []string
	remap        map[string]string
}

// PathFinder provides the common paths required, but its the controller that know
// the specific paths based around this common framework
type PathFinder struct {
	Scheme          string
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
	arity            int
	transparentInput bool
	statics          *staticInfo
	ext              *extensionTransformation
}

func (f *PathFinder) JournalFile(item *nav.TraverseItem) string {
	file := FilenameWithoutExtension(item.Extension.Name) + f.statics.journal

	return filepath.Join(item.Extension.Parent, file)
}

// Transfer creates a path for the input; should return empty
// string for the folder, if no move is required (ie non transparent).
// The FileManager will only call this function when the input
// is not transparent. When the --Trash option is present, it will
// determine the destination path for the input.
func (f *PathFinder) Transfer(info *pathInfo) (folder, file string) {
	to := lo.TernaryF(f.Trash != "",
		func() string {
			return f.Trash
		},
		func() string {
			return info.origin
		},
	)

	folder = func() string {
		segments := pfTemplates[pfPathInputDestinationFolder]

		return pfTemplates.evaluate(pfFieldValues{
			"${{INPUT-DESTINATION}}": to,
			"${{ITEM-SUB-PATH}}":     info.item.Extension.SubPath,
			"${{SUPPLEMENT}}":        f.supplement(),
			"${{TRASH-LABEL}}":       f.statics.trash,
		}, segments...)
	}()

	file = func() string {
		segments := pfTemplates[pfPathInputDestinationFileOriginalExt]

		return pfTemplates.evaluate(pfFieldValues{
			"${{ITEM-FULL-NAME}}": info.item.Extension.Name,
		}, segments...)
	}()

	return folder, file
}

func (f *PathFinder) mutateExtension(file string) string {
	extension := filepath.Ext(file)
	withoutDot := extension[1:]

	if _, found := f.ext.remap[withoutDot]; found {
		extension = "." + f.ext.remap[withoutDot]
	}

	if len(f.ext.transformers) > 0 {
		base := FilenameWithoutExtension(file)

		for _, transform := range f.ext.transformers {
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
func (f *PathFinder) Result(info *pathInfo) (folder, file string) {
	to := lo.TernaryF(f.Output != "",
		func() string {
			return f.Output
		},
		func() string {
			return info.origin
		},
	)

	folder = func() string {
		segments := pfTemplates[pfPathInputDestinationFolder]

		return lo.TernaryF(f.transparentInput && f.arity == 1,
			func() string {
				// The result file has to be in the same folder
				// as the input
				//
				segments = pfTemplates[pfPathTxInputDestinationFolder]

				return pfTemplates.evaluate(pfFieldValues{
					"${{OUTPUT-ROOT}}": info.origin,
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
					"${{ITEM-SUB-PATH}}": info.item.Extension.SubPath,
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
			"${{RESULT-NAME}}": info.item.Extension.Name,
		}, segments...)
	}()

	return folder, f.mutateExtension(file)
}

func (f *PathFinder) supplement() string {
	return lo.TernaryF(f.Scheme == "" && f.ExplicitProfile == "",
		func() string {
			adhocLabel := f.statics.adhoc
			return adhocLabel
		},
		func() string {
			return filepath.Join(f.Scheme, f.ExplicitProfile)
		},
	)
}
