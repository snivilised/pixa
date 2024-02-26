package filing

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type NewFinderInfo struct {
	Advanced   common.AdvancedConfig
	Schemes    common.SchemesConfig
	Scheme     string
	Profile    string
	OutputPath string
	TrashPath  string
	Observer   common.PathFinder
	Arity      uint
}

func NewFinder(
	info *NewFinderInfo,
) common.PathFinder {
	advanced := info.Advanced
	extensions := advanced.Extensions()
	finder := &PathFinder{
		Sch:   info.Scheme,
		Stats: common.NewStaticInfoFromConfig(advanced),
		Ext: &ExtensionTransformation{
			Transformers: strings.Split(extensions.Transforms(), ","),
			Remap:        extensions.Map(),
		},
	}

	finder.init(info)

	if info.Observer != nil {
		return info.Observer.Observe(finder)
	}

	return finder
}

type pfPath uint

const (
	pfPathUndefined pfPath = iota
	pfPathInputTransferFolder
	pfPathTxInputDestinationFolder
	pfPathResultFolder
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
- TRANSFER-DESTINATION: the path where the input file is moved to
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
			"${{TRANSFER-DESTINATION}}",
			"${{ITEM-SUB-PATH}}",
			"${{DEJA-VU}}",
			"${{SUPPLEMENT}}",
		},
		pfPathTxInputDestinationFolder: templateSegments{
			"${{OUTPUT-ROOT}}",
		},
		pfPathResultFolder: templateSegments{
			"${{OUTPUT-ROOT}}",
			"${{ITEM-SUB-PATH}}",
			"${{SUPPLEMENT}}",
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
	Sch    string
	Origin string // Origin is the parent of the item (item.Parent)

	// Output is the output as indicated by --output. If not set, then it is
	// derived:
	// - sampling: (inline) -- item.parent; => item.parent/SHRINK/<supplement>
	// - full: (inline) -- item.parent
	Output           string
	Trash            string
	Ext              *ExtensionTransformation
	transparentInput bool
	Stats            *common.StaticInfo
}

func (f *PathFinder) init(info *NewFinderInfo) {
	f.Output = info.OutputPath
	f.Trash = info.TrashPath

	// When the user specifies an alternative location for the results to be sent to
	// with the --output flag, then the input is no longer transparent, as the user has
	// to go to the output location to see the result.
	f.transparentInput = info.OutputPath == "" && info.Arity == 1
}

func (f *PathFinder) JournalFullPath(item *nav.TraverseItem) string {
	file := f.Stats.JournalLocation(
		item.Extension.Name, item.Extension.Parent,
	)

	return file
}

func (f *PathFinder) Statics() *common.StaticInfo {
	return f.Stats
}

func (f *PathFinder) Scheme() string {
	return f.Sch
}

// Transfer creates a path for the input; should return empty
// string for the folder, if no move is required (ie non transparent).
// The FileManager will only call this function when the input
// is not transparent. When the --Trash option is present, it will
// determine the destination path for the input.
func (f *PathFinder) Transfer(info *common.PathInfo) (folder, file string) {
	folder = func() string {
		if info.IsCuddling || info.IsSampling {
			return ""
		}

		if info.Output != "" && info.Trash == "" {
			// When output folder is specified, then the results will be diverted there.
			// This means there is no need to transfer the input, unless trash has
			// also been specified.
			//
			return ""
		}

		segments := pfTemplates[pfPathInputTransferFolder]
		to := lo.TernaryF(f.Trash != "",
			func() string {
				return f.Trash
			},
			func() string {
				return info.Origin
			},
		)
		// eventually, we need to use the cuddle option here
		//
		return pfTemplates.evaluate(pfFieldValues{
			"${{TRANSFER-DESTINATION}}": to,
			"${{ITEM-SUB-PATH}}":        info.Item.Extension.SubPath,
			"${{DEJA-VU}}":              f.Stats.TrashTag(),
			"${{SUPPLEMENT}}":           f.FolderSupplement(info.Profile),
		}, segments...)
	}()

	file = func() string {
		if (info.Output != "" && info.Trash == "") || info.IsSampling {
			// When output folder is specified, then the results will be diverted there.
			// This means there is no need to transfer the input, unless trash has
			// also been specified.
			//
			return ""
		}

		if info.IsCuddling {
			supp := fmt.Sprintf("%v.%v", f.Stats.TrashTag(),
				f.FileSupplement(info.Profile, ""),
			)

			return SupplementFilename(
				info.Item.Extension.Name, supp, f.Stats,
			)
		}

		return info.Item.Extension.Name
	}()

	if filepath.Join(folder, file) == info.Item.Path {
		// Since we have derived a path that is the same as the input,
		// we should return nothing to indicate to the file manager
		// that it does not have to move/rename the input.
		//
		return "", ""
	}

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
	folder = func() string {
		return lo.TernaryF(f.transparentInput || info.IsCuddling || info.IsSampling,
			func() string {
				// The result file has to be in the same folder
				// as the input
				//
				segments := pfTemplates[pfPathTxInputDestinationFolder]

				return pfTemplates.evaluate(pfFieldValues{
					"${{OUTPUT-ROOT}}": info.Origin,
				}, segments...)
			},
			func() string {
				segments := pfTemplates[pfPathResultFolder]
				to := lo.TernaryF(f.Output != "",
					func() string {
						return f.Output
					},
					func() string {
						return info.Origin
					},
				)
				// If there is no scheme or profile, then the user is
				// only relying on flags on the command line, ie running adhoc
				// so the result path should include an adhoc label. Otherwise,
				// the result should reflect the supplementary path.
				//

				return pfTemplates.evaluate(pfFieldValues{
					"${{OUTPUT-ROOT}}":   to,
					"${{ITEM-SUB-PATH}}": info.Item.Extension.SubPath,
					"${{SUPPLEMENT}}":    f.FolderSupplement(info.Profile),
				}, segments...)
			},
		)
	}()

	file = func() string {
		// The file name just matches the input file name. The folder name
		// provides the context.
		//
		if info.IsCuddling || info.IsSampling {
			// decorate the input file to get the result file
			//
			withSampling := ""
			if info.IsSampling {
				withSampling = f.Stats.Sample
			}

			supp := lo.TernaryF(info.IsSampling,
				func() string {
					return f.SampleFileSupplement(withSampling)
				},
				func() string {
					return f.FileSupplement(info.Profile, withSampling)
				},
			)

			return SupplementFilename(
				info.Item.Extension.Name, supp, f.Stats,
			)
		}

		return info.Item.Extension.Name
	}()

	return folder, f.mutateExtension(file)
}

func (f *PathFinder) FolderSupplement(profile string) string {
	return lo.TernaryF(f.Sch == "" && profile == "",
		func() string {
			adhocLabel := f.Stats.Adhoc
			return adhocLabel
		},
		func() string {
			return filepath.Join(f.Sch, profile)
		},
	)
}

func FileSupplement(scheme, profile, adhoc, withSampling string) string {
	var (
		result string
	)

	switch {
	case scheme != "" && profile != "":
		result = fmt.Sprintf("%v.%v", scheme, profile)

	case scheme != "":
		result = scheme

	case profile != "":
		result = profile

	default:
		result = adhoc
	}

	if withSampling != "" {
		result = fmt.Sprintf("$%v$.%v", withSampling, result)
	}

	return result
}

func (f *PathFinder) FileSupplement(profile, withSampling string) string {
	return FileSupplement(f.Sch, profile, f.Stats.Adhoc, withSampling) // todo: is adhoc ok here?
}

func (f *PathFinder) SampleFileSupplement(withSampling string) string {
	return fmt.Sprintf("$%v$", withSampling)
}

func (f *PathFinder) TransparentInput() bool {
	return f.transparentInput
}

func (f *PathFinder) Observe(t common.PathFinder) common.PathFinder {
	return t
}
