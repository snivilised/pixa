package filing

import (
	"path/filepath"

	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

func ComposeFake(name, label string) string {
	return FilenameWithoutExtension(name) + label + filepath.Ext(name)
}

type dryRunPathFinderDecorator struct {
	decorated *PathFinder
}

func (d *dryRunPathFinderDecorator) Transfer(info *common.PathInfo) (folder, file string) {
	if d.decorated.TransparentInput() {
		folder = info.Origin
		file = info.Item.Extension.Name
	} else {
		folder, file = d.decorated.Transfer(info)
	}

	return folder, file
}

func (d *dryRunPathFinderDecorator) Result(info *common.PathInfo) (folder, file string) {
	if d.decorated.TransparentInput() {
		folder = info.Origin
		file = ComposeFake(info.Item.Extension.Name, d.decorated.Statics().Fake)
	} else {
		folder, file = d.decorated.Result(info)
	}

	return folder, file
}

func (d *dryRunPathFinderDecorator) TransparentInput() bool {
	return d.decorated.TransparentInput()
}

func (d *dryRunPathFinderDecorator) JournalFullPath(item *nav.TraverseItem) string {
	return d.decorated.JournalFullPath(item)
}

func (d *dryRunPathFinderDecorator) Statics() *common.StaticInfo {
	return d.decorated.Statics()
}

func (d *dryRunPathFinderDecorator) Scheme() string {
	return d.decorated.Scheme()
}
