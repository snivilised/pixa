package proxy

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/snivilised/extendio/xfs/nav"
	"github.com/snivilised/pixa/src/cfg"
)

type filterSetup struct {
	inputs *ShrinkCommandInputs
	config cfg.AdvancedConfig
}

func (s *filterSetup) getDefs(statics *staticInfo) *nav.FilterDefinitions {
	// the filter we expect the user to provide does not include the file suffix,
	// it only applies to the base name and we define the suffix part of the filter
	// internally.
	//
	var (
		file, folder  *nav.FilterDef
		defs          *nav.FilterDefinitions
		folderDefined = true
		pattern       string
		suffixes      = s.config.Extensions().Suffixes()
	)

	switch {
	case s.inputs.PolyFam.Native.Files != "":
		exclusion := statics.JournalFilterGlob()
		// 📚 The pattern defined uses an exclusion, but this is no longer
		// necessary here in pixa, because we provide a custom ReadDirectory
		// hook which filters out any journal file. So ideally this exclusion
		// should be taken out. However, one of the aim of this project is
		// to demonstrate features and usage of extendio, cobrass and lorax,
		// so this exclusion filtering will remain in place.
		//
		pattern = fmt.Sprintf("%v/%v|%v",
			s.inputs.PolyFam.Native.Files,
			exclusion,
			suffixes,
		)
		file = &nav.FilterDef{
			Type:        nav.FilterTypeExtendedGlobEn,
			Description: fmt.Sprintf("--files(F): '%v'", pattern),
			Pattern:     pattern,
			Scope:       nav.ScopeFileEn,
		}

	case s.inputs.PolyFam.Native.FilesRexEx != "":
		pattern = statics.JournalFilterRegex(s.inputs.PolyFam.Native.FilesRexEx, suffixes)

		file = &nav.FilterDef{
			Type:        nav.FilterTypeRegexEn,
			Description: fmt.Sprintf("--files-rx(X): '%v'", pattern),
			Pattern:     pattern,
			Scope:       nav.ScopeFileEn,
		}

	default:
		exclusion := statics.JournalFilterGlob()
		pattern = fmt.Sprintf("*/%v|%v", exclusion, suffixes)
		file = &nav.FilterDef{
			Type:        nav.FilterTypeExtendedGlobEn,
			Description: fmt.Sprintf("default extended glob filter: '%v'", pattern),
			Pattern:     pattern,
			Scope:       nav.ScopeFileEn,
		}
	}

	switch {
	case s.inputs.Root.FoldersFam.Native.FoldersGlob != "":
		pattern = s.inputs.Root.FoldersFam.Native.FoldersRexEx
		folder = &nav.FilterDef{
			Type:        nav.FilterTypeGlobEn,
			Description: fmt.Sprintf("--folders-gb(Z): '%v'", pattern),
			Pattern:     pattern,
			Scope:       nav.ScopeFolderEn | nav.ScopeLeafEn,
		}

	case s.inputs.Root.FoldersFam.Native.FoldersRexEx != "":
		pattern = s.inputs.Root.FoldersFam.Native.FoldersRexEx
		folder = &nav.FilterDef{
			Type:        nav.FilterTypeRegexEn,
			Description: fmt.Sprintf("--folders-rx(Y): '%v'", pattern),
			Pattern:     pattern,
			Scope:       nav.ScopeFolderEn | nav.ScopeLeafEn,
		}

	default:
		folderDefined = false
	}

	switch {
	case folderDefined:
		defs = &nav.FilterDefinitions{
			Node: nav.FilterDef{
				Type: nav.FilterTypePolyEn,
				Poly: &nav.PolyFilterDef{
					File:   *file,
					Folder: *folder,
				},
			},
		}

	default:
		defs = &nav.FilterDefinitions{
			Node: *file,
		}
	}

	return lo.TernaryF(pattern != "",
		func() *nav.FilterDefinitions {
			return defs
		},
		func() *nav.FilterDefinitions {
			return nil
		},
	)
}
