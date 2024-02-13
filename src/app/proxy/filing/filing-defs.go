package filing

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/snivilised/pixa/src/app/proxy/common"
)

func FilenameWithoutExtension(name string) string {
	return strings.TrimSuffix(name, path.Ext(name))
}

func SupplementFilename(name, supp string, statics *common.StaticInfo) string {
	withoutExt := FilenameWithoutExtension(name)

	return statics.FileSupplement(withoutExt, supp) + path.Ext(name)
}

func SupplementFolder(directory, supp string) string {
	return filepath.Join(directory, supp)
}
