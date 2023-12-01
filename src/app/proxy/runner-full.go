package proxy

import (
	"github.com/snivilised/cobrass"
	"github.com/snivilised/extendio/xfs/nav"
)

type FullRunner struct {
	baseRunner
}

func (r *FullRunner) OnNewShrinkItem(item *nav.TraverseItem,
	positional []string,
	thirdPartyCL cobrass.ThirdPartyCommandLine,
) error {
	_ = item
	_ = positional
	_ = thirdPartyCL

	return nil
}
