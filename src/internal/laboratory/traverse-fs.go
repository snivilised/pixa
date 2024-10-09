package lab

import (
	"io/fs"
	"os"
	"strings"
	"testing/fstest"

	"github.com/samber/lo"
	"github.com/snivilised/traverse/locale"
)

var (
	perms = struct {
		File fs.FileMode
		Dir  fs.FileMode
	}{File: 0o666, Dir: 0o777} //nolint:mnd // ok
)

type testMapFile struct {
	f fstest.MapFile
}

type TestTraverseFS struct {
	fstest.MapFS
}

func (f *TestTraverseFS) FileExists(name string) bool {
	if mapFile, found := f.MapFS[name]; found && !mapFile.Mode.IsDir() {
		return true
	}

	return false
}

func (f *TestTraverseFS) DirectoryExists(name string) bool {
	if mapFile, found := f.MapFS[name]; found && mapFile.Mode.IsDir() {
		return true
	}

	return false
}

func (f *TestTraverseFS) Create(name string) (*os.File, error) {
	if _, err := f.Stat(name); err == nil {
		return nil, fs.ErrExist
	}

	file := &fstest.MapFile{
		Mode: perms.File,
	}

	f.MapFS[name] = file
	dummy := &os.File{}
	return dummy, nil
}

func (f *TestTraverseFS) MkDirAll(name string, perm os.FileMode) error {
	if !fs.ValidPath(name) {
		return locale.NewInvalidPathError(name)
	}

	segments := strings.Split(name, "/")

	_ = lo.Reduce(segments,
		func(acc []string, s string, _ int) []string {
			acc = append(acc, s)
			path := strings.Join(acc, "/")

			if _, found := f.MapFS[path]; !found {
				f.MapFS[path] = &fstest.MapFile{
					Mode: perm | os.ModeDir,
				}
			}

			return acc
		}, []string{},
	)

	return nil
}

func (f *TestTraverseFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	if _, err := f.Stat(name); err == nil {
		return fs.ErrExist
	}

	f.MapFS[name] = &fstest.MapFile{
		Data: data,
		Mode: perm,
	}

	return nil
}
