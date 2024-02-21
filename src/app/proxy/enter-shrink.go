package proxy

import (
	"io/fs"
	"log/slog"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/app/proxy/filing"
	"github.com/snivilised/pixa/src/app/proxy/ipc"
	"github.com/snivilised/pixa/src/app/proxy/orc"
	"github.com/snivilised/pixa/src/app/proxy/user"
)

type ShrinkEntry struct {
	EntryBase
	Inputs *common.ShrinkCommandInputs
}

func (e *ShrinkEntry) DiscoverOptionsFn(o *nav.TraverseOptions) {
	e.ConfigureOptions(o)
	o.Callback = &nav.LabelledTraverseCallback{
		Label: "Discovery: Shrink Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			if strings.Contains(item.Path, e.FileManager.Finder().Statics().TrashTag()) {
				return fs.SkipDir
			}
			journal := e.FileManager.Finder().JournalFullPath(item)

			presentation := lo.Ternary(e.Inputs.Root.TextualFam.Native.IsNoTui,
				"🧩 linear", "💄 textual",
			)

			e.Log.Debug("💎💎 Shrink Discovery Callback",
				slog.String("name", item.Extension.Name),
				slog.String("sub-path", item.Extension.SubPath),
				slog.Int("depth", item.Extension.Depth),
				slog.String("presentation", presentation),
			)

			if e.Inputs.Root.PreviewFam.Native.DryRun {
				return nil
			}

			return e.FileManager.Create(journal, false)
		},
	}
}

func (e *ShrinkEntry) PrincipalOptionsFn(o *nav.TraverseOptions) {
	e.ConfigureOptions(o)

	o.Notify.OnBegin = func(_ *nav.NavigationState) {
		e.Log.Info("===> 🛡️  beginning traversal ...")
	}

	o.Callback = e.EntryBase.Interaction.Decorate(&nav.LabelledTraverseCallback{
		Label: "Principal: Shrink Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			if strings.Contains(item.Path, e.FileManager.Finder().Statics().TrashTag()) {
				return fs.SkipDir
			}

			presentation := lo.Ternary(e.Inputs.Root.TextualFam.Native.IsNoTui,
				"🧩 linear", "💄 textual",
			)

			e.Log.Debug("🌀🌀 Shrink Principle Callback",
				slog.String("name", item.Extension.Name),
				slog.String("sub-path", item.Extension.SubPath),
				slog.Int("depth", item.Extension.Depth),
				slog.String("presentation", presentation),
			)

			controller := e.Registry.Get()
			defer e.Registry.Put(controller)

			return controller.OnNewShrinkItem(item)
		},
	})
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

func (e *ShrinkEntry) resumeFn(item *nav.TraverseItem) error {
	if strings.HasPrefix(item.Extension.Name, e.FileManager.Finder().Statics().TrashTag()) {
		return fs.SkipDir
	}

	depth := item.Extension.Depth

	e.Log.Debug("🎙️🎙️ Shrink Restore Callback",
		slog.String("name", item.Extension.Name),
		slog.String("sub-path", item.Extension.SubPath),
		slog.Int("depth", depth),
	)

	controller := e.Registry.Get()
	defer e.Registry.Put(controller)

	return controller.OnNewShrinkItem(item)
}

func (e *ShrinkEntry) run() (result *nav.TraverseResult, err error) {
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

	result, err = e.EntryBase.Interaction.Traverse(user.NewWalkInfo(e.DiscoverOptionsFn,
		e.PrincipalOptionsFn,
		runnerWith,
		resumption,
		e.Inputs,
	))

	return result, err
}

type ShrinkParams struct {
	Inputs *common.ShrinkCommandInputs
	Viper  configuration.ViperConfig
	Logger *slog.Logger
	Vfs    storage.VirtualFS
}

func EnterShrink(
	params *ShrinkParams,
) (*nav.TraverseResult, error) {
	var (
		agent common.ExecutionAgent
		err   error
	)

	schemes := params.Inputs.Root.Configs.Schemes
	selectedScheme := params.Inputs.Root.ProfileFam.Native.Scheme
	scheme, _ := schemes.Scheme(selectedScheme)
	arity := lo.TernaryF(scheme == nil,
		func() uint {
			return 1
		},
		func() uint {
			return uint(len(scheme.Profiles()))
		},
	)
	finder := filing.NewFinder(&filing.NewFinderInfo{
		Advanced:   params.Inputs.Root.Configs.Advanced,
		Schemes:    schemes,
		Scheme:     selectedScheme,
		OutputPath: params.Inputs.ParamSet.Native.OutputPath,
		TrashPath:  params.Inputs.ParamSet.Native.TrashPath,
		Observer:   params.Inputs.Root.Observers.PathFinder,
		Arity:      arity,
	})
	fileManager := filing.NewManager(params.Vfs, finder,
		params.Inputs.Root.PreviewFam.Native.DryRun,
	)

	if agent, err = ipc.New(
		params.Inputs.Root.Configs.Advanced,
		params.Inputs.ParamSet.Native.KnownBy,
		fileManager,
		params.Inputs.Root.PreviewFam.Native.DryRun,
	); err != nil {
		if errors.Is(err, ipc.ErrUseDummyExecutor) {
			params.Logger.Warn("===> 💥💥💥 REVERTING TO DUMMY EXECUTOR !!!!")

			agent = ipc.Pacify(
				params.Inputs.Root.Configs.Advanced,
				params.Inputs.ParamSet.Native.KnownBy,
				fileManager,
				ipc.PacifyWithDummy,
			)
		} else if errors.Is(err, ipc.ErrUnsupportedExecutor) {
			params.Logger.Error("===> 💥💥💥 Undefined EXECUTOR: '%v' !!!!",
				slog.String("name", params.Inputs.Root.Configs.Advanced.Executable().Symbol()),
			)

			return nil, err
		}
	}

	interaction := user.NewInteraction(
		params.Inputs,
		params.Logger,
		arity,
	)
	entry := &ShrinkEntry{
		EntryBase: EntryBase{
			Inputs:      params.Inputs.Root,
			Agent:       agent,
			Interaction: interaction,
			Viper:       params.Viper,
			Log:         params.Logger,
			Vfs:         params.Vfs,
			FileManager: fileManager,
			FilterSetup: &filterSetup{
				inputs: params.Inputs,
			},
			Registry: orc.NewRegistry(&common.SessionControllerInfo{
				Agent:       agent,
				Inputs:      params.Inputs,
				FileManager: fileManager,
				Interaction: interaction,
			},
				params.Inputs.Root.Configs,
			),
		},
		Inputs: params.Inputs,
	}

	return entry.run()
}
