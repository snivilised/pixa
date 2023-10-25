package proxy

import (
	"fmt"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/lorax/boost"
)

type configProfile struct {
	args []string `mapstructure:"string_slice"`
}

type Executor interface {
	ProgName() string
	Look() (string, error)
	Execute(args ...string) error
}

const (
	DefaultJobsChSize = 10
)

var (
	navigatorRoutineName = boost.GoRoutineName("✨ pixa-navigator")
)

type RootEntry struct {
	EntryBase

	files []string
}

func (e *RootEntry) ConfigureOptions(o *nav.TraverseOptions) {
	o.Notify.OnBegin = func(_ *nav.NavigationState) {
		fmt.Printf("===> 🛡️ beginning traversal ...\n")
	}
	o.Notify.OnEnd = func(result *nav.TraverseResult) {
		fmt.Printf("===> 🚩 finished traversal - folders '%v'\n",
			result.Metrics.Count(nav.MetricNoFoldersInvokedEn),
		)
	}
	o.Callback = &nav.LabelledTraverseCallback{
		Label: "Root Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			depth := item.Extension.Depth
			indicator := lo.Ternary(len(item.Children) > 0, "☀️", "🌊")

			for _, entry := range item.Children {
				fullPath := filepath.Join(item.Path, entry.Name())
				e.files = append(e.files, fullPath)
			}

			fmt.Printf(
				"---> %v ROOT-CALLBACK: (depth:%v, files:%v) '%v'\n",
				indicator,
				depth, len(item.Children),
				item.Path,
			)

			return e.Program.Execute("--version")
		},
	}
	o.Store.Subscription = nav.SubscribeFoldersWithFiles
	o.Store.DoExtend = true
	e.EntryBase.ConfigureOptions(o)
}

func (e *RootEntry) run() error {
	files := []string{}
	runnerWith := composeWith(e.Inputs.ParamSet)
	resumption := &nav.Resumption{
		RestorePath: "/json-path-to-come-from-a-flag-option/restore.json",
		Restorer: func(o *nav.TraverseOptions, active *nav.ActiveState) {
			o.Callback = &nav.LabelledTraverseCallback{
				Label: "Root Entry Callback",
				Fn: func(item *nav.TraverseItem) error {
					depth := item.Extension.Depth
					indicator := lo.Ternary(len(item.Children) > 0, "☀️", "🌊")

					for _, entry := range item.Children {
						fullPath := filepath.Join(item.Path, entry.Name())
						files = append(files, fullPath)
					}

					fmt.Printf(
						"---> %v ROOT-CALLBACK: (depth:%v, files:%v) '%v'\n",
						indicator,
						depth, len(item.Children),
						item.Path,
					)

					return nil
				},
			}
		},
		Strategy: nav.ResumeStrategySpawnEn, // to come from an arg
	}

	after := func(result *nav.TraverseResult, err error) {
		for _, file := range e.files {
			fmt.Printf("		===> 🔆 candidate file: '%v'\n", file)
		}
	}

	return e.navigate(GetTraverseOptionsFunc(e), runnerWith, resumption, after)
}

func composeWith(rps *assistant.ParamSet[RootParameterSet]) nav.CreateNewRunnerWith {
	with := nav.RunnerDefault

	if rps.Native.CPU || rps.Native.NoW >= 0 {
		with |= nav.RunnerWithPool
	}

	return with
}

func EnterRoot(
	inputs *RootCommandInputs,
	program Executor,
	config configuration.ViperConfig,
) error {
	fmt.Printf("---> 🦠🦠🦠 Directory: '%v'\n", inputs.ParamSet.Native.Directory)

	entry := &RootEntry{
		EntryBase: EntryBase{
			Inputs:  inputs,
			Program: program,
			Config:  config,
		},
	}

	return entry.run()
}
