package cfg

import (
	"time"

	"github.com/snivilised/pixa/src/app/proxy/common"
)

type MsInteractionConfig struct {
	Tui MsTuiConfig `mapstructure:"tui"`
}

func (c *MsInteractionConfig) TuiConfig() common.TuiConfig {
	return &c.Tui
}

type MsTuiConfig struct {
	Delay string `mapstructure:"per-item-delay"`
}

func (c *MsTuiConfig) PerItemDelay() time.Duration {
	if c.Delay == "" {
		return 0
	}

	duration, err := time.ParseDuration(c.Delay)
	if err != nil {
		duration = time.Second
	}

	return duration
}
