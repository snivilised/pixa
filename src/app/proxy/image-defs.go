package proxy

import (
	"github.com/snivilised/cobrass"
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/store"
)

type RootParameterSet struct {
	GeneralParameters
	FilterParameters
	Directory string
	CPU       bool
	Language  string
}

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

type ModeEnum int

const (
	_ ModeEnum = iota
	ModeTidyEn
	ModePreserveEn
)

var ModeEnumInfo = assistant.NewEnumInfo(assistant.AcceptableEnumValues[ModeEnum]{
	ModeTidyEn:     []string{"tidy", "t"},
	ModePreserveEn: []string{"preserve", "p"},
})

type ThirdPartySet struct {
	GaussianBlur     float32
	SamplingFactorEn assistant.EnumValue[SamplingFactorEnum]
	InterlaceEn      assistant.EnumValue[InterlaceEnum]
	Strip            bool
	Quality          int
	// Auxiliary
	//
	Present cobrass.SpecifiedFlagsCollection
	KnownBy cobrass.KnownByCollection
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
	MirrorPath string
	ModeEn     assistant.EnumValue[ModeEnum]
}

type RootCommandInputs struct {
	ParamSet      *assistant.ParamSet[RootParameterSet]
	PreviewFam    *assistant.ParamSet[store.PreviewParameterSet]
	WorkerPoolFam *assistant.ParamSet[store.WorkerPoolParameterSet]
	FoldersFam    *assistant.ParamSet[store.FoldersFilterParameterSet]
	ProfileFam    *assistant.ParamSet[store.ProfileParameterSet]
}

type ShrinkCommandInputs struct {
	RootInputs *RootCommandInputs
	ParamSet   *assistant.ParamSet[ShrinkParameterSet]
	FilesFam   *assistant.ParamSet[store.FilesFilterParameterSet]
}
