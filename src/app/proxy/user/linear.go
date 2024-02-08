package user

import (
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type LinearUser struct {
	baseUser
}

// Decorate allows the interaction to provide a wrapper around the callback.
// If the interaction does not need it, then it just returns nil. Only
// the Primary callback is decorated.
func (u *LinearUser) Decorate(target *nav.LabelledTraverseCallback) *nav.LabelledTraverseCallback {
	return target
}

// Discover navigation represents the pre-parse stage of the navigation
func (u *LinearUser) Discover(optionsFn nav.TraverseOptionFn,
	with nav.CreateNewRunnerWith,
	resumption *nav.Resumption,
) error {
	return u.navigate(optionsFn, with, resumption)
}

// Primary represents the main work stage of the navigation
func (u *LinearUser) Primary(optionsFn nav.TraverseOptionFn,
	with nav.CreateNewRunnerWith,
	resumption *nav.Resumption,
	after ...common.AfterFunc,
) error {
	return u.navigate(optionsFn, with, resumption, after...)
}
