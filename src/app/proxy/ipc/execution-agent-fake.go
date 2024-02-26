package ipc

import (
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type fakeAgent struct {
	baseAgent
	fm       common.FileManager
	advanced common.AdvancedConfig
}

func (a *fakeAgent) IsInstalled() bool {
	return true
}

func (a *fakeAgent) Invoke(thirdPartyCL clif.ThirdPartyCommandLine,
	source, destination string,
) error {
	before := []string{source}

	// >>> if err := a.fm.Create(destination, false); err != nil {
	// 	return err
	// }

	return a.program.Execute(
		clif.Expand(before, thirdPartyCL, destination)...,
	)
}
