package ipc

import (
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/cfg"
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
	advanced cfg.AdvancedConfig,
	knownBy clif.KnownByCollection,
	vfs storage.VirtualFS,
	dryRun bool,
) (common.ExecutionAgent, error) {
	var (
		agent common.ExecutionAgent
		err   error
	)

	if dryRun {
		return Pacify(advanced, knownBy, vfs, PacifyWithFake), nil
	}

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
			err = ErrUseDummyExecutor
		}

	case "dummy":
		agent = Pacify(advanced, knownBy, vfs, PacifyWithDummy)

	case "fake":
		agent = Pacify(advanced, knownBy, vfs, PacifyWithFake)

	default:
		err = ErrUnsupportedExecutor
	}

	return agent, err
}

func Pacify(
	advanced cfg.AdvancedConfig,
	knownBy clif.KnownByCollection,
	vfs storage.VirtualFS,
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
		vfs: vfs,
	}
}
