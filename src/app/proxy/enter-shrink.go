package proxy

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
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

func FilenameWithoutExtension(name string) string {
	return strings.TrimSuffix(name, path.Ext(name))
}

func (e *ShrinkEntry) principalFn(item *nav.TraverseItem) error {
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
}

func (e *ShrinkEntry) LookAheadOptionsFn(o *nav.TraverseOptions) {
	e.ConfigureOptions(o)
	o.Callback = &nav.LabelledTraverseCallback{
		Label: "LookAhead: Shrink Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			withoutExt := FilenameWithoutExtension(item.Extension.Name) + ".journal.txt"
			pathWithoutExt := filepath.Join(item.Extension.Parent, withoutExt)

			// TODO: only create file if not exits
			//
			file, err := os.Create(pathWithoutExt) // TODO: use vfs

			if err == nil {
				defer file.Close()
			}

			fmt.Printf(
				"---> 📜 SHRINK-JOURNAL-FILE: (create journal:%v) '%v'\n",
				pathWithoutExt,
				item.Path,
			)

			return err
		},
	}
}

func (e *ShrinkEntry) PrincipalOptionsFn(o *nav.TraverseOptions) {
	e.ConfigureOptions(o)
	o.Callback = &nav.LabelledTraverseCallback{
		Label: "Principal: Shrink Entry Callback",
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
	o.Store.DoExtend = true

	e.EntryBase.ConfigureOptions(o)
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
		"---> %v SHRINK-RESTORE-CALLBACK: (depth:%v) '%v'\n",
		indicator,
		depth,
		item.Path,
	)

	positional := []string{
		fmt.Sprintf("'%v'", item.Path),
	}

	return e.Program.Execute(clif.Expand(positional, e.ThirdPartyCL)...)
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
		Strategy: nav.ResumeStrategySpawnEn, // to come from an arg
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
