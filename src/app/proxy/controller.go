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
	shared  *SharedControllerInfo
	private *privateControllerInfo
}

func (c *controller) OnNewShrinkItem(item *nav.TraverseItem) error {
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

func (c *controller) profileSequence(
	pi *pathInfo,
) Sequence {
	changed := c.shared.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	cl := c.composeProfileCL(pi.profile, changed)
	step := &executionStep{
		shared:       c.shared,
		thirdPartyCL: cl,
		sourcePath:   pi.item.Path,
		profile:      pi.profile,
		outputPath:   c.shared.Inputs.ParamSet.Native.OutputPath,
		// journalPath: ,
	}

	return Sequence{step}
}

func (c *controller) schemeSequence(
	pi *pathInfo,
) Sequence {
	changed := c.shared.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	schemeCfg, _ := c.shared.schemes.Scheme(pi.scheme) // scheme already validated
	sequence := make(Sequence, 0, len(schemeCfg.Profiles))

	for _, current := range schemeCfg.Profiles {
		cl := c.composeProfileCL(current, changed)
		step := &executionStep{
			shared:       c.shared,
			thirdPartyCL: cl,
			sourcePath:   pi.item.Path,
			profile:      current,
			outputPath:   c.shared.Inputs.ParamSet.Native.OutputPath,
			// journalPath: ,
		}

		sequence = append(sequence, step)
	}

	return sequence
}

func (c *controller) adhocSequence(
	pi *pathInfo,
) Sequence {
	changed := c.shared.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	step := &executionStep{
		shared:       c.shared,
		thirdPartyCL: changed,
		sourcePath:   pi.item.Path,
		outputPath:   c.shared.Inputs.ParamSet.Native.OutputPath,
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
		zero Step
		err  error
	)

	iterator := collections.ForwardRunIt[Step, error](sequence, zero)
	each := func(step Step) error {
		return step.Run(&c.private.pi)
	}
	while := func(_ Step, e error) bool {
		if err == nil {
			err = e
		}

		// TODO: this needs to change according to a new, not yet defined
		// setting, 'ContinueOnError'
		//
		return e == nil
	}

	c.private.pi = pathInfo{
		item:   item,
		origin: item.Parent.Path,
		scheme: c.shared.finder.Scheme,
	}

	// TODO: need to decide a proper policy for cleaning up
	// in the presence of an error. Do we allow the journal
	// file to remain in place? What happens if there is a timeout?
	// There are a few more things to decide about error handling.
	// Perhaps we have an error policy including one that implements
	// a retry.
	//
	if c.private.pi.runStep.Source, err = c.shared.fileManager.Setup(
		&c.private.pi,
	); err != nil {
		return err
	}

	iterator.RunAll(each, while)

	return c.shared.fileManager.Tidy(&c.private.pi)
}

func (c *controller) Reset() {}
