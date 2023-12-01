package proxy

import (
	"github.com/snivilised/extendio/xfs/nav"
)

type SamplerRunner struct {
	baseRunner
}

func (r *SamplerRunner) OnNewShrinkItem(item *nav.TraverseItem,
	positional []string,
) error {
	_ = positional

	profileName := r.shared.Inputs.Root.ProfileFam.Native.Profile
	schemeName := r.shared.Inputs.Root.ProfileFam.Native.Scheme

	var sequence Sequence

	switch {
	case profileName != "":
		sequence = r.profileSequence(profileName, item.Path)

	case schemeName != "":
		sequence = r.schemeSequence(schemeName, item.Path)

	default:
		sequence = r.adhocSequence(item.Path)
	}

	return r.Run(item, sequence)
}
