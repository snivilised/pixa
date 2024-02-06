package common

import (
	"fmt"
	"time"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/clif"
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

type (
	Configs struct {
		Profiles ProfilesConfig
		Schemes  SchemesConfig
		Sampler  SamplerConfig
		Advanced AdvancedConfig
		Logging  LoggingConfig
	}

	ConfigInfo struct {
		Name       string
		ConfigType string
		ConfigPath string
		Viper      configuration.ViperConfig
	}
)

type (
	ConfigRunner interface {
		Run() error
		DefaultPath() string
	}

	ProfilesConfig interface {
		Profile(name string) (clif.ChangedFlagsMap, bool)
	}

	ProfilesConfigReader interface {
		Read(viper configuration.ViperConfig) (ProfilesConfig, error)
	}

	SchemeConfig interface {
		Profiles() []string
	}

	SchemesConfig interface {
		Validate(name string, profiles ProfilesConfig) error
		Scheme(name string) (SchemeConfig, bool)
	}

	SchemesConfigReader interface {
		Read(viper configuration.ViperConfig) (SchemesConfig, error)
	}

	SamplerConfig interface {
		NoFiles() uint
		NoFolders() uint
	}

	SamplerConfigReader interface {
		Read(viper configuration.ViperConfig) (SamplerConfig, error)
	}

	ExtensionsConfig interface {
		Suffixes() string
		Transforms() string
		Map() map[string]string
	}

	ExecutableConfig interface {
		Symbol() string
		ProgramTimeout() (duration time.Duration, err error)
		NoRetries() uint
	}

	AdvancedConfig interface {
		AbortOnError() bool
		AdhocLabel() string
		JournalLabel() string
		LegacyLabel() string
		TrashLabel() string
		FakeLabel() string
		Extensions() ExtensionsConfig
		Executable() ExecutableConfig
	}

	AdvancedConfigReader interface {
		Read(viper configuration.ViperConfig) (AdvancedConfig, error)
	}

	LoggingConfig interface {
		Path() string
		MaxSizeInMb() uint
		MaxNoOfBackups() uint
		MaxAgeInDays() uint
		Level() string
		TimeFormat() string
	}

	LoggingConfigReader interface {
		Read(viper configuration.ViperConfig) (LoggingConfig, error)
	}
)
