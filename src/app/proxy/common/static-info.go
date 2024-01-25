package common

import (
	"fmt"
	"path/filepath"
	"strings"
)

type StaticInfo struct {
	Adhoc  string
	Meta   JournalMetaInfo
	Legacy string
	Trash  string
}

func (i *StaticInfo) JournalLocation(name, parent string) string {
	file := name + i.Meta.Journal
	journalFile := filepath.Join(parent, file)

	return journalFile
}

func (i *StaticInfo) JournalFilterGlob() string {
	return fmt.Sprintf("*%v%v*", i.Meta.Tag, i.Meta.Core)
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
