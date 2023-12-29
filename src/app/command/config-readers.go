package command

import (
	"fmt"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy"
)

type MsProfilesConfigReader struct {
}

func (r *MsProfilesConfigReader) Read(viper configuration.ViperConfig) (proxy.ProfilesConfig, error) {
	// Ideally, the ProfileParameterSet would perform a check against
	// the config, but extendio is not aware of config, so it can't
	// check. Instead, we can check here.
	//
	profilesCFG := &proxy.MsProfilesConfig{
		Profiles: make(proxy.ProfilesConfigMap),
	}

	if raw := viper.Get("profiles"); raw != nil {
		if profiles, ok := raw.(proxy.ProfilesFlagOptionAsAnyPair); ok {
			for profile, pv := range profiles {
				if pair, ok := pv.(proxy.ProfilesFlagOptionAsAnyPair); ok {
					profilesCFG.Profiles[profile] = make(clif.ChangedFlagsMap)

					for flag, optionAsAny := range pair {
						profilesCFG.Profiles[profile][flag] = fmt.Sprint(optionAsAny)
					}
				}
			}
		} else {
			return nil, fmt.Errorf("invalid type for 'profiles'")
		}
	}

	return profilesCFG, nil
}

type MsSchemesConfigReader struct{}

func (r *MsSchemesConfigReader) Read(viper configuration.ViperConfig) (proxy.SchemesConfig, error) {
	var (
		schemesCFG proxy.MsSchemesConfig
	)

	err := viper.UnmarshalKey("schemes", &schemesCFG)

	return schemesCFG, err
}

type MsSamplerConfigReader struct{}

func (r *MsSamplerConfigReader) Read(viper configuration.ViperConfig) (proxy.SamplerConfig, error) {
	var (
		samplerCFG proxy.MsSamplerConfig
	)

	err := viper.UnmarshalKey("sampler", &samplerCFG)

	return &samplerCFG, err
}

type MsAdvancedConfigReader struct{}

func (r *MsAdvancedConfigReader) Read(viper configuration.ViperConfig) (proxy.AdvancedConfig, error) {
	var (
		advancedCFG proxy.MsAdvancedConfig
	)

	err := viper.UnmarshalKey("advanced", &advancedCFG)

	return &advancedCFG, err
}

type ConfigReaders struct {
	Profiles proxy.ProfilesConfigReader
	Schemes  proxy.SchemesConfigReader
	Sampler  proxy.SamplerConfigReader
	Advanced proxy.AdvancedConfigReader
}
