package common

import (
	"io/fs"

	"github.com/snivilised/extendio/xfs/nav"
)

const (
	DejaVu           = "$pixa$"
	JournalExtension = ".txt"
	JournalTag       = ".$"
)

type (
	JournalMetaInfo struct {
		Core       string // without any decoration
		Journal    string // used as part of the journal file name
		WithoutExt string
		Extension  string // .txt
		Tag        string // the journal file discriminator (.$)
	}

	PathInfo struct {
		Item    *nav.TraverseItem
		Origin  string
		Scheme  string
		Profile string
		RunStep RunStepInfo
	}

	PathFinder interface {
		Transfer(info *PathInfo) (folder, file string)
		Result(info *PathInfo) (folder, file string)
		TransparentInput() bool
		JournalFullPath(item *nav.TraverseItem) string
		Statics() *StaticInfo
		Scheme() string
	}

	FileManager interface {
		Finder() PathFinder
		Create(path string, overwrite bool) error
		Setup(pi *PathInfo) (destination string, err error)
		Tidy(pi *PathInfo) error
	}

	permissions struct {
		Write fs.FileMode
	}
)

const (
	write = 0o766
)

var Permissions = permissions{
	Write: write,
}
