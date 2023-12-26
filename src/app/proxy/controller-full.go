package proxy

import (
	"github.com/snivilised/cobrass"
	"github.com/snivilised/extendio/xfs/nav"
)

type FullController struct {
	controller
}

func (c *FullController) OnNewShrinkItem(
	item *nav.TraverseItem,
	thirdPartyCL cobrass.ThirdPartyCommandLine,
) error {
	_ = item
	_ = thirdPartyCL

	return nil
}
