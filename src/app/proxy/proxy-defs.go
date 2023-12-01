package proxy

import (
	"github.com/snivilised/extendio/xfs/nav"
)

type RunnerTypeEnum uint

const (
	_ RunnerTypeEnum = iota
	RunnerTypeFullEn
	RunnerTypeSamplerEn
)

type SharedRunnerInfo struct {
	Type        RunnerTypeEnum
	Options     *nav.TraverseOptions
	program     Executor
	profiles    ProfilesConfig
	sampler     SamplerConfig
	Inputs      *ShrinkCommandInputs
	finder      *PathFinder
	fileManager *FileManager
}

// ItemController
type ItemRunner interface {
	OnNewShrinkItem(item *nav.TraverseItem,
		positional []string,
	) error
	Reset()
}
