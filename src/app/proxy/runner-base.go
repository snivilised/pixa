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

type baseRunner struct { // rename to be a controller instead of a runner
	shared *SharedRunnerInfo
}

func (r *baseRunner) profileSequence(
	name, itemPath string,
) Sequence {
	changed := r.shared.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	cl := r.composeProfileCL(name, changed)
	step := &magickStep{
		shared:       r.shared,
		thirdPartyCL: cl,
		sourcePath:   itemPath,
		profile:      name,
		// outputPath: ,
		// journalPath: ,
	}

	return Sequence{step}
}

func (r *baseRunner) schemeSequence(
	name, itemPath string,
) Sequence {
	changed := r.shared.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	schemeCfg, _ := r.shared.sampler.Scheme(name) // scheme already validated
	sequence := make(Sequence, 0, len(schemeCfg.Profiles))

	for _, current := range schemeCfg.Profiles {
		cl := r.composeProfileCL(current, changed)
		step := &magickStep{
			shared:       r.shared,
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

func (r *baseRunner) adhocSequence(
	itemPath string,
) Sequence {
	changed := r.shared.Inputs.ParamSet.Native.ThirdPartySet.LongChangedCL
	step := &magickStep{
		shared:       r.shared,
		thirdPartyCL: changed,
		sourcePath:   itemPath,
		// outputPath: ,
		// journalPath: ,
	}

	return Sequence{step}
}

func (r *baseRunner) composeProfileCL(
	profileName string,
	secondary clif.ThirdPartyCommandLine,
) clif.ThirdPartyCommandLine {
	primary, _ := r.shared.profiles.Profile(profileName) // profile already validated

	return cobrass.Evaluate(
		primary,
		r.shared.Inputs.ParamSet.Native.ThirdPartySet.KnownBy,
		secondary,
	)
}

func (r *baseRunner) Run(item *nav.TraverseItem, sequence Sequence) error {
	var (
		zero      Step
		resultErr error
	)

	iterator := collections.ForwardRunIt[Step, error](sequence, zero)
	each := func(s Step) error {
		return s.Run(r.shared)
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
	if err := r.shared.fileManager.Setup(item); err != nil {
		return err
	}

	iterator.RunAll(each, while)

	return r.shared.fileManager.Tidy()
}

func (r *baseRunner) Reset() {
}
