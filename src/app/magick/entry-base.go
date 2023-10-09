package magick

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/lorax/boost"
)

type afterFunc func(*nav.TraverseResult, error)

type EntryBase struct {
	RootPS          *assistant.ParamSet[RootParameterSet]
	Program         Executor
	Config          configuration.ViperConfig
	thirdPartyFlags []string
}

func (e *EntryBase) ConfigureOptions(o *nav.TraverseOptions) {
	// TODO: to apply the folder filters in combination with these
	// file filters, we need to define a custom compound
	// filter.
	//
	switch {
	case e.RootPS.Native.FilesGlob != "":
		o.Store.FilterDefs = &nav.FilterDefinitions{
			Node: nav.FilterDef{
				Type:        nav.FilterTypeGlobEn,
				Description: fmt.Sprintf("--files-gb(G): '%v'", e.RootPS.Native.FilesGlob),
				Pattern:     e.RootPS.Native.FilesGlob,
				Scope:       nav.ScopeLeafEn,
			},
		}

	case e.RootPS.Native.FilesRexEx != "":
		o.Store.FilterDefs = &nav.FilterDefinitions{
			Node: nav.FilterDef{
				Type:        nav.FilterTypeRegexEn,
				Description: fmt.Sprintf("--files-rx(X): '%v'", e.RootPS.Native.FilesRexEx),
				Pattern:     e.RootPS.Native.FilesRexEx,
				Scope:       nav.ScopeLeafEn,
			},
		}

	default:
		filterType := nav.FilterTypeRegexEn
		description := "Default image types supported by pixa"
		pattern := "\\.(jpe?g|png|gif)$"

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

func (e *EntryBase) evaluate() {
	e.thirdPartyFlags = e.readProfile3rdPartyFlags()
}

/*
	func (e *EntryBase) expand(positional []string) []string {
		flags := append(positional, e.thirdPartyFlags...) //nolint:gocritic // no alternative option
		others := []string{"--version"}                   // need to extract the appropriate ones from ps
		flags = append(flags, others...)

		return flags
	}
*/
func (e *EntryBase) expand(positional []string) []string {
	flags := make([]string, 0, len(positional)+len(e.thirdPartyFlags))
	flags = append(flags, positional...)
	flags = append(flags, e.thirdPartyFlags...)

	return flags
}

func (e *EntryBase) readProfile3rdPartyFlags() []string {
	fmt.Printf("------> 💦💦💦 readProfile3rdPartyFlags: '%v'\n", e.RootPS.Native.Profile)

	return lo.TernaryF(e.RootPS.Native.Profile != "",
		func() []string {
			configPath := "profiles." + e.RootPS.Native.Profile
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
			Path:      e.RootPS.Native.Directory,
			OptionsFn: optionsFn,
		},
		ResumeInfo: resumption,
		AccelerationInfo: &nav.Acceleration{
			WgAn:        wgan,
			RoutineName: navigatorRoutineName,
			NoW:         e.RootPS.Native.NoW,
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

func (e *EntryBase) ReadProfile() []string {
	var c map[string]interface{}

	configPath := "profs." + e.RootPS.Native.Profile

	// ??? func(dc *mapstructure.DecoderConfig) {}
	if err := e.Config.UnmarshalKey(configPath, &c); err == nil {
		fmt.Printf("---> 🎆 configProfile: '%+v'\n", c)
	} else {
		fmt.Printf("---> 🎆 configProfile FAILED (%v)\n", err)
	}

	return []string{}
}
