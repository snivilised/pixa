package proxy

import (
	"strings"

	"github.com/snivilised/pixa/src/cfg"
)

func newPathFinder(
	inputs *ShrinkCommandInputs,
	advancedCFG cfg.AdvancedConfig,
	schemesCFG cfg.SchemesConfig,
) *PathFinder {
	extensions := advancedCFG.Extensions()
	finder := &PathFinder{
		Scheme:          inputs.Root.ProfileFam.Native.Scheme,
		ExplicitProfile: inputs.Root.ProfileFam.Native.Profile,
		arity:           1,
		statics: &staticInfo{
			adhoc:  advancedCFG.AdhocLabel(),
			legacy: advancedCFG.LegacyLabel(),
			trash:  advancedCFG.TrashLabel(),
		},
		ext: &extensionTransformation{
			transformers: strings.Split(extensions.Transforms(), ","),
			remap:        extensions.Map(),
		},
	}

	if finder.Scheme != "" {
		schemeCFG, _ := schemesCFG.Scheme(finder.Scheme)
		finder.arity = len(schemeCFG.Profiles())
	}

	if inputs.ParamSet.Native.OutputPath != "" {
		finder.Output = inputs.ParamSet.Native.OutputPath
	} else {
		finder.transparentInput = true
	}

	if inputs.ParamSet.Native.TrashPath != "" {
		finder.Trash = inputs.ParamSet.Native.TrashPath
	}

	journal := advancedCFG.JournalLabel()

	if !strings.HasSuffix(journal, journalExtension) {
		journal += journalExtension
	}

	if !strings.HasPrefix(journal, journalTag) {
		journal = journalTag + journal
	}

	withoutExt := strings.TrimSuffix(journal, journalExtension)
	core := strings.TrimPrefix(withoutExt, journalTag)

	finder.statics.meta = journalMetaInfo{
		core:       core,
		journal:    journal,
		withoutExt: withoutExt,
		extension:  journalExtension,
		tag:        journalTag,
	}

	return finder
}
