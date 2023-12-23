package proxy

import (
	"github.com/snivilised/extendio/xfs/nav"
)

type SamplerController struct {
	controller
}

func (c *SamplerController) OnNewShrinkItem(item *nav.TraverseItem,
	positional []string,
) error {
	_ = positional

	// create a master path info here and pass into the sequences
	// to replace the individual properties on the step
	//
	pi := &pathInfo{
		item:    item,
		scheme:  c.shared.Inputs.Root.ProfileFam.Native.Scheme,
		profile: c.shared.Inputs.Root.ProfileFam.Native.Profile,
		origin:  item.Extension.Parent,
	}

	var sequence Sequence

	switch {
	case pi.profile != "":
		sequence = c.profileSequence(pi)

	case pi.scheme != "":
		sequence = c.schemeSequence(pi)

	default:
		sequence = c.adhocSequence(pi)
	}

	return c.Run(item, sequence)
}
