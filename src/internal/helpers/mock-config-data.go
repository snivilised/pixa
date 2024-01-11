package helpers

import (
	"fmt"
	"time"

	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy"
)

// need to re-think this mock data as this is currently sub-optimal
// because of the unintended duplication of functionality

const (
	noSampleFiles   = 2
	noSampleFolders = 1
	noRetries       = 2

	maxLogSizeInMb  = 10
	maxLogBackups   = 3
	maxLogAgeInDays = 30
)

var (
	BackyardWorldsPlanet9Scan01First2 []string
	BackyardWorldsPlanet9Scan01First4 []string
	BackyardWorldsPlanet9Scan01First6 []string

	BackyardWorldsPlanet9Scan01Last4 []string

	ProfilesConfigData proxy.ProfilesConfigMap
	SchemesConfigData  proxy.SchemesConfig
	SamplerConfigData  proxy.SamplerConfig
	AdvancedConfigData proxy.AdvancedConfig
	LoggingConfigData  proxy.LoggingConfig
)

type testProfilesConfig struct {
	profiles proxy.ProfilesConfigMap
}

func (cfg testProfilesConfig) Profile(name string) (clif.ChangedFlagsMap, bool) {
	profile, found := cfg.profiles[name]

	return profile, found
}

type (
	testSchemes       map[string]proxy.SchemeConfig
	testSchemesConfig struct {
		schemes testSchemes
	}

	testSchemeConfig struct {
		profiles []string
	}
)

func (cfg *testSchemeConfig) Profiles() []string {
	return cfg.profiles
}

func (cfg *testSchemesConfig) Validate(name string, profiles proxy.ProfilesConfig) error {
	if name == "" {
		return nil
	}

	var (
		found  bool
		scheme proxy.SchemeConfig
	)

	if scheme, found = cfg.schemes[name]; !found {
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

func (cfg *testSchemesConfig) Scheme(name string) (proxy.SchemeConfig, bool) {
	config, found := cfg.schemes[name]

	return config, found
}

type testSamplerConfig struct {
	Files   uint
	Folders uint
}

func (cfg *testSamplerConfig) NoFiles() uint {
	return cfg.Files
}

func (cfg *testSamplerConfig) NoFolders() uint {
	return cfg.Folders
}

type testLabelsConfig struct {
	Adhoc   string
	Journal string
	Legacy  string
	Trash   string
}

type testExtensionsConfig struct {
	Suffixes string
}
type testAdvancedConfig struct {
	Abort            bool
	Timeout          string
	NoProgramRetries uint
	Labels           testLabelsConfig
	Extensions       testExtensionsConfig
}

func (cfg *testAdvancedConfig) AbortOnError() bool {
	return cfg.Abort
}

func (cfg *testAdvancedConfig) ProgramTimeout() (duration time.Duration, err error) {
	return time.ParseDuration(cfg.Timeout)
}

func (cfg *testAdvancedConfig) NoRetries() uint {
	return cfg.NoProgramRetries
}

func (cfg *testAdvancedConfig) AdhocLabel() string {
	return cfg.Labels.Adhoc
}

func (cfg *testAdvancedConfig) JournalLabel() string {
	return cfg.Labels.Journal
}

func (cfg *testAdvancedConfig) LegacyLabel() string {
	return cfg.Labels.Legacy
}

func (cfg *testAdvancedConfig) TrashLabel() string {
	return cfg.Labels.Trash
}

func (cfg *testAdvancedConfig) Suffixes() string {
	return cfg.Extensions.Suffixes
}

type testLoggingConfig struct {
	LogPath    string
	MaxSize    uint
	MaxBackups uint
	MaxAge     uint
	LogLevel   string
	Format     string
}

func (cfg *testLoggingConfig) Path() string {
	return cfg.LogPath
}

func (cfg *testLoggingConfig) MaxSizeInMb() uint {
	return cfg.MaxSize
}

func (cfg *testLoggingConfig) MaxNoOfBackups() uint {
	return cfg.MaxBackups
}

func (cfg *testLoggingConfig) MaxAgeInDays() uint {
	return cfg.MaxAge
}

func (cfg *testLoggingConfig) Level() string {
	return cfg.LogLevel
}

func (cfg *testLoggingConfig) TimeFormat() string {
	return cfg.Format
}

func init() {
	BackyardWorldsPlanet9Scan01First2 = []string{
		"01_Backyard-Worlds-Planet-9_s01.jpg",
		"02_Backyard-Worlds-Planet-9_s01.jpg",
	}

	BackyardWorldsPlanet9Scan01First4 = BackyardWorldsPlanet9Scan01First2
	BackyardWorldsPlanet9Scan01First4 = append(
		BackyardWorldsPlanet9Scan01First4,
		[]string{
			"03_Backyard-Worlds-Planet-9_s01.jpg",
			"04_Backyard-Worlds-Planet-9_s01.jpg",
		}...,
	)

	BackyardWorldsPlanet9Scan01First6 = BackyardWorldsPlanet9Scan01First4
	BackyardWorldsPlanet9Scan01First6 = append(
		BackyardWorldsPlanet9Scan01First6,
		[]string{
			"05_Backyard-Worlds-Planet-9_s01.jpg",
			"06_Backyard-Worlds-Planet-9_s01.jpg",
		}...,
	)

	BackyardWorldsPlanet9Scan01Last4 = []string{
		"03_Backyard-Worlds-Planet-9_s01.jpg",
		"04_Backyard-Worlds-Planet-9_s01.jpg",
		"05_Backyard-Worlds-Planet-9_s01.jpg",
		"06_Backyard-Worlds-Planet-9_s01.jpg",
	}

	ProfilesConfigData = proxy.ProfilesConfigMap{
		"blur": clif.ChangedFlagsMap{
			"strip":         "true",
			"interlace":     "plane",
			"gaussian-blur": "0.05",
		},
		"sf": clif.ChangedFlagsMap{
			"dry-run":         "true",
			"strip":           "true",
			"interlace":       "plane",
			"sampling-factor": "4:2:0",
		},
		"adaptive": clif.ChangedFlagsMap{
			"strip":           "true",
			"interlace":       "plane",
			"gaussian-blur":   "0.25",
			"adaptive-resize": "60",
		},
	}

	SchemesConfigData = &testSchemesConfig{
		schemes: testSchemes{
			"blur-sf": &testSchemeConfig{
				profiles: []string{"blur", "sf"},
			},
			"adaptive-sf": &testSchemeConfig{
				profiles: []string{"adaptive", "sf"},
			},
			"adaptive-blur": &testSchemeConfig{
				profiles: []string{"adaptive", "blur"},
			},
			"singleton": &testSchemeConfig{
				profiles: []string{"adaptive"},
			},
		},
	}

	SamplerConfigData = &testSamplerConfig{
		Files:   noSampleFiles,
		Folders: noSampleFolders,
	}

	AdvancedConfigData = &testAdvancedConfig{
		Abort:            false,
		Timeout:          "10s",
		NoProgramRetries: noRetries,
		Labels: testLabelsConfig{
			Adhoc:   "ADHOC",
			Journal: ".journal.txt",
			Legacy:  ".LEGACY",
			Trash:   "TRASH",
		},
		Extensions: testExtensionsConfig{
			Suffixes: "jpg,jpeg,png",
		},
	}

	LoggingConfigData = &testLoggingConfig{
		LogPath:    "",
		MaxSize:    maxLogSizeInMb,
		MaxBackups: maxLogBackups,
		MaxAge:     maxLogAgeInDays,
	}
}
