package user

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

type linearUI struct {
	interaction
}

// Decorate allows the interaction to provide a wrapper around the callback.
// If the interaction does not need it, then it just returns the target. Only
// the Principal callback is decorated.
func (ui *linearUI) Decorate(target *nav.LabelledTraverseCallback) *nav.LabelledTraverseCallback {
	return target
}

// Performs the full traversal which consists of a discovery navigation followed
// by the principal navigation.
func (ui *linearUI) Traverse(di common.DriverTraverseInfo) (*nav.TraverseResult, error) {
	// we could simple call Next, then call the principal, but we
	// could change the meaning of next which automatically calls principal
	//
	with := di.RunWith()

	if _, err := ui.navigate(di, clearResumeFromWith(with)); err != nil {
		return nil, errors.Wrap(err, "shrink look-ahead phase failed")
	}

	di.Next()

	return ui.navigate(di, with)
}

// Tick allows the model to be updated, as activity occurs during
// the traversal.
func (ui *linearUI) Tick(msg *common.ProgressMsg) {
	bc := bodyContent{
		source:      msg.Source,
		destination: msg.Destination,
		emoji:       randemoji(),
	}

	fmt.Printf(
		`
	===
%v`,
		bc.view(),
	)
}

func (ui *linearUI) summariseAfter(result *nav.TraverseResult, err error) {
	content := summary(result, err)

	fmt.Printf(`
	===
	%v
`, content)
}
