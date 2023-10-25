package proxy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/lorax/boost"
)

type afterFunc func(*nav.TraverseResult, error)

type EntryBase struct {
	Inputs       *RootCommandInputs
	Program      Executor
	Config       configuration.ViperConfig
	ThirdPartyCL cobrass.ThirdPartyCommandLine
}

func (e *EntryBase) ConfigureOptions(o *nav.TraverseOptions) {
	// TODO: to apply the folder filters in combination with these
	// file filters, we need to define a custom compound
	// filter.
	//
	switch {
	case e.Inputs.ParamSet.Native.FilesGlob != "":
		o.Store.FilterDefs = &nav.FilterDefinitions{
			Node: nav.FilterDef{
				Type:        nav.FilterTypeGlobEn,
				Description: fmt.Sprintf("--files-gb(G): '%v'", e.Inputs.ParamSet.Native.FilesGlob),
				Pattern:     e.Inputs.ParamSet.Native.FilesGlob,
				Scope:       nav.ScopeLeafEn,
			},
		}

	case e.Inputs.ParamSet.Native.FilesRexEx != "":
		o.Store.FilterDefs = &nav.FilterDefinitions{
			Node: nav.FilterDef{
				Type:        nav.FilterTypeRegexEn,
				Description: fmt.Sprintf("--files-rx(X): '%v'", e.Inputs.ParamSet.Native.FilesRexEx),
				Pattern:     e.Inputs.ParamSet.Native.FilesRexEx,
				Scope:       nav.ScopeLeafEn,
			},
		}

	default:
		filterType := nav.FilterTypeRegexEn
		description := "Default image types supported by pixa"
		pattern := "\\.(jpe?g|png)$"

		o.Store.FilterDefs = &nav.FilterDefinitions{
			Node: nav.FilterDef{
				Type:        filterType,
				Description: description,
				Pattern:     pattern,
			},
			Children: nav.CompoundFilterDef{
				Type:        filterType,
				Description: description,
				Pattern:     pattern,
			},
		}
	}
}

func ResolvePath(path string) string {
	result := path

	if result[0] == '~' {
		if h, err := os.UserHomeDir(); err == nil {
			result = filepath.Join(h, result[1:])
		}
	} else {
		if absolute, absErr := filepath.Abs(path); absErr == nil {
			result = absolute
		}
	}

	return result
}

func (e *EntryBase) readProfile3rdPartyFlags() cobrass.ThirdPartyCommandLine {
	fmt.Printf("------> 💦💦💦 readProfile3rdPartyFlags: '%v'\n", e.Inputs.ParamSet.Native.Profile)

	return lo.TernaryF(e.Inputs.ParamSet.Native.Profile != "",
		func() []string {
			configPath := "profiles." + e.Inputs.ParamSet.Native.Profile
			return e.Config.GetStringSlice(configPath)
		},
		func() []string {
			return []string{}
		},
	)
}

func (e *EntryBase) navigate(
	optionsFn nav.TraverseOptionFn,
	with nav.CreateNewRunnerWith,
	resumption *nav.Resumption,
	after ...afterFunc,
) error {
	wgan := boost.NewAnnotatedWaitGroup("🍂 traversal")
	wgan.Add(1, navigatorRoutineName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runnerInfo := &nav.RunnerInfo{
		PrimeInfo: &nav.Prime{
			Path:      e.Inputs.ParamSet.Native.Directory,
			OptionsFn: optionsFn,
		},
		ResumeInfo: resumption,
		AccelerationInfo: &nav.Acceleration{
			WgAn:        wgan,
			RoutineName: navigatorRoutineName,
			NoW:         e.Inputs.ParamSet.Native.NoW,
			JobsChOut:   make(boost.JobStream[nav.TraverseItemInput], DefaultJobsChSize),
		},
	}

	result, err := nav.New().With(with, runnerInfo).Run(
		nav.IfWithPoolUseContext(with, ctx, cancel)...,
	)

	if len(after) > 0 {
		after[0](result, err)
	}

	measure := fmt.Sprintf("started: '%v', elapsed: '%v'",
		result.Session.StartedAt().Format(time.RFC1123), result.Session.Elapsed(),
	)
	files := result.Metrics.Count(nav.MetricNoFilesInvokedEn)
	folders := result.Metrics.Count(nav.MetricNoFoldersInvokedEn)
	summary := fmt.Sprintf("files: %v, folders: %v", files, folders)
	message := lo.Ternary(err == nil,
		fmt.Sprintf("navigation completed (%v) ✔️ [%v]", summary, measure),
		fmt.Sprintf("error occurred during navigation (%v)❌ [%v]", err, measure),
	)
	fmt.Println(message)

	return err
}
