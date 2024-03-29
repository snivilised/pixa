package orc

import (
	"path/filepath"

	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

// controllerStep uses the agent to combine parameters together so that the program
// can be invoked correctly; but it does not know how to compose the input and
// output file names; this is the responsibility of the controller, which uses
// the path-finder to accomplish that task.
type controllerStep struct {
	session      *common.SessionControllerInfo
	thirdPartyCL clif.ThirdPartyCommandLine
	profile      string
	sourcePath   string
	outputPath   string
	journalPath  string
}

// Run
func (s *controllerStep) Run(pi *common.PathInfo) error {
	pi.Profile = s.profile
	folder, file := s.session.FileManager.Finder().Result(pi)
	destination := filepath.Join(folder, file)
	err := s.session.Agent.Invoke(
		s.thirdPartyCL, pi.RunStep.Source, destination,
	)

	s.session.Interaction.Tick(&common.ProgressMsg{
		Source:      pi.RunStep.Source,
		Destination: destination,
		Scheme:      pi.Scheme,
		Profile:     s.profile,
		Err:         err,
	})

	return err
}
