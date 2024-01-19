package proxy

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/cfg"
)

type ShrinkEntry struct {
	EntryBase
	Inputs *ShrinkCommandInputs
}

func FilenameWithoutExtension(name string) string {
	return strings.TrimSuffix(name, path.Ext(name))
}

func (e *ShrinkEntry) LookAheadOptionsFn(o *nav.TraverseOptions) {
	e.ConfigureOptions(o)
	o.Callback = &nav.LabelledTraverseCallback{
		Label: "LookAhead: Shrink Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			if strings.Contains(item.Path, DejaVu) {
				return fs.SkipDir
			}
			journal := e.FileManager.Finder.JournalFullPath(item)

			return e.FileManager.Create(journal)
		},
	}
}

func (e *ShrinkEntry) PrincipalOptionsFn(o *nav.TraverseOptions) {
	e.ConfigureOptions(o)

	o.Callback = &nav.LabelledTraverseCallback{
		Label: "Principal: Shrink Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			if strings.Contains(item.Path, DejaVu) {
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
	o.Store.FilterDefs = e.FilterSetup.getDefs(e.FileManager.Finder.statics)
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
	if strings.HasPrefix(item.Extension.Name, DejaVu) {
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
	Inputs      *ShrinkCommandInputs
	Config      configuration.ViperConfig
	ProfilesCFG cfg.ProfilesConfig
	SchemesCFG  cfg.SchemesConfig
	SamplerCFG  cfg.SamplerConfig
	AdvancedCFG cfg.AdvancedConfig
	Logger      *slog.Logger
	Vfs         storage.VirtualFS
}

func EnterShrink(
	params *ShrinkParams,
) error {
	agent := newAgent(
		params.AdvancedCFG,
		params.Inputs.ParamSet.Native.KnownBy,
	)
	finder := newPathFinder(params.Inputs, params.AdvancedCFG, params.SchemesCFG)
	fileManager := &FileManager{
		vfs:    params.Vfs,
		Finder: finder,
	}
	entry := &ShrinkEntry{
		EntryBase: EntryBase{
			Inputs:      params.Inputs.Root,
			Agent:       agent,
			Config:      params.Config,
			ProfilesCFG: params.ProfilesCFG,
			SchemesCFG:  params.SchemesCFG,
			SamplerCFG:  params.SamplerCFG,
			AdvancedCFG: params.AdvancedCFG,
			Log:         params.Logger,
			Vfs:         params.Vfs,
			FileManager: fileManager,
			FilterSetup: &filterSetup{
				inputs: params.Inputs,
				config: params.AdvancedCFG,
			},
			Registry: NewControllerRegistry(&SharedControllerInfo{
				agent:       agent,
				profiles:    params.ProfilesCFG,
				schemes:     params.SchemesCFG,
				sampler:     params.SamplerCFG,
				advanced:    params.AdvancedCFG,
				Inputs:      params.Inputs,
				finder:      fileManager.Finder,
				fileManager: fileManager,
			}),
		},
		Inputs: params.Inputs,
	}

	return entry.run(params.Config)
}
