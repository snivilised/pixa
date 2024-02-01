package helpers

import (
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/cfg"
)

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

	BackyardWorldsPlanet9Scan02 []string

	ProfilesConfigData cfg.ProfilesConfigMap
	SchemesConfigData  cfg.SchemesConfig
	SamplerConfigData  cfg.SamplerConfig
	AdvancedConfigData cfg.AdvancedConfig
	LoggingConfigData  cfg.LoggingConfig
)

func init() {
	// ✅ Keep this up to date with "nasa-scientist-index.xml"
	//
	BackyardWorldsPlanet9Scan01First2 = []string{
		"01_Backyard-Worlds-Planet-9_s01.jpeg",
		"02_Backyard-Worlds-Planet-9_s01.JPG",
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

	BackyardWorldsPlanet9Scan02 = []string{
		"Backyard-Worlds-Planet-9_s02.jpg",
		"Backyard-Worlds-Planet-9_s02.png",
	}

	ProfilesConfigData = cfg.ProfilesConfigMap{
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

	SchemesConfigData = &cfg.MsSchemesConfig{
		"blur-sf": &cfg.MsSchemeConfig{
			ProfilesData: []string{"blur", "sf"},
		},
		"adaptive-sf": &cfg.MsSchemeConfig{
			ProfilesData: []string{"adaptive", "sf"},
		},
		"adaptive-blur": &cfg.MsSchemeConfig{
			ProfilesData: []string{"adaptive", "blur"},
		},
		"singleton": &cfg.MsSchemeConfig{
			ProfilesData: []string{"adaptive"},
		},
	}

	SamplerConfigData = &cfg.MsSamplerConfig{
		Files:   noSampleFiles,
		Folders: noSampleFolders,
	}

	AdvancedConfigData = &cfg.MsAdvancedConfig{
		Abort: false,
		LabelsCFG: cfg.MsLabelsConfig{
			Adhoc:   "ADHOC",
			Journal: "journal",
			Legacy:  ".LEGACY",
			Trash:   "TRASH",
			Fake:    ".FAKE",
		},
		ExtensionsCFG: cfg.MsExtensionsConfig{
			FileSuffixes:  "jpg,jpeg,png",
			TransformsCSV: "lower",
			Remap: map[string]string{
				"jpeg": "jpg",
			},
		},
		ExecutableCFG: cfg.MsExecutableConfig{
			ProgramName:      "fake",
			Timeout:          "10s",
			NoProgramRetries: noRetries,
		},
	}

	LoggingConfigData = &cfg.MsLoggingConfig{
		LogPath:    "",
		MaxSize:    maxLogSizeInMb,
		MaxBackups: maxLogBackups,
		MaxAge:     maxLogAgeInDays,
	}
}
