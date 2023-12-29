package proxy

import (
	"fmt"
	"time"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/clif"
)

//go:generate mockgen -destination ../mocks/mocks-config.go -package mocks -source config.go

type (
	MsSchemeConfig struct {
		Profiles []string `mapstructure:"profiles"`
	}

	MsSchemesConfig map[string]MsSchemeConfig

	ProfilesConfig interface {
		Profile(name string) (clif.ChangedFlagsMap, bool)
	}

	ProfilesConfigReader interface {
		Read(configuration.ViperConfig) (ProfilesConfig, error)
	}

	SchemesConfig interface {
		Validate(name string, profiles ProfilesConfig) error
		Scheme(name string) (MsSchemeConfig, bool)
	}

	SchemesConfigReader interface {
		Read(configuration.ViperConfig) (SchemesConfig, error)
	}

	SamplerConfig interface {
		NoFiles() uint
		NoFolders() uint
	}

	SamplerConfigReader interface {
		Read(configuration.ViperConfig) (SamplerConfig, error)
	}

	AdvancedConfig interface {
		AbortOnError() bool
		ProgramTimeout() (duration time.Duration, err error)
		NoRetries() uint
		AdhocLabel() string
		JournalLabel() string
		LegacyLabel() string
		TrashLabel() string
	}

	AdvancedConfigReader interface {
		Read(configuration.ViperConfig) (AdvancedConfig, error)
	}
)

type MsProfilesConfig struct {
	Profiles ProfilesConfigMap
}

func (cfg MsProfilesConfig) Profile(name string) (clif.ChangedFlagsMap, bool) {
	profile, found := cfg.Profiles[name]

	return profile, found
}

func (cfg MsSchemesConfig) Validate(name string, profiles ProfilesConfig) error {
	if name == "" {
		return nil
	}

	var (
		found  bool
		scheme MsSchemeConfig
	)

	if scheme, found = cfg[name]; !found {
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

func (cfg MsSchemesConfig) Scheme(name string) (MsSchemeConfig, bool) {
	config, found := cfg[name]

	return config, found
}

type MsSamplerConfig struct {
	Files   uint `mapstructure:"files"`
	Folders uint `mapstructure:"folders"`
}

func (cfg *MsSamplerConfig) NoFiles() uint {
	return cfg.Files
}

func (cfg *MsSamplerConfig) NoFolders() uint {
	return cfg.Folders
}

type MsLabelsConfig struct {
	Adhoc   string `mapstructure:"adhoc"`
	Journal string `mapstructure:"journal-suffix"`
	Legacy  string `mapstructure:"legacy"`
	Trash   string `mapstructure:"trash"`
}

type MsAdvancedConfig struct {
	Abort            bool           `mapstructure:"abort-on-error"`
	Timeout          string         `mapstructure:"program-timeout"`
	NoProgramRetries uint           `mapstructure:"no-program-retries"`
	Labels           MsLabelsConfig `mapstructure:"labels"`
}

func (cfg *MsAdvancedConfig) AbortOnError() bool {
	return cfg.Abort
}

func (cfg *MsAdvancedConfig) ProgramTimeout() (duration time.Duration, err error) {
	return time.ParseDuration(cfg.Timeout)
}

func (cfg *MsAdvancedConfig) NoRetries() uint {
	return cfg.NoProgramRetries
}

func (cfg *MsAdvancedConfig) AdhocLabel() string {
	return cfg.Labels.Adhoc
}

func (cfg *MsAdvancedConfig) JournalLabel() string {
	return cfg.Labels.Journal
}

func (cfg *MsAdvancedConfig) LegacyLabel() string {
	return cfg.Labels.Legacy
}

func (cfg *MsAdvancedConfig) TrashLabel() string {
	return cfg.Labels.Trash
}
