package magick

import (
	"github.com/snivilised/extendio/xfs/nav"
)

type GenericEntry interface {
	ConfigureOptions(o *nav.TraverseOptions)
}

func GetTraverseOptionsFunc[E GenericEntry](entry E) func(o *nav.TraverseOptions) {
	return func(o *nav.TraverseOptions) {
		entry.ConfigureOptions(o)
	}
}
