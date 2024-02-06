package ipc

import (
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

const (
	PacifyWithDummy = true
	PacifyWithFake  = false
)

type baseAgent struct {
	knownBy clif.KnownByCollection
	program common.Executor
}

func New(
	advanced common.AdvancedConfig,
	knownBy clif.KnownByCollection,
	fm common.FileManager,
	dryRun bool,
) (common.ExecutionAgent, error) {
	var (
		agent common.ExecutionAgent
		err   error
	)

	if dryRun {
		return Pacify(advanced, knownBy, fm, PacifyWithFake), nil
	}

	switch advanced.Executable().Symbol() {
	case common.Definitions.ThirdParty.Magick:
		agent = &magickAgent{
			baseAgent{
				knownBy: knownBy,
				program: &ProgramExecutor{
					Name: advanced.Executable().Symbol(),
				},
			},
		}

		if !agent.IsInstalled() {
			err = ErrUseDummyExecutor
		}

	case common.Definitions.ThirdParty.Dummy:
		agent = Pacify(advanced, knownBy, fm, PacifyWithDummy)

	case common.Definitions.ThirdParty.Fake:
		agent = Pacify(advanced, knownBy, fm, PacifyWithFake)

	default:
		err = ErrUnsupportedExecutor
	}

	return agent, err
}

func Pacify(
	advanced common.AdvancedConfig,
	knownBy clif.KnownByCollection,
	fm common.FileManager,
	dummy bool,
) common.ExecutionAgent {
	if dummy {
		return &magickAgent{
			baseAgent{
				knownBy: knownBy,
				program: &ProgramExecutor{
					Name: advanced.Executable().Symbol(),
				},
			},
		}
	}

	return &fakeAgent{
		baseAgent: baseAgent{
			knownBy: knownBy,
			program: &ProgramExecutor{
				Name: advanced.Executable().Symbol(),
			},
		},
		fm: fm,
	}
}
