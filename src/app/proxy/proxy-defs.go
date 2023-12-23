package proxy

import (
	"github.com/snivilised/extendio/xfs/nav"
)

type ControllerTypeEnum uint

const (
	_ ControllerTypeEnum = iota
	ControllerTypeFullEn
	ControllerTypeSamplerEn
)

type SharedControllerInfo struct {
	Type        ControllerTypeEnum
	Options     *nav.TraverseOptions
	program     Executor
	profiles    ProfilesConfig
	sampler     SamplerConfig
	Inputs      *ShrinkCommandInputs
	finder      *PathFinder
	fileManager *FileManager
}

type localControllerInfo struct {
	destination string
}

// ItemController
type ItemController interface {
	OnNewShrinkItem(item *nav.TraverseItem,
		positional []string,
	) error
	Reset()
}
