package proxy

import (
	"fmt"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/pkg/errors"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/app/proxy/filing"
	"github.com/snivilised/pixa/src/app/proxy/ipc"
	"github.com/snivilised/pixa/src/app/proxy/orc"
)

type ShrinkEntry struct {
	EntryBase
	Inputs *common.ShrinkCommandInputs
}

func (e *ShrinkEntry) LookAheadOptionsFn(o *nav.TraverseOptions) {
	e.ConfigureOptions(o)
	o.Callback = &nav.LabelledTraverseCallback{
		Label: "LookAhead: Shrink Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			if strings.Contains(item.Path, common.DejaVu) {
				return fs.SkipDir
			}
			journal := e.FileManager.Finder().JournalFullPath(item)

			return e.FileManager.Create(journal)
		},
	}
}

func (e *ShrinkEntry) PrincipalOptionsFn(o *nav.TraverseOptions) {
	e.ConfigureOptions(o)

	o.Callback = &nav.LabelledTraverseCallback{
		Label: "Principal: Shrink Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			if strings.Contains(item.Path, common.DejaVu) {
				return fs.SkipDir
			}

			depth := item.Extension.Depth

			e.Log.Debug("🌀🌀 Shrink Principle Callback",
				slog.String("path", item.Extension.SubPath),
				slog.Int("depth", depth),
			)

			controller := e.Registry.Get()
			defer e.Registry.Put(controller)

			return controller.OnNewShrinkItem(item)
		},
	}
	o.Notify.OnBegin = func(_ *nav.NavigationState) {
		fmt.Printf("===> 🛡️  beginning traversal ...\n")
	}
}

func (e *ShrinkEntry) ConfigureOptions(o *nav.TraverseOptions) {
	e.EntryBase.ConfigureOptions(o)

	o.Notify.OnEnd = func(result *nav.TraverseResult) {
		e.Log.Info("finished traversal",
			slog.Int("files", int(result.Metrics.Count(nav.MetricNoFilesInvokedEn))),
			slog.Int("folders", int(result.Metrics.Count(nav.MetricNoFoldersInvokedEn))),
		)
	}
	o.Store.Subscription = nav.SubscribeFiles
	o.Store.FilterDefs = e.FilterSetup.getDefs(e.FileManager.Finder().Statics())
}

func clearResumeFromWith(with nav.CreateNewRunnerWith) nav.CreateNewRunnerWith {
	// ref: https://go.dev/ref/spec#Arithmetic_operators
	//
	return (with &^ nav.RunnerWithResume)
}

func (e *ShrinkEntry) navigateWithLookAhead(
	with nav.CreateNewRunnerWith,
	resumption *nav.Resumption,
	after ...afterFunc,
) error {
	var nilResumption *nav.Resumption

	if err := e.navigate(
		e.LookAheadOptionsFn,
		clearResumeFromWith(with),
		nilResumption,
	); err != nil {
		return errors.Wrap(err, "shrink look-ahead phase failed")
	}

	return e.navigate(
		e.PrincipalOptionsFn,
		with,
		resumption,
		after...,
	)
}

func (e *ShrinkEntry) resumeFn(item *nav.TraverseItem) error {
	if strings.HasPrefix(item.Extension.Name, common.DejaVu) {
		return fs.SkipDir
	}

	depth := item.Extension.Depth

	e.Log.Debug("🎙️🎙️ Shrink Restore Callback",
		slog.String("path", item.Extension.SubPath),
		slog.Int("depth", depth),
	)

	controller := e.Registry.Get()
	defer e.Registry.Put(controller)

	return controller.OnNewShrinkItem(item)
}

func (e *ShrinkEntry) run(_ configuration.ViperConfig) error {
	runnerWith := composeWith(e.Inputs.Root)
	resumption := &nav.Resumption{
		// actually, we need to come up with a convenient way for the restore
		// file to be found. Let's assume we declare a specific location for
		// resume files, eg ~/.pixa/resume/resumeAt.<timestamp>.json
		// this way, we can make it easy for the user and potentially
		// an auto-select feature, assuming only a single file was found.
		// If there are multiple found, then display a menu from which
		// the user can select.
		//
		RestorePath: "/json-path-to-come-from-a-flag-option/restore.json",
		Restorer: func(o *nav.TraverseOptions, active *nav.ActiveState) {
			o.Callback = &nav.LabelledTraverseCallback{
				Label: "Resume Shrink Entry Callback",
				Fn:    e.resumeFn,
			}
		},
		Strategy: nav.ResumeStrategySpawnEn, // TODO: to come from an arg
	}

	return e.navigateWithLookAhead(
		runnerWith,
		resumption,
		summariseAfter,
	)
}

type ShrinkParams struct {
	Inputs  *common.ShrinkCommandInputs
	Viper   configuration.ViperConfig
	Configs *common.Configs
	Logger  *slog.Logger
	Vfs     storage.VirtualFS
}

func EnterShrink(
	params *ShrinkParams,
) error {
	var (
		agent common.ExecutionAgent
		err   error
	)

	if agent, err = ipc.New(
		params.Configs.Advanced,
		params.Inputs.ParamSet.Native.KnownBy,
	); err != nil {
		if errors.Is(err, ipc.ErrUsingDummyExecutor) {
			// todo: notify ui via bubbletea
			//
			fmt.Printf("===> 💥💥💥 REVERTING TO DUMMY EXECUTOR !!!!\n")
		}
	}

	finder := filing.NewFinder(params.Inputs, params.Configs.Advanced, params.Configs.Schemes)
	fileManager := filing.NewManager(params.Vfs, finder)
	entry := &ShrinkEntry{
		EntryBase: EntryBase{
			Inputs:      params.Inputs.Root,
			Agent:       agent,
			Viper:       params.Viper,
			Configs:     params.Configs,
			Log:         params.Logger,
			Vfs:         params.Vfs,
			FileManager: fileManager,
			FilterSetup: &filterSetup{
				inputs: params.Inputs,
				config: params.Configs.Advanced,
			},
			Registry: orc.NewRegistry(&common.SharedControllerInfo{
				Agent:       agent,
				Inputs:      params.Inputs,
				Finder:      fileManager.Finder(),
				FileManager: fileManager,
			},
				params.Configs,
			),
		},
		Inputs: params.Inputs,
	}

	return entry.run(params.Viper)
}
