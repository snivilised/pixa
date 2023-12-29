package proxy

import (
	"github.com/snivilised/cobrass/src/clif"
)

const (
	defaultNoFiles          = 3
	defaultNoFolders        = 3
	defaultNoProgramRetries = 2
)

var (
	DefaultProfilesConfig *MsProfilesConfig
	DefaultSamplerConfig  *MsSamplerConfig
	DefaultSchemesConfig  *MsSchemesConfig
	DefaultAdvancedConfig *MsAdvancedConfig
)

func init() {
	// TODO: these defaults are not real defaults; they are just test
	// values that don't mean anything. Update to real useable defaults
	//
	DefaultProfilesConfig = &MsProfilesConfig{
		Profiles: ProfilesConfigMap{
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
		},
	}

	DefaultSchemesConfig = &MsSchemesConfig{
		"blur-sf": MsSchemeConfig{
			Profiles: []string{"blur", "sf"},
		},
		"adaptive-sf": MsSchemeConfig{
			Profiles: []string{"adaptive", "sf"},
		},
		"adaptive-blur": MsSchemeConfig{
			Profiles: []string{"adaptive", "blur"},
		},
	}

	DefaultSamplerConfig = &MsSamplerConfig{
		Files:   defaultNoFiles,
		Folders: defaultNoFolders,
	}

	DefaultAdvancedConfig = &MsAdvancedConfig{
		Abort:            false,
		Timeout:          "10s",
		NoProgramRetries: defaultNoProgramRetries,
		Labels: MsLabelsConfig{
			Adhoc:   "ADHOC",
			Journal: ".journal.txt",
			Legacy:  ".LEGACY",
			Trash:   "TRASH",
		},
	}
}
