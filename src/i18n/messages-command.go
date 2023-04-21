package i18n

import (
	"github.com/snivilised/extendio/i18n"
)

// ⚠️ for the usage definitions, make sure that the first token inside the "Other"
// field is the name of the flag as this is used to look up the short code definition.
// failure to comply may or may not result in a error in defining the flag on the
// command's parameter set.

// RootCmdShortDescTemplData
// 🧊
type RootCmdShortDescTemplData struct {
	pixaTemplData
}

func (td RootCmdShortDescTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command.short-description",
		Description: "short description for the root command",
		Other:       "A brief description of your application",
	}
}

// RootCmdLongDescTemplData
// 🧊
type RootCmdLongDescTemplData struct {
	pixaTemplData
}

func (td RootCmdLongDescTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command.long-description",
		Description: "long description for the root command",
		Other: `A longer description that spans multiple lines and likely contains
		examples and usage of using your application. For example:
		
		Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
	}
}

// RootCmdConfigFileUsageTemplData
// 🧊
type RootCmdConfigFileUsageTemplData struct {
	pixaTemplData
}

func (td RootCmdConfigFileUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command-config-file.param-usage",
		Description: "root command config flag usage",
		Other:       "configuration file",
	}
}

// RootCmdLangUsageTemplData
// 🧊
type RootCmdLangUsageTemplData struct {
	pixaTemplData
}

func (td RootCmdLangUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command-language.param-usage",
		Description: "root command lang usage",
		Other:       "lang defines the language according to IETF BCP 47",
	}
}

// RootCmdFolderRexExParamUsageTemplData
// 🧊
type RootCmdFolderRexExParamUsageTemplData struct {
	pixaTemplData
}

func (td RootCmdFolderRexExParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command-folder-regex.param-usage",
		Description: "root command folder regex filter (negate-able with leading !)",
		Other:       "folder-rx folder regular expression filter (negate-able with leading !)",
	}
}

// RootCmdFolderGlobParamUsageTemplData
// 🧊
type RootCmdFolderGlobParamUsageTemplData struct {
	pixaTemplData
}

func (td RootCmdFolderGlobParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command-folder-glob.param-usage",
		Description: "root command folder glob (negate-able with leading !)",
		Other:       "folder-gb folder glob filter (negate-able with leading !)",
	}
}

// RootCmdFilesRegExParamUsageTemplData
// 🧊
type RootCmdFilesRegExParamUsageTemplData struct {
	pixaTemplData
}

func (td RootCmdFilesRegExParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command-files-regex.param-usage",
		Description: "root command files regex filter (negate-able with leading !)",
		Other:       "files-rx folder regular expression filter (negate-able with leading !)",
	}
}

// RootCmdFilesGlobParamUsageTemplData
// 🧊
type RootCmdFilesGlobParamUsageTemplData struct {
	pixaTemplData
}

func (td RootCmdFilesGlobParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command-files-glob.param-usage",
		Description: "root command files glob (negate-able with leading !)",
		Other:       "files-gb files glob filter (negate-able with leading !)",
	}
}

// TODO: add shrink parameter usage here ...

// ShrinkCmdGaussianBlurParamUsageTemplData
// 🧊
type ShrinkCmdGaussianBlurParamUsageTemplData struct {
	pixaTemplData
}

func (td ShrinkCmdGaussianBlurParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-cmd-gaussian-blur.param-usage",
		Description: "shrink command gaussian blur parameter usage (see magick documentation for more info)",
		Other:       "gaussian-blur (see magick documentation for more info)",
	}
}

// ShrinkCmdSamplingFactorParamUsageTemplData
// 🧊
type ShrinkCmdSamplingFactorParamUsageTemplData struct {
	pixaTemplData
}

func (td ShrinkCmdSamplingFactorParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-cmd-sampling-factor.param-usage",
		Description: "shrink command sampling factor parameter usage (see magick documentation for more info)",
		Other:       "sampling-factor (see magick documentation for more info)",
	}
}

// ShrinkCmdInterlaceParamUsageTemplData
// 🧊
type ShrinkCmdInterlaceParamUsageTemplData struct {
	pixaTemplData
}

func (td ShrinkCmdInterlaceParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-cmd-interlace.param-usage",
		Description: "shrink command interlace parameter usage (see magick documentation for more info)",
		Other:       "interlace (see magick documentation for more info)",
	}
}

// ShrinkCmdStripParamUsageTemplData
// 🧊
type ShrinkCmdStripParamUsageTemplData struct {
	pixaTemplData
}

func (td ShrinkCmdStripParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-cmd-strip.param-usage",
		Description: "shrink strip parameter usage (see magick documentation for more info)",
		Other:       "strip (see magick documentation for more info)",
	}
}

// ShrinkCmdQualityParamUsageTemplData
// 🧊
type ShrinkCmdQualityParamUsageTemplData struct {
	pixaTemplData
}

func (td ShrinkCmdQualityParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-cmd-quality.param-usage",
		Description: "shrink quality parameter usage (see magick documentation for more info)",
		Other:       "quality (see magick documentation for more info)",
	}
}

// ShrinkCmdMirrorPathParamUsageTemplData
// 🧊
type ShrinkCmdMirrorPathParamUsageTemplData struct {
	pixaTemplData
}

func (td ShrinkCmdMirrorPathParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-cmd-mirror-path.param-usage",
		Description: "shrink mirror path creates a mirror of the source directory tree containing processed images",
		Other:       "mirror-path creates a mirror of the source directory tree containing processed images",
	}
}

// ShrinkCmdModeParamUsageTemplData
// 🧊
type ShrinkCmdModeParamUsageTemplData struct {
	pixaTemplData
}

func (td ShrinkCmdModeParamUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-cmd-mode.param-usage",
		Description: "shrink command mode parameter usage",
		Other:       "mode defines the output behaviour of the image processing traversal batch",
	}
}

// ShrinkCmdShortDefinitionTemplData
// 🧊
type ShrinkCmdShortDefinitionTemplData struct {
	pixaTemplData
}

func (td ShrinkCmdShortDefinitionTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-command.short-description",
		Description: "Short description for shrink command",
		Other:       "bulk image compressor",
	}
}

// ShrinkLongDefinitionTemplData
// 🧊
type ShrinkLongDefinitionTemplData struct {
	pixaTemplData
}

func (td ShrinkLongDefinitionTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-command.long-description",
		Description: "Long description for shrink command",
		Other:       "Directory tree based bulk image processor (using ImageMagick)",
	}
}
