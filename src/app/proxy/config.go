package proxy

import (
	"time"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/clif"
)

//go:generate mockgen -destination ../mocks/mocks-config.go -package mocks -source config.go

type (
	ProfilesConfig interface {
		Profile(name string) (clif.ChangedFlagsMap, bool)
	}

	ProfilesConfigReader interface {
		Read(configuration.ViperConfig) (ProfilesConfig, error)
	}

	SchemeConfig interface {
		Profiles() []string
	}

	SchemesConfig interface {
		Validate(name string, profiles ProfilesConfig) error
		Scheme(name string) (SchemeConfig, bool)
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
