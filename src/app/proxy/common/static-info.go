package common

import (
	"fmt"
	"path/filepath"
	"strings"
)

type StaticInfo struct {
	Adhoc      string
	Journal    JournalMetaInfo
	Legacy     string
	Trash      string
	Fake       string
	Supplement string
}

func (i *StaticInfo) JournalLocation(name, parent string) string {
	file := name + i.Journal.Actual
	journalFile := filepath.Join(parent, file)

	return journalFile
}

func (i *StaticInfo) JournalFilterGlob() string {
	return fmt.Sprintf("*%v%v*", i.Journal.Discriminator, i.Journal.Core)
}

func (i *StaticInfo) JournalFilterRegex(sourcePattern, suffixesCSV string) string {
	suffixes := strings.Split(suffixesCSV, ",")

	// we make the regex non case specific and also use a dot to match
	// any character before the suffix. Perhaps we need extendio to define
	// an extended regex filter that has similar suffix functionality to
	// the extended glob
	//
	return fmt.Sprintf("(?i).%v.*(%v)$", sourcePattern, strings.Join(suffixes, "|"))
}

func (i *StaticInfo) FileSupplement(baseFilename, supp string) string {
	return fmt.Sprintf("%v.%v", baseFilename, supp)
}

func (i *StaticInfo) TrashTag() string {
	return fmt.Sprintf("$%v$", i.Trash)
}
