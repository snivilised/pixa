package user

import (
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type walkInfo struct {
	discoverOptionsFn  nav.TraverseOptionFn
	principalOptionsFn nav.TraverseOptionFn
	activeOptionsFn    nav.TraverseOptionFn
	with               nav.CreateNewRunnerWith
	resumption         *nav.Resumption
	dryRun             bool
}

func NewWalkInfo(discoverOptionsFn nav.TraverseOptionFn,
	principalOptionsFn nav.TraverseOptionFn,
	with nav.CreateNewRunnerWith,
	resumption *nav.Resumption,
	dryRun bool,
) common.DriverTraverseInfo {
	return &walkInfo{
		discoverOptionsFn:  discoverOptionsFn,
		principalOptionsFn: principalOptionsFn,
		activeOptionsFn:    discoverOptionsFn,
		with:               with,
		resumption:         resumption,
		dryRun:             dryRun,
	}
}

func (wi *walkInfo) ActiveOptionsFn() nav.TraverseOptionFn {
	return wi.activeOptionsFn
}

func (wi *walkInfo) RunWith() nav.CreateNewRunnerWith {
	return wi.with
}

func (wi *walkInfo) Resumption() *nav.Resumption {
	return wi.resumption
}

func (wi *walkInfo) IsDryRun() bool {
	return wi.dryRun
}

func (wi *walkInfo) Next() {
	wi.activeOptionsFn = wi.principalOptionsFn
}
