package common

import "github.com/snivilised/extendio/xfs/nav"

type (
	SharedControllerInfo struct {
		Agent       ExecutionAgent
		Inputs      *ShrinkCommandInputs
		Finder      PathFinder
		FileManager FileManager
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
