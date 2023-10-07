package magick

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/lorax/boost"
)

const (
	DefaultJobsChSize = 10
)

var (
	navigatorRoutineName = boost.GoRoutineName("✨ pixa-navigator")
)

type RootEntry struct {
	EntryBase

	rps   *assistant.ParamSet[RootParameterSet]
	files []string
}

func (e *RootEntry) ConfigureOptions(o *nav.TraverseOptions) {
	o.Notify.OnBegin = func(_ *nav.NavigationState) {
		fmt.Printf("===> 🛡️ beginning traversal ...\n")
	}
	o.Notify.OnEnd = func(result *nav.TraverseResult) {
		fmt.Printf("===> 🚩 finished traversal - folders '%v'\n",
			result.Metrics.Count(nav.MetricNoFoldersInvokedEn),
		)
	}
	o.Callback = nav.LabelledTraverseCallback{
		Label: "Root Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			depth := item.Extension.Depth
			indicator := lo.Ternary(len(item.Children) > 0, "☀️", "🌊")

			lo.ForEach(item.Children, func(de fs.DirEntry, index int) {
				fullPath := filepath.Join(item.Path, de.Name())
				e.files = append(e.files, fullPath)
			})

			fmt.Printf(
				"---> %v ROOT-CALLBACK: (depth:%v, files:%v) '%v'\n",
				indicator,
				depth, len(item.Children),
				item.Path,
			)

			return nil
		},
	}
	o.Store.Subscription = nav.SubscribeFoldersWithFiles
	o.Store.DoExtend = true
	e.EntryBase.ConfigureOptions(o)
}

func GetRootTraverseOptionsFunc(entry *RootEntry) func(o *nav.TraverseOptions) {
	// make this a generic?
	//
	return func(o *nav.TraverseOptions) {
		entry.ConfigureOptions(o)
	}
}

func EnterRoot(
	rps *assistant.ParamSet[RootParameterSet],
) error {
	fmt.Printf("---> 🦠🦠🦠 Directory: '%v'\n", rps.Native.Directory)

	entry := &RootEntry{
		rps: rps,
	}
	wgan := boost.NewAnnotatedWaitGroup("🍂 traversal")
	wgan.Add(1, navigatorRoutineName)

	ctx, cancel := context.WithCancel(context.Background())
	files := []string{}

	// createWith needs to be set according to command line parameters
	// nav.RunnerWithResume | nav.RunnerWithPool
	//
	createWith := nav.RunnerWithPool

	result, err := nav.New().With(createWith, &nav.RunnerInfo{
		PrimeInfo: &nav.Prime{
			Path:      rps.Native.Directory,
			OptionsFn: GetRootTraverseOptionsFunc(entry),
		},
		ResumeInfo: &nav.Resumption{
			RestorePath: "/json-path-to-come-from-a-flag-option/restore.json",
			Restorer: func(o *nav.TraverseOptions, active *nav.ActiveState) {
				o.Callback = nav.LabelledTraverseCallback{
					Label: "Root Entry Callback",
					Fn: func(item *nav.TraverseItem) error {
						depth := item.Extension.Depth
						indicator := lo.Ternary(len(item.Children) > 0, "☀️", "🌊")

						lo.ForEach(item.Children, func(de fs.DirEntry, index int) {
							fullPath := filepath.Join(item.Path, de.Name())
							files = append(files, fullPath)
						})

						fmt.Printf(
							"---> %v ROOT-CALLBACK: (depth:%v, files:%v) '%v'\n",
							indicator,
							depth, len(item.Children),
							item.Path,
						)

						return nil
					},
				}
			},
			Strategy: nav.ResumeStrategySpawnEn, // to come from an arg
		},
		AccelerationInfo: &nav.Acceleration{
			WgAn:        wgan,
			RoutineName: navigatorRoutineName,
			NoW:         rps.Native.NoW,
			JobsChOut:   make(boost.JobStream[nav.TraverseItemInput], DefaultJobsChSize),
		},
	}).Run(
		nav.IfWithPoolUseContext(createWith, ctx, cancel)...,
	)

	lo.ForEach(entry.files, func(f string, _ int) {
		fmt.Printf("		===> 🔆 candidate file: '%v'\n", f)
	})

	no := result.Metrics.Count(nav.MetricNoChildFilesFoundEn)
	summary := fmt.Sprintf("files: %v", no)
	message := lo.Ternary(err == nil,
		fmt.Sprintf("navigation completed (%v) ✔️", summary),
		fmt.Sprintf("error occurred during navigation (%v)❌\n", err),
	)
	fmt.Println(message)

	return err
}
