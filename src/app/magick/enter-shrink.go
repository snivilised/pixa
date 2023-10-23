package magick

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/extendio/xfs/nav"
)

type ShrinkEntry struct {
	EntryBase
	Inputs *ShrinkCommandInputs
	jobs   []string
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
			return e.Program.Execute(clif.Expand(positional, e.ThirdPartyCL)...)
		},
	}
	o.Store.Subscription = nav.SubscribeFiles
	o.Store.DoExtend = true

	e.EntryBase.ConfigureOptions(o)
}

func (e *ShrinkEntry) run(config configuration.ViperConfig) error {
	_ = config

	e.ThirdPartyCL = cobrass.Evaluate(
		e.Inputs.ParamSet.Native.ThirdPartySet.Present,
		e.Inputs.ParamSet.Native.ThirdPartySet.KnownBy,
		e.readProfile3rdPartyFlags(),
	)

	runnerWith := composeWith(e.Inputs.RootInputs.ParamSet)
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

					return e.Program.Execute(clif.Expand(positional, e.ThirdPartyCL)...)
				},
			}
		},
		Strategy: nav.ResumeStrategySpawnEn, // to come from an arg
	}

	return e.navigate(GetTraverseOptionsFunc(e), runnerWith, resumption)
}

func EnterShrink(
	inputs *ShrinkCommandInputs,
	program Executor,
	config configuration.ViperConfig,
) error {
	fmt.Printf("---> 🦠🦠🦠 Directory: '%v'\n", inputs.RootInputs.ParamSet.Native.Directory)

	entry := &ShrinkEntry{
		EntryBase: EntryBase{
			Inputs:  inputs.RootInputs,
			Program: program,
			Config:  config,
		},
		Inputs: inputs,
	}

	return entry.run(config)
}
