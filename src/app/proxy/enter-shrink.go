package proxy

import (
	"fmt"
	"path"
	"path/filepath"
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
			// TODO: get the journal filename from path-finder
			//
			withoutExt := FilenameWithoutExtension(item.Extension.Name) + ".journal.txt"
			pathWithoutExt := filepath.Join(item.Extension.Parent, withoutExt)

			// TODO(put this back in): only create file if not exits
			//
			// file, err := os.Create(pathWithoutExt) // TODO: use vfs

			// ??? if err == nil {
			// 	defer file.Close()
			// }

			fmt.Printf(
				"---> 💧💧 SHRINK-JOURNAL-FILE(disabled!!): (create journal:%v) '%v'\n",
				pathWithoutExt,
				item.Path,
			)

			return nil
		},
	}

	switch {
	case e.Inputs.FilesFam.Native.FilesGlob != "":
		pattern := e.Inputs.FilesFam.Native.FilesGlob
		o.Store.FilterDefs = &nav.FilterDefinitions{
			Node: nav.FilterDef{
				Type:        nav.FilterTypeGlobEn,
				Description: fmt.Sprintf("--files-gb(G): '%v'", pattern),
				Pattern:     pattern,
				Scope:       nav.ScopeFileEn,
			},
		}

	case e.Inputs.FilesFam.Native.FilesRexEx != "":
		pattern := e.Inputs.FilesFam.Native.FilesRexEx
		o.Store.FilterDefs = &nav.FilterDefinitions{
			Node: nav.FilterDef{
				Type:        nav.FilterTypeRegexEn,
				Description: fmt.Sprintf("--files-rx(X): '%v'", pattern),
				Pattern:     pattern,
				Scope:       nav.ScopeFileEn,
			},
		}
	}
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

			positional := []string{
				fmt.Sprintf("'%v'", item.Path),
			}

			runner := e.Registry.Get()
			defer e.Registry.Put(runner)

			return runner.OnNewShrinkItem(item, positional)
		},
	}
}

func (e *ShrinkEntry) createFinder() *PathFinder {
	finder := &PathFinder{
		Scheme:  e.Inputs.Root.ProfileFam.Native.Scheme,
		Profile: e.Inputs.Root.ProfileFam.Native.Profile,
		behaviours: strategies{
			output:   &inlineOutputStrategy{},
			deletion: &inlineDeletionStrategy{},
		},
	}

	if e.Inputs.ParamSet.Native.OutputPath != "" {
		finder.behaviours.output = &ejectOutputStrategy{}
	}

	if e.Inputs.ParamSet.Native.TrashPath != "" {
		finder.behaviours.deletion = &ejectOutputStrategy{}
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
	e.Registry = NewRunnerRegistry(&SharedRunnerInfo{
		Type:     RunnerTypeSamplerEn, // TODO: to come from an arg !!!
		Options:  e.Options,
		program:  e.Program,
		profiles: e.ProfilesCFG,
		sampler:  e.SamplerCFG,
		Inputs:   e.Inputs,
		finder:   finder,
		fileManager: &FileManager{
			vfs:    e.Vfs,
			finder: finder,
		},
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

	positional := []string{
		fmt.Sprintf("'%v'", item.Path),
	}

	runner := e.Registry.Get()
	defer e.Registry.Put(runner)

	return runner.OnNewShrinkItem(item, positional)
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
	samplerCFG SamplerConfig,
	vfs storage.VirtualFS,
) error {
	fmt.Printf("---> 🔊🔊 Directory: '%v'\n", inputs.Root.ParamSet.Native.Directory)

	entry := &ShrinkEntry{
		EntryBase: EntryBase{
			Inputs:      inputs.Root,
			Program:     program,
			Config:      config,
			ProfilesCFG: profilesCFG,
			SamplerCFG:  samplerCFG,
			Vfs:         vfs,
		},
		Inputs: inputs,
	}

	return entry.run(config)
}
