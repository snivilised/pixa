package cfg

import (
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

// 📚 see https://sagikazarmark.hu/blog/decoding-custom-formats-with-viper/
// for advanced configuration reading using mapstructure
// and viper: https://github.com/spf13/viper

type (
	ProfilesFlagOptionAsAnyPair = map[string]any
	ProfilesConfigMap           map[string]clif.ChangedFlagsMap
	SchemesConfigMap            map[string][]string
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

type MsSchemesConfig map[string]common.SchemeConfig

func (c MsSchemesConfig) Validate(name string, profiles common.ProfilesConfig) error {
	if name == "" {
		return nil
	}

	var (
		found  bool
		scheme common.SchemeConfig
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

func (c MsSchemesConfig) Scheme(name string) (common.SchemeConfig, bool) {
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
	Fake    string `mapstructure:"fake"`
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

type MsExecutableConfig struct {
	ProgramName      string `mapstructure:"program-name"`
	Timeout          string `mapstructure:"timeout"`
	NoProgramRetries uint   `mapstructure:"no-retries"`
}

func (c *MsExecutableConfig) Symbol() string {
	return c.ProgramName
}

func (c *MsExecutableConfig) ProgramTimeout() (duration time.Duration, err error) {
	return time.ParseDuration(c.Timeout)
}

func (c *MsExecutableConfig) NoRetries() uint {
	return c.NoProgramRetries
}

type MsAdvancedConfig struct {
	Abort         bool               `mapstructure:"abort-on-error"`
	LabelsCFG     MsLabelsConfig     `mapstructure:"labels"`
	ExtensionsCFG MsExtensionsConfig `mapstructure:"extensions"`
	ExecutableCFG MsExecutableConfig `mapstructure:"executable"`
}

func (c *MsAdvancedConfig) AbortOnError() bool {
	return c.Abort
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

func (c *MsAdvancedConfig) FakeLabel() string {
	return c.LabelsCFG.Fake
}

func (c *MsAdvancedConfig) Extensions() common.ExtensionsConfig {
	return &c.ExtensionsCFG
}

func (c *MsAdvancedConfig) Executable() common.ExecutableConfig {
	return &c.ExecutableCFG
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

type MsMasterConfig struct {
	Profiles    ProfilesConfigMap   `mapstructure:"profiles"`
	Schemes     SchemesConfigMap    `mapstructure:"schemes"`
	Sampler     MsSamplerConfig     `mapstructure:"sampler"`
	Interaction MsInteractionConfig `mapstructure:"interaction"`
	Advanced    MsAdvancedConfig    `mapstructure:"advanced"`
	Logging     MsLoggingConfig     `mapstructure:"logging"`
}

func (c *MsMasterConfig) Read(vc configuration.ViperConfig) (*common.Configs, error) {
	if err := vc.Unmarshal(c); err != nil {
		return nil, err
	}

	schemes := make(MsSchemesConfig)

	for k, v := range c.Schemes {
		schemes[k] = &MsSchemeConfig{
			ProfilesData: v,
		}
	}

	configs := &common.Configs{
		Profiles: MsProfilesConfig{
			Profiles: c.Profiles,
		},
		Schemes:     schemes,
		Sampler:     &c.Sampler,
		Interaction: &c.Interaction,
		Advanced:    &c.Advanced,
		Logging:     &c.Logging,
	}

	return configs, c.validate(configs)
}

func (c *MsMasterConfig) validate(configs *common.Configs) error {
	extensions := configs.Advanced.Extensions()
	mappings := extensions.Map()
	keys := lo.Keys(mappings)
	values := lo.Values(mappings)

	// In theory, we could could check that every scheme is valid,
	// ie it only contains profiles that have been defined. However,
	// if the current invocation does not refer to a scheme, then we
	// are putting up a barrier unnecessarily. The command is best
	// placed to perform that validation as it has access to the
	// inputs and can act accordingly.

	// extensions
	//
	if err := validateSuffixes(keys, "extensions.map/keys"); err != nil {
		return err
	}

	if err := validateSuffixes(values, "extensions.map/values"); err != nil {
		return err
	}

	suffixes := strings.Split(extensions.Suffixes(), ",")
	suffixes = lo.Map(suffixes, func(s string, _ int) string {
		return strings.TrimSpace(s)
	})

	if err := validateSuffixes(suffixes, "extensions.suffixes"); err != nil {
		return err
	}

	// executable
	//
	executable := configs.Advanced.Executable()

	return validateProgramName(executable.Symbol())
}
