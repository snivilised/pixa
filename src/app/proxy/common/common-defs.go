package common

import (
	"fmt"

	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/cfg"
)

type (
	Configs struct {
		Profiles cfg.ProfilesConfig
		Schemes  cfg.SchemesConfig
		Sampler  cfg.SamplerConfig
		Advanced cfg.AdvancedConfig
	}
)

type (
	ProfilesFlagOptionAsAnyPair = map[string]any
	ProfilesConfigMap           map[string]clif.ChangedFlagsMap
)

func (pc ProfilesConfigMap) Validate(name string) error {
	if name != "" {
		if _, found := pc[name]; !found {
			return fmt.Errorf("no such profile: '%v'", name)
		}
	}

	return nil
}
