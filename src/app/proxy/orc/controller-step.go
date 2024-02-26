package orc

import (
	"fmt"
	"path/filepath"

	"github.com/samber/lo"
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
	finder := s.session.FileManager.Finder()
	folder, file := finder.Result(pi)
	destination := filepath.Join(folder, file)

	err := lo.TernaryF(s.session.FileManager.FileExists(destination),
		func() error {
			return fmt.Errorf("skipping file: '%v'", destination)
		},
		func() error {
			// todo: if sample file exists, rename it to the destination,
			// then skip the invoke
			//
			destination = filepath.Join(folder, file)

			if s.session.FileManager.FileExists(destination) {
				// todo: rename the sample
				//
				return fmt.Errorf("skipping existing sample file: '%v'", destination)
			}

			return s.session.Agent.Invoke(
				s.thirdPartyCL, pi.RunStep.Source, destination,
			)
		},
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
