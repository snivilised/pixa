package locale

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/snivilised/li18ngo"
)

// ShrinkCmdSamplingFactorInvalidTemplData
// ❌
type ShrinkCmdSamplingFactorInvalidTemplData struct {
	pixaTemplData
	Value      string
	Acceptable string
}

func (td ShrinkCmdSamplingFactorInvalidTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-cmd-sampling-factor-invalid.error",
		Description: "shrink command sampling factor failed validation",
		Other:       "invalid sampling factor value: {{.Value}}, acceptable: {{.Acceptable}}",
	}
}

// InvalidSamplingFactorErrorBehaviourQuery used to query if an error is:
// "invalid sampling factor value"
type InvalidSamplingFactorErrorBehaviourQuery interface {
	SamplingFactorValidationFailure() bool
}

type InvalidSamplingFactorError struct {
	li18ngo.LocalisableError
}

func NewInvalidSamplingFactorError(value, acceptable string) InvalidSamplingFactorError {
	return InvalidSamplingFactorError{
		LocalisableError: li18ngo.LocalisableError{
			Data: ShrinkCmdSamplingFactorInvalidTemplData{
				Value:      value,
				Acceptable: acceptable,
			},
		},
	}
}

// ShrinkCmdInterlaceInvalidTemplData
// ❌
type ShrinkCmdInterlaceInvalidTemplData struct {
	pixaTemplData
	Value      string
	Acceptable string
}

func (td ShrinkCmdInterlaceInvalidTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-cmd-interlace-invalid.error",
		Description: "shrink command interlace failed validation",
		Other:       "invalid interlace value: {{.Value}}, acceptable: {{.Acceptable}}",
	}
}

// InvalidInterlaceErrorBehaviourQuery used to query if an error is:
// "invalid interlace value"
type InvalidInterlaceErrorBehaviourQuery interface {
	SamplingFactorValidationFailure() bool
}

type InvalidInterlaceError struct {
	li18ngo.LocalisableError
}

func NewInterlaceError(value, acceptable string) InvalidInterlaceError {
	return InvalidInterlaceError{
		LocalisableError: li18ngo.LocalisableError{
			Data: ShrinkCmdInterlaceInvalidTemplData{
				Value:      value,
				Acceptable: acceptable,
			},
		},
	}
}

// ShrinkCmdOutputPathDoesNotExistTemplData
// ❌
type ShrinkCmdOutputPathDoesNotExistTemplData struct {
	pixaTemplData
	Path string
}

func (td ShrinkCmdOutputPathDoesNotExistTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "shrink-cmd-output-path-does-not-exist.error",
		Description: "shrink command mirror path does not exist validation",
		Other:       "output path: {{.Path}}, does not exist",
	}
}

// InvalidInterlaceErrorBehaviourQuery used to query if an error is:
// "invalid interlace value"
type OutputPathDoesNotExistBehaviourQuery interface {
	OutputPathValidationFailure() bool
}

type OutputPathDoesNotExistError struct {
	li18ngo.LocalisableError
}

func NewOutputPathDoesNotExistError(path string) OutputPathDoesNotExistError {
	return OutputPathDoesNotExistError{
		LocalisableError: li18ngo.LocalisableError{
			Data: ShrinkCmdOutputPathDoesNotExistTemplData{
				Path: path,
			},
		},
	}
}