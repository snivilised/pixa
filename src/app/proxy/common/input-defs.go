package common

import (
	"github.com/snivilised/cobrass"
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/store"
)

type (
	RootParameterSet struct { // should contain RootCommandInputs
		Directory  string
		IsSampling bool
		NoFiles    uint
		NoFolders  uint
		Last       bool
	}
)

type InterlaceEnum int

const (
	_ InterlaceEnum = iota
	InterlaceNoneEn
	InterlaceLineEn
	InterlacePlaneEn
	InterlacePartitionEn
	InterlaceJPEGEn
	InterlaceGIFEn
	InterlacePNGEn
)

var InterlaceEnumInfo = assistant.NewEnumInfo(assistant.AcceptableEnumValues[InterlaceEnum]{
	InterlaceNoneEn:      []string{"none", "n"},
	InterlaceLineEn:      []string{"line", "l"},
	InterlacePlaneEn:     []string{"plane", "pl"},
	InterlacePartitionEn: []string{"partition", "pa"},
	InterlaceJPEGEn:      []string{"jpeg", "j"},
	InterlaceGIFEn:       []string{"gif", "g"},
	InterlacePNGEn:       []string{"png", "p"},
})

type SamplingFactorEnum int

const (
	_ SamplingFactorEnum = iota
	SamplingFactor420En
	SamplingFactor2x1En
)

var SamplingFactorEnumInfo = assistant.NewEnumInfo(assistant.AcceptableEnumValues[SamplingFactorEnum]{
	SamplingFactor420En: []string{"4:2:0", "420", "4"},
	SamplingFactor2x1En: []string{"2x1", "21", "2"},
})

// ThirdPartySet represents flags that are only of use to the third party application
// being invoked (ie magick). These flags are of no significance to pixa, but we have
// to define them explicitly, because of a deficiency in cobra in the way it handles
// third party args. The convention in command line interfaces is that the double dash
// delineates arguments for third parties, and cobra does support this, but what it does
// not support is to extract and provide some way to access those args. Pixa needs this
// functionality so it can pass them onto magick. As it stands, we can't access those
// args (after the --), so we have to define them explicitly, then pass them on. This
// is less than desirable, because magick has a vast flag set, which in theory would
// mean re-implementing them all on pixa. We only define the ones relevant to
// compressing images.
type ThirdPartySet struct {
	GaussianBlur     float32
	SamplingFactorEn assistant.EnumValue[SamplingFactorEnum]
	InterlaceEn      assistant.EnumValue[InterlaceEnum]
	Strip            bool
	Quality          int
	// Auxiliary
	//
	LongChangedCL cobrass.ThirdPartyCommandLine
	KnownBy       cobrass.KnownByCollection
}

// [blur]
// magick source.jpg -strip -interlace Plane -gaussian-blur 0.05 -quality 85% result.jpg
// [sampler]
// magick source.jpg -strip -interlace Plane -sampling-factor 4:2:0 -quality 85% result.jpg
// [vanilla]
// magick source.jpg -strip -interlace Plane -quality 85% result.jpg

type ShrinkParameterSet struct {
	ThirdPartySet
	//
	OutputPath string
	TrashPath  string
}

type RootCommandInputs struct {
	ParamSet      *assistant.ParamSet[RootParameterSet]
	PreviewFam    *assistant.ParamSet[store.PreviewParameterSet]
	WorkerPoolFam *assistant.ParamSet[store.WorkerPoolParameterSet]
	FoldersFam    *assistant.ParamSet[store.FoldersFilterParameterSet] // !!!
	ProfileFam    *assistant.ParamSet[store.ProfileParameterSet]
	CascadeFam    *assistant.ParamSet[store.CascadeParameterSet]
	SamplingFam   *assistant.ParamSet[store.SamplingParameterSet]
	TextualFam    *assistant.ParamSet[store.TextualInteractionParameterSet]
	Configs       *Configs
}

type ShrinkCommandInputs struct {
	Root     *RootCommandInputs
	ParamSet *assistant.ParamSet[ShrinkParameterSet]
	PolyFam  *assistant.ParamSet[store.PolyFilterParameterSet]
}
