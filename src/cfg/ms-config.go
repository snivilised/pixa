package cfg

import (
	"fmt"
	"time"

	"github.com/snivilised/cobrass/src/clif"
)

type (
	ProfilesFlagOptionAsAnyPair = map[string]any
	ProfilesConfigMap           map[string]clif.ChangedFlagsMap
)

type MsProfilesConfig struct {
	Profiles ProfilesConfigMap
}

func (c MsProfilesConfig) Profile(name string) (clif.ChangedFlagsMap, bool) {
	profile, found := c.Profiles[name]

	return profile, found
}

type MsSchemeConfig struct {
	ProfilesData []string `mapstructure:"profiles"`
}

func (c *MsSchemeConfig) Profiles() []string {
	return c.ProfilesData
}

type MsSchemesConfig map[string]SchemeConfig

func (c MsSchemesConfig) Validate(name string, profiles ProfilesConfig) error {
	if name == "" {
		return nil
	}

	var (
		found  bool
		scheme SchemeConfig
	)

	if scheme, found = c[name]; !found {
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

func (c MsSchemesConfig) Scheme(name string) (SchemeConfig, bool) {
	config, found := c[name]

	return config, found
}

type MsSamplerConfig struct {
	Files   uint `mapstructure:"files"`
	Folders uint `mapstructure:"folders"`
}

func (c *MsSamplerConfig) NoFiles() uint {
	return c.Files
}

func (c *MsSamplerConfig) NoFolders() uint {
	return c.Folders
}

type MsLabelsConfig struct {
	Adhoc   string `mapstructure:"adhoc"`
	Journal string `mapstructure:"journal-suffix"`
	Legacy  string `mapstructure:"legacy"`
	Trash   string `mapstructure:"trash"`
}

type MsExtensionsConfig struct {
	FileSuffixes  string            `mapstructure:"suffixes-csv"`
	TransformsCSV string            `mapstructure:"transforms-csv"`
	Remap         map[string]string `mapstructure:"map"`
}

func (c *MsExtensionsConfig) Suffixes() string {
	return c.FileSuffixes
}

func (c *MsExtensionsConfig) Transforms() string {
	return c.TransformsCSV
}

func (c *MsExtensionsConfig) Map() map[string]string {
	return c.Remap
}

type MsAdvancedConfig struct {
	Abort            bool               `mapstructure:"abort-on-error"`
	Timeout          string             `mapstructure:"program-timeout"`
	NoProgramRetries uint               `mapstructure:"no-program-retries"`
	LabelsCFG        MsLabelsConfig     `mapstructure:"labels"`
	ExtensionsCFG    MsExtensionsConfig `mapstructure:"extensions"`
}

func (c *MsAdvancedConfig) AbortOnError() bool {
	return c.Abort
}

func (c *MsAdvancedConfig) ProgramTimeout() (duration time.Duration, err error) {
	return time.ParseDuration(c.Timeout)
}

func (c *MsAdvancedConfig) NoRetries() uint {
	return c.NoProgramRetries
}

func (c *MsAdvancedConfig) AdhocLabel() string {
	return c.LabelsCFG.Adhoc
}

func (c *MsAdvancedConfig) JournalLabel() string {
	return c.LabelsCFG.Journal
}

func (c *MsAdvancedConfig) LegacyLabel() string {
	return c.LabelsCFG.Legacy
}

func (c *MsAdvancedConfig) TrashLabel() string {
	return c.LabelsCFG.Trash
}

func (c *MsAdvancedConfig) Extensions() ExtensionsConfig {
	return &c.ExtensionsCFG
}

type MsLoggingConfig struct {
	LogPath    string `mapstructure:"log-path"`
	MaxSize    uint   `mapstructure:"max-size-in-mb"`
	MaxBackups uint   `mapstructure:"max-backups"`
	MaxAge     uint   `mapstructure:"max-age-in-days"`
	LogLevel   string `mapstructure:"level"`
	Format     string `mapstructure:"time-format"`
}

func (c *MsLoggingConfig) Path() string {
	return c.LogPath
}

func (c *MsLoggingConfig) MaxSizeInMb() uint {
	return c.MaxSize
}

func (c *MsLoggingConfig) MaxNoOfBackups() uint {
	return c.MaxBackups
}

func (c *MsLoggingConfig) MaxAgeInDays() uint {
	return c.MaxAge
}

func (c *MsLoggingConfig) Level() string {
	return c.LogLevel
}

func (c *MsLoggingConfig) TimeFormat() string {
	return c.Format
}
