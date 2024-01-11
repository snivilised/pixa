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

type MsExtensionsConfig struct {
	Suffixes string `mapstructure:"suffixes"`
}

type MsAdvancedConfig struct {
	Abort            bool               `mapstructure:"abort-on-error"`
	Timeout          string             `mapstructure:"program-timeout"`
	NoProgramRetries uint               `mapstructure:"no-program-retries"`
	Labels           MsLabelsConfig     `mapstructure:"labels"`
	Extensions       MsExtensionsConfig `mapstructure:"extensions"`
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

func (cfg *MsAdvancedConfig) Suffixes() string {
	return cfg.Extensions.Suffixes
}

type MsLoggingConfig struct {
	LogPath    string `mapstructure:"log-path"`
	MaxSize    uint   `mapstructure:"max-size"`
	MaxBackups uint   `mapstructure:"max-backups"`
	MaxAge     uint   `mapstructure:"max-age"`
	LogLevel   string `mapstructure:"level"`
	Format     string `mapstructure:"time-format"`
}

func (cfg *MsLoggingConfig) Path() string {
	return cfg.LogPath
}

func (cfg *MsLoggingConfig) MaxSizeInMb() uint {
	return cfg.MaxSize
}

func (cfg *MsLoggingConfig) MaxNoOfBackups() uint {
	return cfg.MaxBackups
}

func (cfg *MsLoggingConfig) MaxAgeInDays() uint {
	return cfg.MaxAge
}

func (cfg *MsLoggingConfig) Level() string {
	return cfg.LogLevel
}

func (cfg *MsLoggingConfig) TimeFormat() string {
	return cfg.Format
}
