package common

import (
	"github.com/snivilised/extendio/xfs/nav"
)

// both ui's need to perform the discover and primary, navigations
// each navigation:
// - start the navigation
// each navigation needs it own traverse info,
// tea has the following additional requirements:
// - differentiate between the start of navigation and interim completions
// - needs to subscribe to the output job channel so it can update progress
// - needs to access the callback function so it can wrap
//
// So the essential difference between the 2 is the need to access the output
// job channel; this also means it needs to know the end of discovery. This
// is why we need to discard the navigateWithLookAhead as just binds the 2
// phases together. Instead we replace this with the interaction, which
// separates out these 2 phases.

type (
	AfterFunc func(*nav.TraverseResult, error)

	UserInteraction interface {
		// Decorate allows the interaction to provide a wrapper around the callback.
		// If the interaction does not need it, then it just returns nil. Only
		// the Primary callback is decorated.
		//
		Decorate(target *nav.LabelledTraverseCallback) *nav.LabelledTraverseCallback

		// Discover navigation represents the pre-parse stage of the navigation
		//
		Discover(optionsFn nav.TraverseOptionFn,
			with nav.CreateNewRunnerWith,
			resumption *nav.Resumption,
		) error

		// Primary represents the main work stage of the navigation
		//
		Primary(optionsFn nav.TraverseOptionFn,
			with nav.CreateNewRunnerWith,
			resumption *nav.Resumption,
			after ...AfterFunc,
		) error
	}
)
