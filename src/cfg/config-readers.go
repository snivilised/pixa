package cfg

import (
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/clif"
	"golang.org/x/exp/maps"
)

type MsProfilesConfigReader struct {
}

func (r *MsProfilesConfigReader) Read(viper configuration.ViperConfig) (ProfilesConfig, error) {
	// Ideally, the ProfileParameterSet would perform a check against
	// the config, but extendio is not aware of config, so it can't
	// check. Instead, we can check here.
	//
	profilesCFG := &MsProfilesConfig{
		Profiles: make(ProfilesConfigMap),
	}

	if raw := viper.Get("profiles"); raw != nil {
		if profiles, ok := raw.(ProfilesFlagOptionAsAnyPair); ok {
			for profile, pv := range profiles {
				if pair, ok := pv.(ProfilesFlagOptionAsAnyPair); ok {
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

func (r *MsSchemesConfigReader) Read(viper configuration.ViperConfig) (SchemesConfig, error) {
	var (
		schemesCFG MsSchemesConfig
	)

	err := viper.UnmarshalKey("schemes", &schemesCFG)

	return schemesCFG, err
}

type MsSamplerConfigReader struct{}

func (r *MsSamplerConfigReader) Read(viper configuration.ViperConfig) (SamplerConfig, error) {
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

	permittedPrograms = []string{
		"dummy",
		"magick",
	}
)

type MsAdvancedConfigReader struct{}

func (r *MsAdvancedConfigReader) validateSuffixes(suffixes []string, from string) error {
	var (
		err     error
		invalid = map[string]string{}
	)

	for _, v := range suffixes {
		if !slices.Contains(permittedSuffixes, v) {
			invalid[v] = ""
		}
	}

	if len(invalid) > 0 {
		keys := maps.Keys(invalid)
		err = fmt.Errorf("invalid formats found (%v): '%v'", from, strings.Join(keys, ","))
	}

	return err
}

func (r *MsAdvancedConfigReader) validateProgramName(name string) error {
	var (
		err error
	)

	if !slices.Contains(permittedPrograms, name) {
		err = fmt.Errorf("invalid program name found: '%v'", name)
	}

	return err
}

func (r *MsAdvancedConfigReader) Read(viper configuration.ViperConfig) (AdvancedConfig, error) {
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

	if err == nil {
		err = r.validateProgramName(advancedCFG.ExecutableCFG.ProgramName)
	}

	return &advancedCFG, err
}

type MsLoggingConfigReader struct{}

func (r *MsLoggingConfigReader) Read(viper configuration.ViperConfig) (LoggingConfig, error) {
	var (
		loggingCFG MsLoggingConfig
	)

	err := viper.UnmarshalKey("logging", &loggingCFG)

	return &loggingCFG, err
}

type ConfigReaders struct {
	Profiles ProfilesConfigReader
	Schemes  SchemesConfigReader
	Sampler  SamplerConfigReader
	Advanced AdvancedConfigReader
	Logging  LoggingConfigReader
}
