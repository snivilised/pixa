package user

import (
	"context"
	"log/slog"

	"github.com/samber/lo"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/lorax/boost"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

const (
	DefaultJobsChSize = 10
)

var (
	navigatorRoutineName = boost.GoRoutineName("✨ pixa-navigator")
)

type baseUser struct {
	inputs *common.ShrinkCommandInputs
	logger *slog.Logger
}

func (u *baseUser) navigate(
	optionsFn nav.TraverseOptionFn,
	with nav.CreateNewRunnerWith,
	resumption *nav.Resumption,
	after ...common.AfterFunc,
) error {
	wgan := boost.NewAnnotatedWaitGroup("🍂 traversal", u.logger)
	wgan.Add(1, navigatorRoutineName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runnerInfo := &nav.RunnerInfo{
		PrimeInfo: &nav.Prime{
			Path:      u.inputs.Root.ParamSet.Native.Directory,
			OptionsFn: optionsFn,
		},
		ResumeInfo: resumption,
		AccelerationInfo: &nav.Acceleration{
			WgAn:        wgan,
			RoutineName: navigatorRoutineName,
			NoW:         u.inputs.Root.WorkerPoolFam.Native.NoWorkers,
			JobsChOut:   make(boost.JobStream[nav.TraverseItemInput], DefaultJobsChSize),
		},
	}

	result, err := nav.New().With(with, runnerInfo).Run(
		nav.IfWithPoolUseContext(with, ctx, cancel)...,
	)

	for _, fn := range after {
		fn(result, err)
	}

	return err
}

func New(inputs *common.ShrinkCommandInputs, logger *slog.Logger) common.UserInteraction {
	return lo.TernaryF(inputs.Root.TextualFam.Native.IsNoTui,
		func() common.UserInteraction {
			return &LinearUser{
				baseUser: baseUser{
					inputs: inputs,
					logger: logger,
				},
			}
		},
		func() common.UserInteraction {
			return &TextualUser{
				baseUser: baseUser{
					inputs: inputs,
					logger: logger,
				},
			}
		},
	)
}
