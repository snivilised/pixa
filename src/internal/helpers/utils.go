package helpers

import (
	"os"
	"path/filepath"
)

func ResolvePath(path string) string {
	if path == "" {
		return path
	}

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
