package proxy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/snivilised/extendio/xfs/storage"
)

const (
	beezledub        = os.FileMode(0o666)
	perm             = os.FileMode(0o766)
	errorDestination = ""
)

// FileManager knows how to translate requests into invocations on the file
// system and nothing else.
type FileManager struct {
	vfs    storage.VirtualFS
	Finder *PathFinder
}

func (fm *FileManager) Create(path string) error {
	if fm.vfs.FileExists(path) {
		return nil
	}

	file, err := fm.vfs.Create(path)

	if err != nil {
		return err
	}

	defer file.Close()

	return nil
}

// Setup prepares for operation by moving existing file out of the way,
// if applicable. Return the path denoting where the input will be moved to.
func (fm *FileManager) Setup(pi *pathInfo) (destination string, err error) {
	if !fm.Finder.transparentInput {
		// Any result file must not clash with the input file, so the input
		// file must stay in place
		return pi.item.Path, nil
	}

	// https://pkg.go.dev/os#Rename LinkError may result
	//
	// this might not be right. it may be that we want to leave the
	// original alone and create other outputs; in this scenario
	// we don't want to rename/move the source...
	//
	if folder, file := fm.Finder.Transfer(pi); folder != "" {
		if err = fm.vfs.MkdirAll(folder, perm); err != nil {
			return errorDestination, errors.Wrapf(
				err, "could not create parent setup for '%v'", pi.item.Path,
			)
		}

		destination = filepath.Join(folder, file)

		if !fm.vfs.FileExists(pi.item.Path) {
			return errorDestination, fmt.Errorf(
				"source file: '%v' does not exist", pi.item.Path,
			)
		}

		if pi.item.Path != destination {
			if fm.vfs.FileExists(destination) {
				return errorDestination, fmt.Errorf(
					"destination file: '%v' already exists", destination,
				)
			}

			if err := fm.vfs.Rename(pi.item.Path, destination); err != nil {
				return errorDestination, errors.Wrapf(
					err, "could not complete setup for '%v'", pi.item.Path,
				)
			}
		}
	}

	return destination, nil
}

func (fm *FileManager) Tidy(pi *pathInfo) error {
	journalFile := fm.Finder.JournalFullPath(pi.item)

	if !fm.vfs.FileExists(journalFile) {
		return fmt.Errorf("journal file '%v' not found", journalFile)
	}

	return fm.vfs.Remove(journalFile)
}
