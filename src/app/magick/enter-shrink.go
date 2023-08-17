package magick

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/extendio/xfs/nav"
)

type EntryBase struct {
	session nav.TraverseSession
}

func (e *EntryBase) ConfigureOptions(o *nav.TraverseOptions) {
	o.Store.FilterDefs = &nav.FilterDefinitions{
		Children: nav.CompoundFilterDef{
			Type:        nav.FilterTypeRegexEn,
			Description: "Image types supported by pixa",
			Pattern:     "\\.(jpe?g|png|gif)$",
		},
	}
}

type ShrinkEntry struct {
	EntryBase

	rps  *assistant.ParamSet[RootParameterSet]
	ps   *assistant.ParamSet[ShrinkParameterSet]
	jobs []string
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
	o.Callback = nav.LabelledTraverseCallback{
		Label: "Shrink Entry Callback",
		Fn: func(item *nav.TraverseItem) error {
			depth := item.Extension.Depth
			indicator := lo.Ternary(len(item.Children) > 0, "☀️", "🌊")

			lo.ForEach(item.Children, func(de fs.DirEntry, index int) {
				fullPath := filepath.Join(item.Path, de.Name())
				e.jobs = append(e.jobs, fullPath)
			})

			fmt.Printf(
				"---> %v SHRINK-CALLBACK: (depth:%v, files:%v) '%v'\n",
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

func Configure(entry *ShrinkEntry, o *nav.TraverseOptions) {
	entry.ConfigureOptions(o)
}

func GetTraverseOptionsFunc(entry *ShrinkEntry) func(o *nav.TraverseOptions) {
	// make this a generic?
	//
	return func(o *nav.TraverseOptions) {
		Configure(entry, o)
	}
}

func EnterShrink(
	rps *assistant.ParamSet[RootParameterSet],
	ps *assistant.ParamSet[ShrinkParameterSet],
) error {
	fmt.Printf("---> 🦠🦠🦠 Directory: '%v'\n", ps.Native.Directory)

	entry := &ShrinkEntry{
		rps: rps,
		ps:  ps,
	}
	session := &nav.PrimarySession{
		Path:     ps.Native.Directory,
		OptionFn: GetTraverseOptionsFunc(entry),
	}
	_, err := session.Init().Run()

	lo.ForEach(entry.jobs, func(i string, index int) {
		fmt.Printf("		===> ✨ job: '%v'\n", i)
	})

	summary := fmt.Sprintf("files: %v", len(entry.jobs))
	message := lo.Ternary(err == nil,
		fmt.Sprintf("navigation completed (%v) ✔️", summary),
		fmt.Sprintf("error occurred during navigation (%v)❌\n", err),
	)
	fmt.Println(message)

	return err
}
