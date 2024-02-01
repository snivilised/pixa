package ipc

import (
	"fmt"

	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type fakeAgent struct {
	baseAgent
	fm       common.FileManager
	advanced common.AdvancedConfig
}

func (a *fakeAgent) IsInstalled() bool {
	_, err := a.program.Look()

	return err == nil
}

func (a *fakeAgent) Invoke(thirdPartyCL clif.ThirdPartyCommandLine,
	source, destination string,
) error {
	before := []string{source}

	if err := a.fm.Create(destination, false); err != nil {
		return err
	}

	// for this to work, the dry run decorator needs to be in place ...
	//
	fmt.Printf("---> 🚀 created fake destination at '%v'\n", destination)

	return a.program.Execute(
		clif.Expand(before, thirdPartyCL, destination)...,
	)
}
