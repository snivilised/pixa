package common

import (
	"fmt"

	"github.com/snivilised/cobrass/src/clif"
)

type (
	Configs struct {
		Profiles ProfilesConfig
		Schemes  SchemesConfig
		Sampler  SamplerConfig
		Advanced AdvancedConfig
		Logging  LoggingConfig
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
