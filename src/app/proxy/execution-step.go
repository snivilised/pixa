package proxy

import (
	"github.com/snivilised/cobrass/src/clif"
)

// Step
type Step interface {
	Run(*SharedControllerInfo) error
}

// Sequence
type Sequence []Step

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
func (s *magickStep) Run(*SharedControllerInfo) error {
	positional := []string{s.sourcePath}

	return s.shared.program.Execute(clif.Expand(positional, s.thirdPartyCL, s.outputPath)...)
}
