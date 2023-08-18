package magick

import (
	"os"
	"path/filepath"

	"github.com/snivilised/extendio/xfs/nav"
)

type EntryBase struct {
	session nav.TraverseSession
}

func (e *EntryBase) ConfigureOptions(o *nav.TraverseOptions) {
	o.Store.FilterDefs = &nav.FilterDefinitions{
		Children: nav.CompoundFilterDef{
			Type:        nav.FilterTypeRegexEn,
			Description: "Image types supported by pixa",
			Pattern:     "\\.(jpe?g|png|gif)$",
		},
	}
}

func ResolvePath(path string) string {
	result := path

	if result[0] == '~' {
		if h, err := os.UserHomeDir(); err == nil {
			result = filepath.Join(h, result[1:])
		}
	} else {
		if absolute, absErr := filepath.Abs(path); absErr == nil {
			result = absolute
		}
	}

	return result
}
