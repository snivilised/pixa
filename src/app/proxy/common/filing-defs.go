package common

import (
	"io/fs"

	"github.com/snivilised/extendio/xfs/nav"
)

type (
	JournalMetaInfo struct {
		Core          string // without any decoration
		Actual        string // used as part of the journal file name
		WithoutExt    string
		Extension     string // .txt
		Discriminator string // the journal/sample file discriminator (.$)
	}

	PathInfo struct {
		Item    *nav.TraverseItem
		Origin  string
		Scheme  string
		Profile string
		RunStep RunStepInfo
		Cuddle  bool
		Output  string
		Trash   string
	}

	PathFinder interface {
		Transfer(info *PathInfo) (folder, file string)
		Result(info *PathInfo) (folder, file string)
		TransparentInput() bool
		JournalFullPath(item *nav.TraverseItem) string
		Statics() *StaticInfo
		Scheme() string
		Observe(o PathFinder) PathFinder
	}

	FileManager interface {
		Finder() PathFinder
		Create(path string, overwrite bool) error
		Setup(pi *PathInfo) (destination string, err error)
		Tidy(pi *PathInfo) error
	}

	permissions struct {
		Write       fs.FileMode
		Faydeaudeau fs.FileMode
		Beezledub   fs.FileMode
	}
)

const (
	write       = 0o766
	faydeaudeau = 0o777
	beezledub   = 0o666
)

var Permissions = permissions{
	Write:       write,
	Faydeaudeau: faydeaudeau,
	Beezledub:   beezledub,
}
