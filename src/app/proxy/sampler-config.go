package proxy

import (
	"fmt"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/clif"
)

//go:generate mockgen -destination src/app/mocks/mocks-config.go -package mocks -source src/app/proxy/sampler-config.go

type (
	MsSchemeConfig struct {
		Profiles []string `mapstructure:"profiles"`
	}

	MsSamplerSchemesConfig map[string]MsSchemeConfig

	ProfilesConfig interface {
		Profile(name string) (clif.ChangedFlagsMap, bool)
	}

	ProfilesConfigReader interface {
		Read(configuration.ViperConfig) (ProfilesConfig, error)
	}

	SamplerConfig interface {
		Validate(name string, profiles ProfilesConfig) error
		Scheme(name string) (MsSchemeConfig, bool)
		NoFiles() uint
		NoFolders() uint
	}

	SamplerConfigReader interface {
		Read(configuration.ViperConfig) (SamplerConfig, error)
	}

	ConfigReaders struct {
		Profiles ProfilesConfigReader
		Sampler  SamplerConfigReader
	}
)

type MsProfilesConfig struct {
	Profiles ProfilesConfigMap
}

func (cfg MsProfilesConfig) Profile(name string) (clif.ChangedFlagsMap, bool) {
	profile, found := cfg.Profiles[name]

	return profile, found
}

type MsSamplerConfig struct {
	Files   uint                   `mapstructure:"files"`
	Folders uint                   `mapstructure:"folders"`
	Schemes MsSamplerSchemesConfig `mapstructure:"schemes"`
}

func (cfg *MsSamplerConfig) Validate(name string, profiles ProfilesConfig) error {
	if name == "" {
		return nil
	}

	var (
		found  bool
		scheme MsSchemeConfig
	)

	if scheme, found = cfg.Schemes[name]; !found {
		return fmt.Errorf("scheme: '%v' not found in config", name)
	}

	for _, p := range scheme.Profiles {
		if _, found := profiles.Profile(p); !found {
			return fmt.Errorf("profile(referenced by scheme: '%v'): '%v' not found in config",
				name, p,
			)
		}
	}

	return nil
}

func (cfg *MsSamplerConfig) Scheme(name string) (MsSchemeConfig, bool) {
	config, found := cfg.Schemes[name]

	return config, found
}

func (cfg *MsSamplerConfig) NoFiles() uint {
	return cfg.Files
}

func (cfg *MsSamplerConfig) NoFolders() uint {
	return cfg.Folders
}

type MsProfilesConfigReader struct {
}

func (r *MsProfilesConfigReader) Read(viper configuration.ViperConfig) (ProfilesConfig, error) {
	// Ideally, the ProfileParameterSet would perform a check against
	// the config, but extendio is not aware of config, so it can't
	// check. Instead, we can check here.
	//
	profilesCFG := &MsProfilesConfig{
		Profiles: make(ProfilesConfigMap),
	}

	if raw := viper.Get("profiles"); raw != nil {
		if profiles, ok := raw.(ProfilesFlagOptionAsAnyPair); ok {
			for profile, pv := range profiles {
				if pair, ok := pv.(ProfilesFlagOptionAsAnyPair); ok {
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

type MsSamplerConfigReader struct{}

func (r *MsSamplerConfigReader) Read(viper configuration.ViperConfig) (SamplerConfig, error) {
	var (
		samplerCFG MsSamplerConfig
	)

	err := viper.UnmarshalKey("sampler", &samplerCFG)

	return &samplerCFG, err
}
