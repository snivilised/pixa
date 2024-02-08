package user

import (
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type TextualUser struct {
	baseUser
}

// Decorate allows the interaction to provide a wrapper around the callback.
// If the interaction does not need it, then it just returns nil. Only
// the Primary callback is decorated.
func (u *TextualUser) Decorate(target *nav.LabelledTraverseCallback) *nav.LabelledTraverseCallback {
	return &nav.LabelledTraverseCallback{
		Label: "💝💝 Principal Textual Shrink Callback",
		Fn: func(item *nav.TraverseItem) error {
			// todo: do more stuff, probably with the model
			//
			return target.Fn(item)
		},
	}
}

// Discover navigation represents the pre-parse stage of the navigation
func (u *TextualUser) Discover(optionsFn nav.TraverseOptionFn,
	with nav.CreateNewRunnerWith,
	resumption *nav.Resumption,
) error {
	return u.navigate(optionsFn, with, resumption)
}

// Primary represents the main work stage of the navigation
func (u *TextualUser) Primary(optionsFn nav.TraverseOptionFn,
	with nav.CreateNewRunnerWith,
	resumption *nav.Resumption,
	after ...common.AfterFunc,
) error {
	return u.navigate(optionsFn, with, resumption, after...)
}
