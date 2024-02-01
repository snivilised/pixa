package ipc

import (
	"fmt"
	"os"

	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/cfg"
)

type fakeAgent struct {
	baseAgent
	vfs      storage.VirtualFS
	advanced cfg.AdvancedConfig
}

func (a *fakeAgent) IsInstalled() bool {
	_, err := a.program.Look()

	return err == nil
}

func (a *fakeAgent) Invoke(thirdPartyCL clif.ThirdPartyCommandLine, source, destination string) error {
	var (
		err  error
		fake *os.File
	)

	before := []string{source}

	if a.vfs.FileExists(destination) {
		return os.ErrExist
	}

	if fake, err = a.vfs.Create(destination); err != nil {
		return err
	}

	// for this to work, the dry run decorator needs to be in place ...
	//
	fmt.Printf("---> 🚀 created fake destination at '%v'\n", destination)

	defer fake.Close()

	return a.program.Execute(
		clif.Expand(before, thirdPartyCL, destination)...,
	)
}
