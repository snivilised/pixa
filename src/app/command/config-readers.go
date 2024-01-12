package command

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/clif"
	"github.com/snivilised/pixa/src/app/proxy"
	"golang.org/x/exp/maps"
)

type MsProfilesConfigReader struct {
}

func (r *MsProfilesConfigReader) Read(viper configuration.ViperConfig) (proxy.ProfilesConfig, error) {
	// Ideally, the ProfileParameterSet would perform a check against
	// the config, but extendio is not aware of config, so it can't
	// check. Instead, we can check here.
	//
	profilesCFG := &MsProfilesConfig{
		Profiles: make(proxy.ProfilesConfigMap),
	}

	if raw := viper.Get("profiles"); raw != nil {
		if profiles, ok := raw.(proxy.ProfilesFlagOptionAsAnyPair); ok {
			for profile, pv := range profiles {
				if pair, ok := pv.(proxy.ProfilesFlagOptionAsAnyPair); ok {
					profilesCFG.Profiles[profile] = make(clif.ChangedFlagsMap)

					for flag, optionAsAny := range pair {
						profilesCFG.Profiles[profile][flag] = fmt.Sprint(optionAsAny)
					}
				}
			}
		} else {
			return nil, fmt.Errorf("invalid type for 'profiles'")
		}
	}

	return profilesCFG, nil
}

type MsSchemesConfigReader struct{}

func (r *MsSchemesConfigReader) Read(viper configuration.ViperConfig) (proxy.SchemesConfig, error) {
	var (
		schemesCFG MsSchemesConfig
	)

	err := viper.UnmarshalKey("schemes", &schemesCFG)

	return schemesCFG, err
}

type MsSamplerConfigReader struct{}

func (r *MsSamplerConfigReader) Read(viper configuration.ViperConfig) (proxy.SamplerConfig, error) {
	var (
		samplerCFG MsSamplerConfig
	)

	err := viper.UnmarshalKey("sampler", &samplerCFG)

	return &samplerCFG, err
}

var (
	// reference: https://fileinfo.com/software/imagemagick/imagemagick
	permittedSuffixes = []string{
		"apng", "arw", "avif",
		"bmp", "bpg", "brf",
		"cal", "cals", "cin", "cr2", "crw", "cube", "cur", "cut",
		"dcm", "dcx", "dcr", "dcx",
		"dds", "dib", "dicom", "djvu", "dng", "dot", "dpx",
		"emf", "eps", "exr",
		"fax", "ff", "fits", "flif", "fpx",
		"gif",
		"heic", "hpgl", "hrz",
		"ico",
		"j2c", "j2k", "jbig", "jng", "jp2", "jpc", "jpeg", "jpg", "jxl", "jxr",
		"miff", "mng", "mpo", "mvg",
		"nef",
		"ora", "orf", "otb",
		"pam", "pbm", "pcx", "pict", "pix", "png",
		"tiff", "ttf",
		"webp", "wdp", "wmf",
		"xcf", "xpm", "xwd",
		"yuv",
	}
)

type MsAdvancedConfigReader struct{}

func (r *MsAdvancedConfigReader) validateSuffixes(suffixes []string, from string) error {
	var (
		err     error
		invalid = map[string]string{}
	)

	for _, v := range suffixes {
		invalid[v] = ""
	}

	if len(invalid) > 0 {
		err = fmt.Errorf("invalid formats found (%v): '%v'", from, invalid)
	}

	return err
}

func (r *MsAdvancedConfigReader) Read(viper configuration.ViperConfig) (proxy.AdvancedConfig, error) {
	var (
		advancedCFG MsAdvancedConfig
	)

	err := viper.UnmarshalKey("advanced", &advancedCFG)
	keys := maps.Keys(advancedCFG.ExtensionsCFG.Remap)
	values := maps.Values(advancedCFG.ExtensionsCFG.Remap)

	if err == nil {
		err = r.validateSuffixes(keys, "extensions.map/keys")
	}

	if err == nil {
		err = r.validateSuffixes(values, "extensions.map/values")
	}

	if err == nil {
		suffixes := strings.Split(advancedCFG.ExtensionsCFG.Suffixes(), ",")
		suffixes = lo.Map(suffixes, func(s string, _ int) string {
			return strings.TrimSpace(s)
		})

		err = r.validateSuffixes(suffixes, "extensions.suffixes")
	}

	return &advancedCFG, err
}

type MsLoggingConfigReader struct{}

func (r *MsLoggingConfigReader) Read(viper configuration.ViperConfig) (proxy.LoggingConfig, error) {
	var (
		loggingCFG MsLoggingConfig
	)

	err := viper.UnmarshalKey("logging", &loggingCFG)

	return &loggingCFG, err
}

type ConfigReaders struct {
	Profiles proxy.ProfilesConfigReader
	Schemes  proxy.SchemesConfigReader
	Sampler  proxy.SamplerConfigReader
	Advanced proxy.AdvancedConfigReader
	Logging  proxy.LoggingConfigReader
}
