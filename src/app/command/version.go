package command

import (
	_ "embed"
	"strings"
)

// TODO: The version.txt should be updated in ci to contain the version
// number associated with the applied tag. Currently not yet defined in this
// template.
var (
	Version = strings.TrimSpace(version)
	//go:embed version.txt
	version string
)
