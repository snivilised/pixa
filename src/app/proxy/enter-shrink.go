package proxy

import (
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/extendio/xfs/storage"
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
			journalPath := e.FileManager.Finder.JournalFile(item)
			err := e.FileManager.Create(journalPath)

			fmt.Printf(
				"---> 💧💧 SHRINK-JOURNAL-FILE: (create journal:%v) '%v'\n",
				journalPath,
				item.Path,
			)

			return err
		},
	}

	defs := e.getFilterDefs()
	o.Store.FilterDefs = defs
}

func (e *ShrinkEntry) getFilterDefs() *nav.FilterDefinitions {
	var (
		file, folder  *nav.FilterDef
		defs          *nav.FilterDefinitions
		folderDefined = true
		pattern       string
	)

	switch {
	case e.Inputs.FilesFam.Native.FilesGlob != "":
		pattern = e.Inputs.FilesFam.Native.FilesGlob
		file = &nav.FilterDef{
			Type:        nav.FilterTypeGlobEn,
			Description: fmt.Sprintf("--files-gb(G): '%v'", pattern),
			Pattern:     pattern,
			Scope:       nav.ScopeFileEn,
		}

	case e.Inputs.FilesFam.Native.FilesRexEx != "":
		pattern = e.Inputs.FilesFam.Native.FilesRexEx
		file = &nav.FilterDef{
			Type:        nav.FilterTypeRegexEn,
			Description: fmt.Sprintf("--files-rx(X): '%v'", pattern),
			Pattern:     pattern,
			Scope:       nav.ScopeFileEn,
		}

	default:
		pattern = "(?i).(jpe?g|png)$"
		file = &nav.FilterDef{
			Type:        nav.FilterTypeRegexEn,
			Description: fmt.Sprintf("--files-rx(X): '%v'", pattern),
			Pattern:     pattern,
			Scope:       nav.ScopeFileEn,
		}
	}

	switch {
	case e.Inputs.Root.FoldersFam.Native.FoldersGlob != "":
		pattern = e.Inputs.Root.FoldersFam.Native.FoldersRexEx
		folder = &nav.FilterDef{
			Type:        nav.FilterTypeGlobEn,
			Description: fmt.Sprintf("--folders-gb(Z): '%v'", pattern),
			Pattern:     pattern,
			Scope:       nav.ScopeFolderEn | nav.ScopeLeafEn,
		}

	case e.Inputs.Root.FoldersFam.Native.FoldersRexEx != "":
		pattern = e.Inputs.Root.FoldersFam.Native.FoldersRexEx
		folder = &nav.FilterDef{
			Type:        nav.FilterTypeRegexEn,
			Description: fmt.Sprintf("--folders-rx(Y): '%v'", pattern),
			Pattern:     pattern,
			Scope:       nav.ScopeFolderEn | nav.ScopeLeafEn,
		}

	default:
		folderDefined = false
	}

	switch {
	case folderDefined:
		defs = &nav.FilterDefinitions{
			Node: nav.FilterDef{
				Type: nav.FilterTypePolyEn,
				Poly: &nav.PolyFilterDef{
					File:   *file,
					Folder: *folder,
				},
			},
		}

	default:
		defs = &nav.FilterDefinitions{
			Node: *file,
		}
	}

	return lo.TernaryF(pattern != "",
		func() *nav.FilterDefinitions {
			return defs
		},
		func() *nav.FilterDefinitions {
			return nil
		},
	)
}

func (e *ShrinkEntry) PrincipalOptionsFn(o *nav.TraverseOptions) {
	e.ConfigureOptions(o)
	o.Callback = &nav.LabelledTraverseCallback{
		Label: "Principal: Shrink Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			depth := item.Extension.Depth

			fmt.Printf(
				"---> 🌀🌀 SHRINK-CALLBACK-FILE: (depth:%v) '%v'\n",
				depth,
				item.Path,
			)

			controller := e.Registry.Get()
			defer e.Registry.Put(controller)

			return controller.OnNewShrinkItem(item)
		},
	}
}

func (e *ShrinkEntry) createFinder() *PathFinder {
	finder := &PathFinder{
		Scheme:          e.Inputs.Root.ProfileFam.Native.Scheme,
		ExplicitProfile: e.Inputs.Root.ProfileFam.Native.Profile,
		arity:           1,
	}

	if finder.Scheme != "" {
		schemeCFG, _ := e.SchemesCFG.Scheme(finder.Scheme)
		finder.arity = len(schemeCFG.Profiles)
	}

	if e.Inputs.ParamSet.Native.OutputPath != "" {
		finder.Output = e.Inputs.ParamSet.Native.OutputPath
	} else {
		finder.transparentInput = true
	}

	if e.Inputs.ParamSet.Native.TrashPath != "" {
		finder.Trash = e.Inputs.ParamSet.Native.TrashPath
	}

	finder.statics = &staticInfo{
		adhoc:   e.AdvancedCFG.AdhocLabel(),
		journal: e.AdvancedCFG.JournalLabel(),
		legacy:  e.AdvancedCFG.LegacyLabel(),
		trash:   e.AdvancedCFG.TrashLabel(),
	}

	return finder
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
	o.Store.Subscription = nav.SubscribeFiles

	e.EntryBase.ConfigureOptions(o)

	finder := e.createFinder()
	e.FileManager = &FileManager{
		vfs:    e.Vfs,
		Finder: finder,
	}

	e.Registry = NewControllerRegistry(&SharedControllerInfo{
		Options:     e.Options,
		program:     e.Program,
		profiles:    e.ProfilesCFG,
		schemes:     e.SchemesCFG,
		sampler:     e.SamplerCFG,
		advanced:    e.AdvancedCFG,
		Inputs:      e.Inputs,
		finder:      finder,
		fileManager: e.FileManager,
	})
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
	depth := item.Extension.Depth
	indicator := lo.Ternary(len(item.Children) > 0, "☀️", "🌊")

	fmt.Printf(
		"---> 🎙️🎙️ %v SHRINK-RESTORE-CALLBACK: (depth:%v) '%v'\n",
		indicator,
		depth,
		item.Path,
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

func EnterShrink(
	inputs *ShrinkCommandInputs,
	program Executor,
	config configuration.ViperConfig,
	profilesCFG ProfilesConfig,
	schemesCFG SchemesConfig,
	samplerCFG SamplerConfig,
	advancedCFG AdvancedConfig,
	vfs storage.VirtualFS,
) error {
	fmt.Printf("---> 🔊🔊 Directory: '%v'\n", inputs.Root.ParamSet.Native.Directory)

	entry := &ShrinkEntry{
		EntryBase: EntryBase{
			Inputs:      inputs.Root,
			Program:     program,
			Config:      config,
			ProfilesCFG: profilesCFG,
			SchemesCFG:  schemesCFG,
			SamplerCFG:  samplerCFG,
			AdvancedCFG: advancedCFG,
			Vfs:         vfs,
		},
		Inputs: inputs,
	}

	return entry.run(config)
}
