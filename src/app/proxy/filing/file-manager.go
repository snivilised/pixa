package filing

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

const (
	beezledub        = os.FileMode(0o666)
	perm             = os.FileMode(0o766)
	errorDestination = ""
)

func NewManager(vfs storage.VirtualFS, finder common.PathFinder, dryRun bool) common.FileManager {
	return &FileManager{
		Vfs:    vfs,
		finder: finder,
		dryRun: dryRun,
	}
}

// FileManager knows how to translate requests into invocations on the file
// system and nothing else.
type FileManager struct {
	Vfs    storage.VirtualFS
	finder common.PathFinder
	dryRun bool
}

func (fm *FileManager) Finder() common.PathFinder {
	return fm.finder
}

func (fm *FileManager) Create(path string, overwrite bool) error {
	if fm.Vfs.FileExists(path) && !overwrite {
		return errors.Wrapf(os.ErrExist, "could not create file at path: '%v'", path)
	}

	file, err := fm.Vfs.Create(path)

	if err != nil {
		return err
	}

	defer file.Close()

	return nil
}

// Setup prepares for operation by moving existing file out of the way,
// if applicable. Return the path denoting where the input will be moved to.
func (fm *FileManager) Setup(pi *common.PathInfo) (destination string, err error) {
	if !fm.finder.TransparentInput() {
		// Any result file must not clash with the input file, so the input
		// file must stay in place
		return pi.Item.Path, nil
	}

	// https://pkg.go.dev/os#Rename LinkError may result
	//
	// this might not be right. it may be that we want to leave the
	// original alone and create other outputs; in this scenario
	// we don't want to rename/move the source...
	//
	if folder, file := fm.finder.Transfer(pi); folder != "" {
		if !fm.dryRun {
			if err = fm.Vfs.MkdirAll(folder, perm); err != nil {
				return errorDestination, errors.Wrapf(
					err, "could not create parent setup for '%v'", pi.Item.Path,
				)
			}
		}

		destination = filepath.Join(folder, file)

		if !fm.dryRun {
			if !fm.Vfs.FileExists(pi.Item.Path) {
				return errorDestination, fmt.Errorf(
					"source file: '%v' does not exist", pi.Item.Path,
				)
			}

			if pi.Item.Path != destination {
				if fm.Vfs.FileExists(destination) {
					return errorDestination, fmt.Errorf(
						"destination file: '%v' already exists", destination,
					)
				}

				if err := fm.Vfs.Rename(pi.Item.Path, destination); err != nil {
					return errorDestination, errors.Wrapf(
						err, "could not complete setup for '%v'", pi.Item.Path,
					)
				}
			}
		}
	}

	return destination, nil
}

func (fm *FileManager) Tidy(pi *common.PathInfo) error {
	if fm.dryRun {
		return nil
	}

	journalFile := fm.finder.JournalFullPath(pi.Item)

	if !fm.Vfs.FileExists(journalFile) {
		return fmt.Errorf("journal file '%v' not found", journalFile)
	}

	return fm.Vfs.Remove(journalFile)
}

// transparent=true should be the default scenario. This means
// that any changes that occur leave the file system in a state
// where nothing appears to have changed except that files have
// been modified, without name changes. This of course doesn't
// include items that end up in TRASH and can be manually deleted
// by the user. The purpose of this is to by default require
// the least amount of post-processing clean-up from the user.
//
// In sampling mode, transparent may mean something different
// because multiple files could be created for each input file.
// So, in this scenario, the original file should stay in-tact
// and the result(s) should be created into the supplementary
// location.
//
// In full  mode, transparent means the input file is moved
// to a trash location. The output takes the name of the original
// file, so that by the end of processing, the resultant file
// takes the place of the source file, leaving the file system
// in a state that was the same before processing occurred.
//
// So what happens in non transparent scenario? The source file
// remains unchanged, so the user has to look at another location
// to get the result. It uses the SHRINK label to create the
// output filename; but note, we only use the SHRINK label in
// scenarios where there is a potential for a filename clash if
// the output file is in the same location as the input file
// because we want to create the least amount of friction as
// possible. This only occurs when in adhoc mode (no profile
// or scheme)
