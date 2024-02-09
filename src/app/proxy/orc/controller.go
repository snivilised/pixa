package orc

import (
	"github.com/snivilised/cobrass"
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/extendio/collections"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type Controller struct {
	session *common.SessionControllerInfo
	private *common.PrivateControllerInfo
	configs *common.Configs
}

func New(
	session *common.SessionControllerInfo,
	configs *common.Configs,
) *Controller {
	return &Controller{
		session: session,
		private: &common.PrivateControllerInfo{},
		configs: configs,
	}
}

func (c *Controller) OnNewShrinkItem(item *nav.TraverseItem) error {
	// create a master path info here and pass into the sequences
	// to replace the individual properties on the step
	//
	pi := &common.PathInfo{
		Item:    item,
		Scheme:  c.session.Inputs.Root.ProfileFam.Native.Scheme,
		Profile: c.session.Inputs.Root.ProfileFam.Native.Profile,
		Origin:  item.Extension.Parent,
	}

	var sequence common.Sequence

	switch {
	case pi.Profile != "":
		sequence = c.sequence(pi, c.profile)

	case pi.Scheme != "":
		sequence = c.sequence(pi, c.scheme)

	default:
		sequence = c.sequence(pi, c.adhoc)
	}

	return c.Run(item, sequence)
}

type sequenceFunc func(pi *common.PathInfo) common.Sequence

func (c *Controller) sequence(pi *common.PathInfo, fn sequenceFunc) common.Sequence {
	return fn(pi)
}

func (c *Controller) profile(pi *common.PathInfo) common.Sequence {
	changed := c.session.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	combined := c.compose(pi.Profile, changed)
	step := &controllerStep{
		session:      c.session,
		thirdPartyCL: combined,
		sourcePath:   pi.Item.Path,
		profile:      pi.Profile,
		outputPath:   c.session.Inputs.ParamSet.Native.OutputPath,
	}

	return common.Sequence{step}
}

func (c *Controller) scheme(pi *common.PathInfo) common.Sequence {
	changed := c.session.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	schemeCfg, _ := c.configs.Schemes.Scheme(pi.Scheme) // scheme already validated
	sequence := make(common.Sequence, 0, len(schemeCfg.Profiles()))

	for _, current := range schemeCfg.Profiles() {
		combined := c.compose(current, changed)
		step := &controllerStep{
			session:      c.session,
			thirdPartyCL: combined,
			sourcePath:   pi.Item.Path,
			profile:      current,
			outputPath:   c.session.Inputs.ParamSet.Native.OutputPath,
		}

		sequence = append(sequence, step)
	}

	return sequence
}

func (c *Controller) adhoc(pi *common.PathInfo) common.Sequence {
	changed := c.session.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	step := &controllerStep{
		session:      c.session,
		thirdPartyCL: changed,
		sourcePath:   pi.Item.Path,
		outputPath:   c.session.Inputs.ParamSet.Native.OutputPath,
	}

	return common.Sequence{step}
}

func (c *Controller) compose(
	profileName string,
	secondary clif.ThirdPartyCommandLine,
) clif.ThirdPartyCommandLine {
	primary, _ := c.configs.Profiles.Profile(profileName) // profile already validated

	return cobrass.Evaluate(
		primary,
		c.session.Inputs.ParamSet.Native.ThirdPartySet.KnownBy,
		secondary,
	)
}

func (c *Controller) Run(item *nav.TraverseItem, sequence common.Sequence) error {
	var (
		zero common.Step
		err  error
	)

	iterator := collections.ForwardRunIt[common.Step, error](sequence, zero)
	each := func(step common.Step) error {
		// profile here on pi not set when profile set on command line
		// but thats ok, since a profile sequence is created and the
		// executive step itself does have the profile.
		return step.Run(&c.private.Pi)
	}
	while := func(_ common.Step, e error) bool {
		if err == nil {
			err = e
		}

		// TODO: this needs to change according to a new, not yet defined
		// setting, 'ContinueOnError'
		//
		return e == nil
	}

	c.private.Pi = common.PathInfo{
		Item:   item,
		Origin: item.Parent.Path,
		Scheme: c.session.FileManager.Finder().Scheme(),
	}

	// TODO: need to decide a proper policy for cleaning up
	// in the presence of an error. Do we allow the journal
	// file to remain in place? What happens if there is a timeout?
	// There are a few more things to decide about error handling.
	// Perhaps we have an error policy including one that implements
	// a retry.
	//
	if c.private.Pi.RunStep.Source, err = c.session.FileManager.Setup(
		&c.private.Pi,
	); err != nil {
		return err
	}

	iterator.RunAll(each, while)

	return c.session.FileManager.Tidy(&c.private.Pi)
}

func (c *Controller) Reset() {}
