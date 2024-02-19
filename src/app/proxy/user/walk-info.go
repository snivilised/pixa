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
	inputs             *common.ShrinkCommandInputs
}

func NewWalkInfo(discoverOptionsFn nav.TraverseOptionFn,
	principalOptionsFn nav.TraverseOptionFn,
	with nav.CreateNewRunnerWith,
	resumption *nav.Resumption,
	inputs *common.ShrinkCommandInputs,
) common.DriverTraverseInfo {
	return &walkInfo{
		discoverOptionsFn:  discoverOptionsFn,
		principalOptionsFn: principalOptionsFn,
		activeOptionsFn:    discoverOptionsFn,
		with:               with,
		resumption:         resumption,
		inputs:             inputs,
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
	return wi.inputs.Root.PreviewFam.Native.DryRun
}

func (wi *walkInfo) Next() {
	wi.activeOptionsFn = wi.principalOptionsFn
}
