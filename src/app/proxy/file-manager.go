package proxy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/extendio/xfs/storage"
)

const (
	beezledub = os.FileMode(0o666)
)

// FileManager knows how to translate requests into invocations on the file
// system and nothing else.
type FileManager struct {
	vfs    storage.VirtualFS
	finder *PathFinder
}

// Setup prepares for operation by moving existing file out of the way,
// if applicable.
func (fm *FileManager) Setup(item *nav.TraverseItem) error {
	// https://pkg.go.dev/os#Rename LinkError may result
	//
	// this might not be right. it may be that we want to leave the
	// original alone and create other outputs; in this scenario
	// we don't want to rename/move the source...
	//
	from := &destinationInfo{
		item:        item,
		origin:      item.Parent.Path,
		transparent: true, // might come from a flag
	}

	if folder, file := fm.finder.Destination(from); folder != "" {
		if err := fm.vfs.MkdirAll(folder, beezledub); err != nil {
			return errors.Wrapf(err, "could not create parent setup for '%v'", item.Path)
		}

		destination := filepath.Join(folder, file)

		if !fm.vfs.FileExists(item.Path) {
			return fmt.Errorf("source file: '%v' does not exist", item.Path)
		}

		if fm.vfs.FileExists(destination) {
			return fmt.Errorf("destination file: '%v' already exists", destination)
		}

		if err := fm.vfs.Rename(item.Path, destination); err != nil {
			return errors.Wrapf(err, "could not complete setup for '%v'", item.Path)
		}
	}

	return nil
}

func (fm *FileManager) move(from, to string) error {
	_, _ = from, to

	return nil
}

func (fm *FileManager) delete(target string) error {
	_ = target

	return nil
}

func (fm *FileManager) Tidy() error {
	// invoke deletions
	// delete journal file
	//
	return nil
}
