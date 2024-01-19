package proxy

import "github.com/snivilised/cobrass/src/clif"

type magickAgent struct {
	baseAgent
}

func (a *magickAgent) IsInstalled() bool {
	_, err := a.program.Look()

	return err == nil
}

func (a *magickAgent) Invoke(thirdPartyCL clif.ThirdPartyCommandLine, source, destination string) error {
	before := []string{source}

	return a.program.Execute(
		clif.Expand(before, thirdPartyCL, destination)...,
	)
}
