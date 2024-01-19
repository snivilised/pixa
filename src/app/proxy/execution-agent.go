package proxy

import (
	"fmt"

	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/cfg"
)

type baseAgent struct {
	knownBy clif.KnownByCollection
	program Executor
}

func newAgent(
	advanced cfg.AdvancedConfig,
	knownBy clif.KnownByCollection,
) ExecutionAgent {
	var (
		agent ExecutionAgent
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
			fmt.Printf("===> 💥💥💥 REVERTING TO DUMMY EXECUTOR !!!!\n")

			agent = dummy
		}

	case "dummy":
		fmt.Printf("===> 🚫 USING DUMMY EXECUTOR !!!!\n")

		agent = dummy
	}

	return agent
}
