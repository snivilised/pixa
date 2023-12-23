package proxy

import (
	"path/filepath"

	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/extendio/xfs/nav"
)

// Step
type (
	RunStepInfo struct {
		Item   *nav.TraverseItem
		Source string
	}

	Step interface {
		Run(rsi *RunStepInfo) error
	}

	// Sequence
	Sequence []Step
)

// magickStep knows how to combine parameters together so that the program
// can be invoked correctly; but it does not know how to compose the input
// and output file names; this is the responsibility of the controller, which uses
// the path-finder to accomplish that task.
type magickStep struct {
	shared       *SharedControllerInfo
	thirdPartyCL clif.ThirdPartyCommandLine
	scheme       string
	profile      string
	sourcePath   string
	outputPath   string
	journalPath  string
}

// Run
func (s *magickStep) Run(rsi *RunStepInfo) error {
	folder, file := s.shared.finder.Result(&resultInfo{
		pathInfo: pathInfo{
			item:   rsi.Item,
			origin: rsi.Item.Extension.Parent,
		},
		scheme:  s.scheme,
		profile: s.profile,
	})
	result := filepath.Join(folder, file)
	input := []string{rsi.Source}

	// if transparent, then we need to ask the fm to move the
	// existing file out of the way. But shouldn't that already have happened
	// during setup? See, which mean setup in not working properly in
	// this scenario.

	return s.shared.program.Execute(
		clif.Expand(input, s.thirdPartyCL, result)...,
	)
}
