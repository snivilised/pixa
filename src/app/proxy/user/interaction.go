package user

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

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
	emojis               = []rune(
		"🍦🧋🍡🤠👾😭🦊🐯🦆👽👻🍔🍒🍥😈🤮🦁🍰🐶🐸🍕🥐💀💩🥇🫐🏆🤖🌽🍉🥝🍓",
	)
)

func randemoji() string {
	return string(emojis[rand.Intn(len(emojis))]) //nolint:gosec // foo
}

type walker interface {
	navigate(ci common.ClientTraverseInfo,
		after ...common.AfterFunc,
	) (result *nav.TraverseResult, err error)
}

type interaction struct {
	inputs *common.ShrinkCommandInputs
	logger *slog.Logger
	arity  uint
}

func (u *interaction) IfWithPool(with nav.CreateNewRunnerWith, now int, cpu bool) bool {
	// this should go into nav, alongside IfWithPoolUseContext
	return with&nav.RunnerWithPool > 0 && (now >= 0 || cpu)
}

func (u *interaction) IfWithPoolUseContext(with nav.CreateNewRunnerWith, now int, cpu bool, args ...any) []any {
	return lo.TernaryF(u.IfWithPool(with, now, cpu),
		func() []any {
			return args
		},
		func() []any {
			return []any{}
		},
	)
}

func (u *interaction) navigate(ci common.ClientTraverseInfo,
	after ...common.AfterFunc,
) (result *nav.TraverseResult, err error) {
	wgan := boost.NewAnnotatedWaitGroup("🍂 traversal", u.logger)
	wgan.Add(1, navigatorRoutineName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	now := u.inputs.Root.WorkerPoolFam.Native.NoWorkers
	cpu := u.inputs.Root.WorkerPoolFam.Native.CPU
	with := ci.RunWith()
	runnerInfo := &nav.RunnerInfo{
		PrimeInfo: &nav.Prime{
			Path:      u.inputs.Root.ParamSet.Native.Directory,
			OptionsFn: ci.ActiveOptionsFn(),
		},
		ResumeInfo: ci.Resumption(),
		AccelerationInfo: &nav.Acceleration{
			WgAn:        wgan,
			RoutineName: navigatorRoutineName,
			NoW:         now,
			JobsChOut:   make(boost.JobStream[nav.TraverseItemInput], DefaultJobsChSize),
		},
	}

	result, err = nav.New().With(with, runnerInfo).Run(
		u.IfWithPoolUseContext(with, now, cpu, ctx, cancel)...,
	)

	if u.IfWithPool(with, now, cpu) {
		wgan.Wait(boost.GoRoutineName(fmt.Sprintf("👾 %v", ci.Name())))
	}

	for _, fn := range after {
		fn(result, err)
	}

	return result, err
}

func summary(result *nav.TraverseResult, err error) string {
	measure := fmt.Sprintf("started: '%v', elapsed: '%v'",
		result.Session.StartedAt().Format(time.RFC1123), result.Session.Elapsed(),
	)
	files := result.Metrics.Count(nav.MetricNoFilesInvokedEn)
	folders := result.Metrics.Count(nav.MetricNoFoldersInvokedEn)
	numbers := fmt.Sprintf("files: %v, folders: %v", files, folders)
	message := lo.Ternary(err == nil,
		fmt.Sprintf("🔊 navigation completed ok (%v) 💝 [%v]", numbers, measure),
		fmt.Sprintf("🔊 error occurred during navigation (%v)💔 [%v]", err, measure),
	)

	return message
}

func NewInteraction(inputs *common.ShrinkCommandInputs,
	logger *slog.Logger, arity uint,
) common.UserInteraction {
	return lo.TernaryF(inputs.Root.TextualFam.Native.IsNoTui,
		func() common.UserInteraction {
			return &linearUI{
				interaction: interaction{
					inputs: inputs,
					logger: logger,
					arity:  arity,
				},
			}
		},
		func() common.UserInteraction {
			return &textualUI{
				interaction: interaction{
					inputs: inputs,
					logger: logger,
					arity:  arity,
				},
				po: inputs.Root.Presentation,
			}
		},
	)
}

func clearResumeFromWith(with nav.CreateNewRunnerWith) nav.CreateNewRunnerWith {
	// ref: https://go.dev/ref/spec#Arithmetic_operators
	//
	return (with &^ nav.RunnerWithResume)
}
