package proxy

import (
	"fmt"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/lorax/boost"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type configProfile struct {
	args []string
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

func (e *RootEntry) principalFn(item *nav.TraverseItem) error {
	depth := item.Extension.Depth
	indicator := lo.Ternary(len(item.Children) > 0, "🔆", "🌊")

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

	return nil
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
		Fn:    e.principalFn,
	}
	o.Store.Subscription = nav.SubscribeFoldersWithFiles
	e.EntryBase.ConfigureOptions(o)
}

func (e *RootEntry) run() error {
	runnerWith := composeWith(e.Inputs)

	// root does not need to support resume
	//
	var nilResumption *nav.Resumption

	after := func(result *nav.TraverseResult, err error) {
		for _, file := range e.files {
			fmt.Printf("		===> 📒 candidate file: '%v'\n", file)
		}
	}

	principal := func(o *nav.TraverseOptions) {
		e.ConfigureOptions(o)
		o.Callback = &nav.LabelledTraverseCallback{
			Label: "Root Entry Callback",
			Fn:    e.principalFn,
		}
	}

	return e.navigateLegacy(
		principal,
		runnerWith,
		nilResumption,
		after,
		summariseAfter,
	)
}

func composeWith(inputs *common.RootCommandInputs) nav.CreateNewRunnerWith {
	with := nav.RunnerDefault

	if inputs.WorkerPoolFam.Native.CPU || inputs.WorkerPoolFam.Native.NoWorkers >= 0 {
		with |= nav.RunnerWithPool
	}

	return with
}

func EnterRoot(
	inputs *common.RootCommandInputs,
	config configuration.ViperConfig,
) error {
	fmt.Printf("---> 📁📁📁 Directory: '%v'\n", inputs.ParamSet.Native.Directory)

	entry := &RootEntry{
		EntryBase: EntryBase{
			Inputs: inputs,
			Viper:  config,
		},
	}

	return entry.run()
}
