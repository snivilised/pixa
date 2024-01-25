package ipc

import (
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/cfg"
)

type baseAgent struct {
	knownBy clif.KnownByCollection
	program common.Executor
}

func New(
	advanced cfg.AdvancedConfig,
	knownBy clif.KnownByCollection,
) (common.ExecutionAgent, error) {
	var (
		agent common.ExecutionAgent
		// dummy uses the same agent as magick
		//
		dummy = &magickAgent{
			baseAgent{
				knownBy: knownBy,
				program: &DummyExecutor{
					Name: advanced.Executable().Symbol(),
				},
			},
		}
		err error
	)

	switch advanced.Executable().Symbol() {
	case "magick":
		agent = &magickAgent{
			baseAgent{
				knownBy: knownBy,
				program: &ProgramExecutor{
					Name: advanced.Executable().Symbol(),
				},
			},
		}

		if !agent.IsInstalled() {
			err = ErrUsingDummyExecutor

			agent = dummy
		}

	case "dummy":
		err = ErrUsingDummyExecutor

		agent = dummy
	}

	return agent, err
}
