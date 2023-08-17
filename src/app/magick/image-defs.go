package magick

import (
	"github.com/snivilised/cobrass/src/assistant"
)

// CLIENT-TODO: define valid properties on the root parameter set
type RootParameterSet struct {
	GeneralParameters
	FilterParameters
	Viper      bool
	ConfigFile string
	Language   string
}

type RootParameterSetPtr = *assistant.ParamSet[RootParameterSet]

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

// [blur]
// magick source.jpg -strip -interlace Plane -gaussian-blur 0.05 -quality 85% result.jpg
type BlurParameters struct {
	Gaussian float32
}

// [sampler]
// magick source.jpg -strip -interlace Plane -sampling-factor 4:2:0 -quality 85% result.jpg
type SamplingParameters struct {
	FactorEn assistant.EnumValue[SamplingFactorEnum]
}

type CoreParameters struct {
	BlurParameters
	SamplingParameters

	InterlaceEn assistant.EnumValue[InterlaceEnum]
	//
	Strip   bool
	Quality int
}

// [vanilla]
// magick source.jpg -strip -interlace Plane -quality 85% result.jpg

type ShrinkParameterSet struct {
	CoreParameters
	//
	Directory  string
	MirrorPath string
	ModeEn     assistant.EnumValue[ModeEnum]
}
