package cfg

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/exp/maps"
)

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

func validateSuffixes(suffixes []string, from string) error {
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

func validateProgramName(name string) error {
	var (
		err error
	)

	if !slices.Contains(permittedPrograms, name) {
		err = fmt.Errorf("invalid program name found: '%v'", name)
	}

	return err
}
