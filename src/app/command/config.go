package command

import (
	"fmt"
	"time"

	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy"
)

type MsProfilesConfig struct {
	Profiles proxy.ProfilesConfigMap
}

func (cfg MsProfilesConfig) Profile(name string) (clif.ChangedFlagsMap, bool) {
	profile, found := cfg.Profiles[name]

	return profile, found
}

type (
	MsSchemeConfig struct {
		Profiles []string `mapstructure:"profiles"`
	}

	MsSchemesConfig map[string]proxy.SchemeConfig
)

func (cfg MsSchemesConfig) Validate(name string, profiles proxy.ProfilesConfig) error {
	if name == "" {
		return nil
	}

	var (
		found  bool
		scheme proxy.SchemeConfig
	)

	if scheme, found = cfg[name]; !found {
		return fmt.Errorf("scheme: '%v' not found in config", name)
	}

	for _, p := range scheme.Profiles() {
		if _, found := profiles.Profile(p); !found {
			return fmt.Errorf("profile(referenced by scheme: '%v'): '%v' not found in config",
				name, p,
			)
		}
	}

	return nil
}

func (cfg MsSchemesConfig) Scheme(name string) (proxy.SchemeConfig, bool) {
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
