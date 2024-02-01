package common

import (
	"time"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/clif"
)

type (
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
