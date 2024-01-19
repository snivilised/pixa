package proxy

import (
	"path/filepath"

	"github.com/snivilised/cobrass/src/clif"
)

// Step
type (
	RunStepInfo struct {
		Source string
	}

	Step interface {
		Run(pi *pathInfo) error
	}

	// Sequence
	Sequence []Step
)

// controllerStep uses the agent to combine parameters together so that the program
// can be invoked correctly; but it does not know how to compose the input and
// output file names; this is the responsibility of the controller, which uses
// the path-finder to accomplish that task.
type controllerStep struct {
	shared       *SharedControllerInfo
	thirdPartyCL clif.ThirdPartyCommandLine
	profile      string
	sourcePath   string
	outputPath   string
	journalPath  string
}

// Run
func (s *controllerStep) Run(pi *pathInfo) error {
	folder, file := s.shared.finder.Result(pi)
	destination := filepath.Join(folder, file)

	// if transparent, then we need to ask the fm to move the
	// existing file out of the way. But shouldn't that already have happened
	// during setup? See, which mean setup in not working properly in
	// this scenario.

	return s.shared.agent.Invoke(
		s.thirdPartyCL, pi.runStep.Source, destination,
	)
}
