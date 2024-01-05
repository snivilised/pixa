package command

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy"
)

const (
	defaultNoFiles           = 3
	defaultNoFolders         = 3
	defaultNoProgramRetries  = 2
	defaultLogMaxSizeInMb    = 10
	defaultLogMaxNoOfBackups = 3
	defaultLogMaxAgeInDays   = 30
)

type (
	defaultSchemes       map[string]proxy.SchemeConfig
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
	DefaultLoggingConfig  *MsLoggingConfig
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

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(errors.Wrap(err, "could not get home dir"))
	}

	DefaultLoggingConfig = &MsLoggingConfig{
		LogPath:    filepath.Join(userHomeDir, "snivilised", "pixa"),
		MaxSize:    defaultLogMaxSizeInMb,
		MaxBackups: defaultLogMaxNoOfBackups,
		MaxAge:     defaultLogMaxAgeInDays,
	}
}
