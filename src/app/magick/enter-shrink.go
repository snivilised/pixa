package magick

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/nav"
)

type ShrinkEntry struct {
	EntryBase

	ParamSet *assistant.ParamSet[ShrinkParameterSet]
	jobs     []string
}

func (e *ShrinkEntry) ConfigureOptions(o *nav.TraverseOptions) {
	o.Notify.OnBegin = func(_ *nav.NavigationState) {
		fmt.Printf("===> 🛡️ beginning traversal ...\n")
	}
	o.Notify.OnEnd = func(result *nav.TraverseResult) {
		fmt.Printf("===> 🚩 finished traversal - folders '%v'\n",
			result.Metrics.Count(nav.MetricNoFoldersInvokedEn),
		)
	}
	o.Callback = &nav.LabelledTraverseCallback{
		Label: "Shrink Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			depth := item.Extension.Depth

			fmt.Printf(
				"---> 📜 SHRINK-CALLBACK-FILE: (depth:%v) '%v'\n",
				depth,
				item.Path,
			)

			positional := []string{
				fmt.Sprintf("'%v'", item.Path),
			}
			return e.Program.Execute(e.expand(positional)...)
		},
	}
	o.Store.Subscription = nav.SubscribeFiles
	o.Store.DoExtend = true

	e.EntryBase.ConfigureOptions(o)
}

func (e *ShrinkEntry) run(config configuration.ViperConfig) error {
	_ = config

	runnerWith := composeWith(e.RootPS)
	resumption := &nav.Resumption{
		RestorePath: "/json-path-to-come-from-a-flag-option/restore.json",
		Restorer: func(o *nav.TraverseOptions, active *nav.ActiveState) {
			o.Callback = &nav.LabelledTraverseCallback{
				Label: "Shrink Entry Callback",
				Fn: func(item *nav.TraverseItem) error {
					depth := item.Extension.Depth
					indicator := lo.Ternary(len(item.Children) > 0, "☀️", "🌊")

					fmt.Printf(
						"---> %v SHRINK-RESTORE-CALLBACK: (depth:%v) '%v'\n",
						indicator,
						depth,
						item.Path,
					)

					positional := []string{
						fmt.Sprintf("'%v'", item.Path),
					}
					return e.Program.Execute(e.expand(positional)...)
				},
			}
		},
		Strategy: nav.ResumeStrategySpawnEn, // to come from an arg
	}

	return e.navigate(GetTraverseOptionsFunc(e), runnerWith, resumption)
}

func EnterShrink(
	rps *assistant.ParamSet[RootParameterSet],
	ps *assistant.ParamSet[ShrinkParameterSet],
	program Executor,
	config configuration.ViperConfig,
) error {
	fmt.Printf("---> 🦠🦠🦠 Directory: '%v'\n", rps.Native.Directory)

	entry := &ShrinkEntry{
		EntryBase: EntryBase{
			RootPS:  rps,
			Program: program,
			Config:  config,
		},
		ParamSet: ps,
	}
	entry.evaluate()

	return entry.run(config)
}
