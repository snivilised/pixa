package proxy

import (
	"fmt"
	"path/filepath"
	"strings"
)

type journalMetaInfo struct {
	core       string // without any decoration
	journal    string // used as part of the journal file name
	withoutExt string
	extension  string // .txt
	tag        string // the journal file discriminator (.$)
}

type staticInfo struct {
	adhoc  string
	meta   journalMetaInfo
	legacy string
	trash  string
}

func (i *staticInfo) JournalLocation(name, parent string) string {
	file := name + i.meta.journal
	journalFile := filepath.Join(parent, file)

	return journalFile
}

func (i *staticInfo) JournalFilterGlob() string {
	return fmt.Sprintf("*%v%v*", i.meta.tag, i.meta.core)
}

func (i *staticInfo) JournalFilterRegex(sourcePattern, suffixesCSV string) string {
	suffixes := strings.Split(suffixesCSV, ",")

	// we make the regex non case specific and also use a dot to match
	// any character before the suffix. Perhaps we need extendio to define
	// an extended regex filter that has similar suffix functionality to
	// the extended glob
	//
	return fmt.Sprintf("(?i).%v.*(%v)$", sourcePattern, strings.Join(suffixes, "|"))
}
