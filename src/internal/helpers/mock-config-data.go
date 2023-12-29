package helpers

import (
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy"
)

const (
	noSampleFiles   = 2
	noSampleFolders = 1
	noRetries       = 2
)

var (
	BackyardWorldsPlanet9Scan01First2 []string
	BackyardWorldsPlanet9Scan01First4 []string
	BackyardWorldsPlanet9Scan01First6 []string

	BackyardWorldsPlanet9Scan01Last4 []string

	ProfilesConfigData proxy.ProfilesConfigMap
	SchemesConfigData  *proxy.MsSchemesConfig
	SamplerConfigData  *proxy.MsSamplerConfig
	AdvancedConfigData *proxy.MsAdvancedConfig
)

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

	SchemesConfigData = &proxy.MsSchemesConfig{
		"blur-sf": proxy.MsSchemeConfig{
			Profiles: []string{"blur", "sf"},
		},
		"adaptive-sf": proxy.MsSchemeConfig{
			Profiles: []string{"adaptive", "sf"},
		},
		"adaptive-blur": proxy.MsSchemeConfig{
			Profiles: []string{"adaptive", "blur"},
		},
		"singleton": proxy.MsSchemeConfig{
			Profiles: []string{"adaptive"},
		},
	}

	SamplerConfigData = &proxy.MsSamplerConfig{
		Files:   noSampleFiles,
		Folders: noSampleFolders,
	}

	AdvancedConfigData = &proxy.MsAdvancedConfig{
		Abort:            false,
		Timeout:          "10s",
		NoProgramRetries: noRetries,
		Labels: proxy.MsLabelsConfig{
			Adhoc:   "ADHOC",
			Journal: ".journal.txt",
			Legacy:  ".LEGACY",
			Trash:   "TRASH",
		},
	}
}
