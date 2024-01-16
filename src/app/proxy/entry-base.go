package proxy

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/lorax/boost"
	"github.com/snivilised/pixa/src/cfg"
)

type afterFunc func(*nav.TraverseResult, error)

func summariseAfter(result *nav.TraverseResult, err error) {
	measure := fmt.Sprintf("started: '%v', elapsed: '%v'",
		result.Session.StartedAt().Format(time.RFC1123), result.Session.Elapsed(),
	)
	files := result.Metrics.Count(nav.MetricNoFilesInvokedEn)
	folders := result.Metrics.Count(nav.MetricNoFoldersInvokedEn)
	summary := fmt.Sprintf("files: %v, folders: %v", files, folders)
	message := lo.Ternary(err == nil,
		fmt.Sprintf("🚩 navigation completed ok (%v) 💝 [%v]", summary, measure),
		fmt.Sprintf("🚩 error occurred during navigation (%v)💔 [%v]", err, measure),
	)
	fmt.Println(message)
}

// EntryBase is the base entry for all commands in pixa
type EntryBase struct {
	// some parts of the struct should go into a TraverseBase (anything to with
	// navigation such as Options)
	// with the rest going into cobrass.clif
	//
	Inputs      *RootCommandInputs
	Program     Executor
	Config      configuration.ViperConfig
	Options     *nav.TraverseOptions
	Registry    *ControllerRegistry
	ProfilesCFG cfg.ProfilesConfig
	SchemesCFG  cfg.SchemesConfig
	SamplerCFG  cfg.SamplerConfig
	AdvancedCFG cfg.AdvancedConfig
	Log         *slog.Logger
	Vfs         storage.VirtualFS
	FileManager *FileManager
	FilterSetup *filterSetup
}

func (e *EntryBase) ConfigureOptions(o *nav.TraverseOptions) {
	e.Options = o
	o.Hooks.QueryStatus = func(path string) (os.FileInfo, error) {
		fi, err := e.Vfs.Lstat(path)

		return fi, err
	}
	o.Hooks.ReadDirectory = func(dirname string) ([]fs.DirEntry, error) {
		contents, err := e.Vfs.ReadDir(dirname)
		if err != nil {
			return nil, err
		}

		return lo.Filter(contents, func(item fs.DirEntry, index int) bool {
			name := item.Name()

			// todo: get rid of ".$journal" hard-coding and get it form config
			// using finder.statics
			//
			return name != ".DS_Store" && !strings.Contains(name, ".$journal") && !strings.HasPrefix(name, ".")
		}), nil
	}

	if o.Store.FilterDefs == nil {
		switch {
		case e.Inputs.FoldersFam.Native.FoldersGlob != "":
			pattern := e.Inputs.FoldersFam.Native.FoldersGlob
			o.Store.FilterDefs = &nav.FilterDefinitions{
				Node: nav.FilterDef{
					Type:        nav.FilterTypeGlobEn,
					Description: fmt.Sprintf("--folders-gb(Z): '%v'", pattern),
					Pattern:     pattern,
					Scope:       nav.ScopeFolderEn,
				},
			}

		case e.Inputs.FoldersFam.Native.FoldersRexEx != "":
			pattern := e.Inputs.FoldersFam.Native.FoldersRexEx
			o.Store.FilterDefs = &nav.FilterDefinitions{
				Node: nav.FilterDef{
					Type:        nav.FilterTypeRegexEn,
					Description: fmt.Sprintf("--folders-rx(Y): '%v'", pattern),
					Pattern:     pattern,
					Scope:       nav.ScopeFolderEn,
				},
			}

		default:
			// TODO: there is still confusion here. Why do we need to set up
			// a default image filter in base, when base is only interested in folders?
			// shouldn't this default just be in shrink, which is interested in files.
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

	// setup sampling (sampling params needs to be defined on a new family in store)
	// This should not be here; move to root
	//
	if e.Inputs.ParamSet.Native.IsSampling {
		o.Store.Sampling.SampleType = nav.SampleTypeFilterEn
		o.Store.Sampling.SampleInReverse = e.Inputs.ParamSet.Native.Last

		if e.Inputs.ParamSet.Native.NoFiles > 0 {
			o.Store.Sampling.NoOf.Files = e.Inputs.ParamSet.Native.NoFiles
		}

		if e.Inputs.ParamSet.Native.NoFolders > 0 {
			o.Store.Sampling.NoOf.Folders = e.Inputs.ParamSet.Native.NoFolders
		}

		o.Store.Behaviours.Cascade.Skim = e.Inputs.CascadeFam.Native.Skim
		o.Store.Behaviours.Cascade.Depth = e.Inputs.CascadeFam.Native.Depth
	}

	// TODO: get the controller type properly, instead of hard coding to Sampler
	// This should not be here; move to root
	//
	if e.Registry == nil {
		e.Registry = NewControllerRegistry(&SharedControllerInfo{
			// Options:  e.Options,
			program:  e.Program,
			profiles: e.ProfilesCFG,
			sampler:  e.SamplerCFG,
		})
	}

	o.Monitor.Log = e.Log
}

func (e *EntryBase) navigate(
	optionsFn nav.TraverseOptionFn,
	with nav.CreateNewRunnerWith,
	resumption *nav.Resumption,
	after ...afterFunc,
) error {
	wgan := boost.NewAnnotatedWaitGroup("🍂 traversal", e.Log)
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
			NoW:         e.Inputs.WorkerPoolFam.Native.NoWorkers,
			JobsChOut:   make(boost.JobStream[nav.TraverseItemInput], DefaultJobsChSize),
		},
	}

	result, err := nav.New().With(with, runnerInfo).Run(
		nav.IfWithPoolUseContext(with, ctx, cancel)...,
	)

	for _, fn := range after {
		fn(result, err)
	}

	return err
}
