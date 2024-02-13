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
		Profiles    ProfilesConfig
		Schemes     SchemesConfig
		Sampler     SamplerConfig
		Interaction InteractionConfig
		Advanced    AdvancedConfig
		Logging     LoggingConfig
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

	SchemeConfig interface {
		Profiles() []string
	}

	SchemesConfig interface {
		Validate(name string, profiles ProfilesConfig) error
		Scheme(name string) (SchemeConfig, bool)
	}

	SamplerConfig interface {
		NoFiles() uint
		NoFolders() uint
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

	TuiConfig interface {
		PerItemDelay() time.Duration
	}

	InteractionConfig interface {
		TuiConfig() TuiConfig
	}

	AdvancedConfig interface {
		AbortOnError() bool
		AdhocLabel() string
		JournalLabel() string
		LegacyLabel() string
		TrashLabel() string
		FakeLabel() string
		SupplementLabel() string
		Extensions() ExtensionsConfig
		Executable() ExecutableConfig
	}

	LoggingConfig interface {
		Path() string
		MaxSizeInMb() uint
		MaxNoOfBackups() uint
		MaxAgeInDays() uint
		Level() string
		TimeFormat() string
	}
)
