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

	profileName := c.shared.Inputs.Root.ProfileFam.Native.Profile
	schemeName := c.shared.Inputs.Root.ProfileFam.Native.Scheme

	var sequence Sequence

	switch {
	case profileName != "":
		sequence = c.profileSequence(profileName, item.Path)

	case schemeName != "":
		sequence = c.schemeSequence(schemeName, item.Path)

	default:
		sequence = c.adhocSequence(item.Path)
	}

	return c.Run(item, sequence)
}
