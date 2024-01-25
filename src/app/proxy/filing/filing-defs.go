package filing

import (
	"path"
	"strings"
)

var (
	DejaVu = "$pixa$"
)

func FilenameWithoutExtension(name string) string {
	return strings.TrimSuffix(name, path.Ext(name))
}
