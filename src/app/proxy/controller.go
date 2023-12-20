package proxy

import (
	"github.com/snivilised/cobrass"
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/extendio/collections"
	"github.com/snivilised/extendio/xfs/nav"
)

// 3rd party arguments provider: ad-hoc/profile/scheme
// ad-hoc: all on the fly arguments
// profile: ad-hoc and profile
// scheme: adhoc and profile from a scheme
//

type controller struct {
	shared *SharedControllerInfo
}

func (c *controller) profileSequence(
	name, itemPath string,
) Sequence {
	changed := c.shared.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	cl := c.composeProfileCL(name, changed)
	step := &magickStep{
		shared:       c.shared,
		thirdPartyCL: cl,
		sourcePath:   itemPath,
		profile:      name,
		// outputPath: ,
		// journalPath: ,
	}

	return Sequence{step}
}

func (c *controller) schemeSequence(
	name, itemPath string,
) Sequence {
	changed := c.shared.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	schemeCfg, _ := c.shared.sampler.Scheme(name) // scheme already validated
	sequence := make(Sequence, 0, len(schemeCfg.Profiles))

	for _, current := range schemeCfg.Profiles {
		cl := c.composeProfileCL(current, changed)
		step := &magickStep{
			shared:       c.shared,
			thirdPartyCL: cl,
			sourcePath:   itemPath,
			scheme:       name,
			profile:      current,
			// outputPath: ,
			// journalPath: ,
		}

		sequence = append(sequence, step)
	}

	return sequence
}

func (c *controller) adhocSequence(
	itemPath string,
) Sequence {
	changed := c.shared.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	step := &magickStep{
		shared:       c.shared,
		thirdPartyCL: changed,
		sourcePath:   itemPath,
		// outputPath: ,
		// journalPath: ,
	}

	return Sequence{step}
}

func (c *controller) composeProfileCL(
	profileName string,
	secondary clif.ThirdPartyCommandLine,
) clif.ThirdPartyCommandLine {
	primary, _ := c.shared.profiles.Profile(profileName) // profile already validated

	return cobrass.Evaluate(
		primary,
		c.shared.Inputs.ParamSet.Native.ThirdPartySet.KnownBy,
		secondary,
	)
}

func (c *controller) Run(item *nav.TraverseItem, sequence Sequence) error {
	var (
		zero      Step
		resultErr error
	)

	iterator := collections.ForwardRunIt[Step, error](sequence, zero)
	each := func(s Step) error {
		return s.Run(c.shared)
	}
	while := func(_ Step, err error) bool {
		if resultErr == nil {
			resultErr = err
		}

		// TODO: this needs to change according to a new, not yet defined
		// setting, 'ContinueOnError'
		//
		return err == nil
	}

	// TODO: need to decide a proper policy for cleaning up
	// in the presence of an error. Do we allow the journal
	// file to remain in place? What happens if there is a timeout?
	// There are a few more things to decide about error handling.
	// Perhaps we have an error policy including one that implements
	// a retry.
	//
	if err := c.shared.fileManager.Setup(item); err != nil {
		return err
	}

	iterator.RunAll(each, while)

	return c.shared.fileManager.Tidy()
}

func (c *controller) Reset() {
}
