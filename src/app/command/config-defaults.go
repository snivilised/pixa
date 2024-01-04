package command

import (
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy"
)

const (
	defaultNoFiles          = 3
	defaultNoFolders        = 3
	defaultNoProgramRetries = 2
)

type (
	defaultSchemes       map[string]proxy.SchemeConfig // should be the proxy interface
	defaultSchemesConfig struct {
		schemes defaultSchemes
	}

	defaultSchemeConfig struct {
		profiles []string
	}
)

func (cfg defaultSchemeConfig) Profiles() []string {
	return cfg.profiles
}

var (
	DefaultProfilesConfig *MsProfilesConfig
	DefaultSamplerConfig  *MsSamplerConfig
	DefaultSchemesConfig  *defaultSchemesConfig
	DefaultAdvancedConfig *MsAdvancedConfig
)

func init() {
	// TODO: these defaults are not real defaults; they are just test
	// values that don't mean anything. Update to real useable defaults
	//
	DefaultProfilesConfig = &MsProfilesConfig{
		Profiles: proxy.ProfilesConfigMap{
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

	// tbd: repatriate MsSchemesConfig
	DefaultSchemesConfig = &defaultSchemesConfig{
		schemes: defaultSchemes{
			"blur-sf": &defaultSchemeConfig{
				profiles: []string{"blur", "sf"},
			},
			"adaptive-sf": &defaultSchemeConfig{
				profiles: []string{"adaptive", "sf"},
			},
			"adaptive-blur": &defaultSchemeConfig{
				profiles: []string{"adaptive", "blur"},
			},
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
