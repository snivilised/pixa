package common

import "github.com/snivilised/extendio/xfs/nav"

type (
	SessionControllerInfo struct {
		Agent       ExecutionAgent
		Inputs      *ShrinkCommandInputs
		FileManager FileManager
		Interaction UserInteraction
	}

	PrivateControllerInfo struct {
		Destination string
		Pi          PathInfo
	}
	RunStepInfo struct {
		Source string
	}

	Step interface {
		Run(pi *PathInfo) error
	}

	// ItemController
	ItemController interface {
		OnNewShrinkItem(item *nav.TraverseItem) error
		Reset()
	}

	// Sequence
	Sequence []Step
)
