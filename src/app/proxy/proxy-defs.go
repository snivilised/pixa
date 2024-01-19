package proxy

import (
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/cfg"
)

type SharedControllerInfo struct {
	agent       ExecutionAgent
	profiles    cfg.ProfilesConfig
	schemes     cfg.SchemesConfig
	sampler     cfg.SamplerConfig
	advanced    cfg.AdvancedConfig
	Inputs      *ShrinkCommandInputs
	finder      *PathFinder
	fileManager *FileManager
}

type privateControllerInfo struct {
	destination string
	pi          pathInfo
}

// ItemController
type ItemController interface {
	OnNewShrinkItem(item *nav.TraverseItem) error
	Reset()
}

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

type pathInfo struct {
	item    *nav.TraverseItem
	origin  string
	scheme  string
	profile string
	runStep RunStepInfo
}
