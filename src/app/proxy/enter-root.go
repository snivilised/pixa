package proxy

import (
	"log/slog"
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

	// this need to be done properly, in the mean time just populate the log
	// (this should probably be implemented inside the interaction ui)
	//
	e.Log.Info("---> %v ROOT-CALLBACK",
		slog.String(">", indicator),
		slog.Int("depth", depth),
		slog.Int("files", len(item.Children)),
		slog.String("name", item.Extension.Name),
		slog.String("sub-path", item.Extension.SubPath),
	)

	return nil
}

func (e *RootEntry) ConfigureOptions(o *nav.TraverseOptions) {
	o.Notify.OnBegin = func(_ *nav.NavigationState) {
		e.Log.Info("===> 🛡️  beginning traversal ...")
	}
	o.Notify.OnEnd = func(result *nav.TraverseResult) {
		e.Log.Info("===> 🚩 finished traversal - folders",
			slog.Int("folders", int(result.Metrics.Count(nav.MetricNoFoldersInvokedEn))), //nolint:gosec // ok
		)
	}
	o.Callback = &nav.LabelledTraverseCallback{
		Label: "Root Entry Callback",
		Fn:    e.principalFn,
	}
	o.Store.Subscription = nav.SubscribeFoldersWithFiles
	e.EntryBase.ConfigureOptions(o)
}

func (e *RootEntry) run() (*nav.TraverseResult, error) {
	runnerWith := composeWith(e.Inputs)

	// root does not need to support resume
	//
	var nilResumption *nav.Resumption

	after := func(_ *nav.TraverseResult, _ error) {
		for _, file := range e.files {
			e.Log.Info("📒 candidate file: '%v'",
				slog.String("file", file),
			)
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
		// --> summariseAfter,
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
	logger *slog.Logger,
) (*nav.TraverseResult, error) {
	logger.Info("---> 📁📁📁 Directory: '%v'",
		slog.String("directory", inputs.ParamSet.Native.Directory),
	)

	entry := &RootEntry{
		EntryBase: EntryBase{
			Inputs: inputs,
			Viper:  config,
			Log:    logger,
		},
	}

	return entry.run()
}
