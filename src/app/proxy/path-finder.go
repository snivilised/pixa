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
	pfPathSetupInlineDestFolder
	pfPathSetupInlineDestFileOriginalExt
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

func init() {
	pfTemplates = pfTemplatesCollection{
		// we probably have to come up with better key names...
		//
		pfPathSetupInlineDestFolder: templateSegments{
			"${{OUTPUT-ROOT}}",
			"${{ITEM-SUB-PATH}}",
			"${{TRASH-LABEL}}",
		},

		pfPathSetupInlineDestFileOriginalExt: templateSegments{
			"${{ITEM-NAME-ORIG-EXT}}",
		},
	}
}

// expand returns a string as a result of joining the segments
func (tc pfTemplatesCollection) expand(segments ...string) string {
	return filepath.Join(segments...)
}

// evaluate returns a string representing a file system path from a
// template string containing place-holders and field values
func (tc pfTemplatesCollection) evaluate(
	sourceTemplate string,
	placeHolders templateSegments,
	values pfFieldValues,
) string {
	const (
		quantity = 1
	)

	result := lo.Reduce(placeHolders, func(acc, field string, _ int) string {
		return strings.Replace(acc, field, values[field], quantity)
	},
		sourceTemplate,
	)

	return filepath.Clean(result)
}

// INLINE-MODE: EJECT | INLINE (should we call this a strategy?
// they do the same thing but create a different output structure => OutputStrategy)
//
// EJECT: replicate the source directory struct but eject elsewhere
// INLINE: create the file at the same location as the original but rename as required

// The controller is aware of the output strategy moving files accordingly,
// using the path-finder to create the paths and the file-manager to interact
// with the file system, using a vfs.

// Then we can also have a deletion strategy, use a central location or inline
// CENTRAL-LOCATION: is like EJECT
// INLINE: INLINE

// Can we have 2 different strategies at the same time?, ie:
// OUTPUT-STRATEGY: INLINE
// DELETION-STRATEGY: EJECT
//
// ... well in this case, the output file would be in the same folder
// as item.Path, but the TRASH folder would be relative to eject-path (ie in
// the same folder as item.Path) and the

/* eject parameters:

we can't have --eject, because ambiguous, which strategy does this apply to?
(but what we could say is --eject if specified applies to output && deletion)

- the same goes for inline, but --inline would be a switch, not a flag
==> --eject(path) & --inline [still needs a way to specify how to manage renames]
both strategies set to eject or inline(~transparent mode)
if we say this, then --inline could be redundant, ie if --eject is not set,
then we revert to the default which is eject(transparent)

-- then other flags could adjust the transparent mode
if --eject not specified, then ous=inline; des=inline
but it can be adjusted by --output <path>, --trash <path>


-- perhaps we have a transparency mode, ie perform renames such that the new generated
files adopt the existing files, so there is no difference, except for the original
file would be renamed to something else. With transparency enabled, we make all the
decisions to make this possible, we internally make the choice of which strategies
are in place, so the user doesn't have to work this out for themselves. But the
deletion strategy is independent of transparency, so it really only applies to output.
Or, perhaps, do we assume transparent by default and the other options adjust this.

so we have 3 parameters:
* neither --output or --trash specified					[ous=eject; des=eject];
* --output <path>																[ous=eject; des=inline]
* --trash <path>    														[ous=inline; des=eject]
* --output <path> --trash <path>								[ous=eject; des=eject]
*/

// PathFinder provides the common paths required, but its the controller that know
// the specific paths based around this common framework

type strategies struct {
	output   outputStrategy
	deletion deletionStrategy
}

type PathFinder struct {
	Scheme  string
	Profile string
	// Origin is the parent of the item (item.Parent)
	//
	Origin string

	// only the step knows this, so this should be the parent of the output
	// for scheme, this would include scheme/profile
	// for profile, this should include profile
	// only the step knows this, so this should be the parent
	// the associated getter method (maybe GetOutput()) should accept a argument
	//  that denotes intermediate segments, eg "<scheme>/<profile>",
	// perhaps represented as a slice so it can be joined with filepath.Join
	//
	// if Output Path is set, then use this as the output, but also
	// create the intermediate paths in order to implement mirroring.
	// It is the output as indicated by --output. If not set, then it is
	// derived:
	// - sampling: (inline) -- item.parent; => item.parent/SHRINK/<supplement>
	// - full: (inline) -- item.parent
	Output string

	// I think this depends on the mode (tidy/preserve)
	Trash string

	behaviours strategies
}

type staticInfo struct {
	trashLabel  string
	legacyLabel string
}

type destinationInfo struct {
	item   *nav.TraverseItem
	origin string // in:item.Parent.Path, ej:eject-path(output???)
	// statics     *staticInfo
	transparent bool
	//
	// transparent=true should be the default scenario. This means
	// that any changes that occur leave the file system in a state
	// where nothing appears to have changed except that files have
	// been modified, without name changes. This of course doesn't
	// include items that end up in TRASH and can be manually deleted
	// by the user. The purpose of this is to by default require
	// the least amount of post-processing clean-up from the user.
	//
	// In sampling mode, transparent may mean something different
	// because multiple files could be created for each input file.
	// So, in this scenario, the original file should stay in-tact
	// and the result(s) should be created into the supplementary
	// location.
	//
	// In full  mode, transparent means the input file is moved
	// to a trash location. The output takes the name of the original
	// file, so that by the end of processing, the resultant file
	// takes the place of the source file, leaving the file system
	// in a state that was the same before processing occurred.
	//
	// So what happens in non transparent scenario? The source file
	// remains unchanged, so the user has to look at another location
	// to get the result. It uses the SHRINK label to create the
	// output filename; but note, we only use the SHRINK label in
	// scenarios where there is a potential for a filename clash if
	// the output file is in the same location as the input file
	// because we want to create the least amount of friction as
	// possible. This only occurs when in adhoc mode (no profile
	// or scheme)
}

// Destination returns the location of what should be used
// for the specified source path; ie when the program runs, it uses
// a source file and requires the destination location. The source
// and destination may not be n the same folder, so the source's name
// is extracted from the source path and attached to the output
// folder.
//
// should return empty string if no move is required
func (f *PathFinder) Destination(info *destinationInfo) (destinationFolder, destinationFile string) {
	// TODO: we still need to get the rest of the mirror sub-path
	// ./<item.Parent>/<MIRROR-SUB-PATH>/TRASH/<scheme>/<profile>/<.item.Name>.<LEGACY>.ext
	// legacyLabel := "LEGACY"
	trashLabel := "TRASH"

	// this does not take into account transparent, without modification;
	// ie what happens if we don;t want any supplemented paths?

	to := lo.TernaryF(f.Output != "",
		func() string {
			return f.Output // eject
		},
		func() string {
			return info.origin // inline
		},
	)

	destinationFolder = func() string {
		segments := pfTemplates[pfPathSetupInlineDestFolder]
		path := pfTemplates.expand(filepath.Join(segments...))

		return pfTemplates.evaluate(path, segments, pfFieldValues{
			"${{OUTPUT-ROOT}}":   to,
			"${{ITEM-SUB-PATH}}": info.item.Extension.SubPath,
			"${{TRASH-LABEL}}":   trashLabel,
		})
	}()

	destinationFile = func() string {
		segments := pfTemplates[pfPathSetupInlineDestFileOriginalExt]
		path := pfTemplates.expand(filepath.Join(segments...))

		return pfTemplates.evaluate(path, segments, pfFieldValues{
			"${{ITEM-NAME-ORIG-EXT}}": info.item.Extension.Name,
		})
	}()

	return destinationFolder, destinationFile
}

/*
mode: tidy | preserve

*** item-handler

contains:
- the Program
- positional args
- third party CL
- path-manager

*** path-manager needs to provide the following paths
- output directory (inline | eject)
- trash file location (central or local)
- for sampling scheme, we use profile name as part of the relative output path
- output root (depends on eject)

behaviour of naming output files:
- output filename (same as input file with a suffix | backup input, replace input file)

- this is where the extension mapper will be implemented


*** file-manager: perform file system operations such as moving files around
- there will be a file-system service that can perform fs operations. it
will contain the path-finder
- contains file manager as a member
- is populated with the current traversal item


the sample mode is a bit tricky because for 1 file it will do multiple things

we need to capture that concept somehow

- foreach incoming file
FULL: => a single output file
SAMPLE: => multiple files, foreach each profile in the sample create an output
*/
