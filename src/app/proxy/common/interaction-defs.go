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

	// DiscoveredMsg indicates that the discovery phase has completed.
	//
	DiscoveredMsg struct {
		Result *nav.TraverseResult
		Err    error
	}

	// ProgressMsg indicates a job completion
	//
	ProgressMsg struct {
		Source      string
		Destination string
		Scheme      string
		Profile     string
		Err         error
	}

	// FinishedMsg indicates end of traversal
	//
	FinishedMsg struct {
		Result *nav.TraverseResult
		Err    error
	}

	WalkInfo struct {
		DiscoverOptionsFn  nav.TraverseOptionFn
		PrincipalOptionsFn nav.TraverseOptionFn
		With               nav.CreateNewRunnerWith
		Resumption         *nav.Resumption
	}

	TraverseInfoTrash struct {
		DiscoverOptionsFn  nav.TraverseOptionFn
		PrincipalOptionsFn nav.TraverseOptionFn
		With               nav.CreateNewRunnerWith
		Resumption         *nav.Resumption
		IsDryRun           bool
	}

	// ClientTraverseInfo represents an entity which needs start a navigation
	// operation.
	ClientTraverseInfo interface {
		// ActiveOptionsFn allows the client to obtain the options func
		// for the current phase.
		ActiveOptionsFn() nav.TraverseOptionFn

		RunWith() nav.CreateNewRunnerWith
		Resumption() *nav.Resumption
		IsDryRun() bool
	}

	// DriverTraverseInfo represents the controller entity that controls
	// the whole traversal consisting of discovery and principal
	// phases.
	DriverTraverseInfo interface {
		ClientTraverseInfo
		// Next allows the driver to indicate switching over to the next
		// phase of traversal. The underlying info object starts off in
		// discovery state, the driver switches it into the principal
		// phase by calling Next.
		Next()
	}

	UserInteraction interface {
		// Decorate allows the interaction to provide a wrapper around the callback.
		// If the interaction does not need it, then it just returns the target. Only
		// the Principal callback is decorated.
		//
		Decorate(target *nav.LabelledTraverseCallback) *nav.LabelledTraverseCallback

		// Performs the full traversal which consists of a discovery navigation followed
		// by the principal navigation.
		Traverse(di DriverTraverseInfo) (*nav.TraverseResult, error)

		// Tick allows the model to be updated, as activity occurs during
		// the traversal.
		//
		Tick(progress *ProgressMsg)
	}

	PresentationOptions struct {
		WithoutRenderer bool
	}
)
