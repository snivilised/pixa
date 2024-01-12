package i18n

import (
	xi18n "github.com/snivilised/extendio/i18n"
)

// ShrinkCmdSamplingFactorInvalidTemplData
// ❌
type ShrinkCmdSamplingFactorInvalidTemplData struct {
	pixaTemplData
	Value      string
	Acceptable string
}

func (td ShrinkCmdSamplingFactorInvalidTemplData) Message() *xi18n.Message {
	return &xi18n.Message{
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
	xi18n.LocalisableError
}

func NewInvalidSamplingFactorError(value, acceptable string) InvalidSamplingFactorError {
	return InvalidSamplingFactorError{
		LocalisableError: xi18n.LocalisableError{
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

func (td ShrinkCmdInterlaceInvalidTemplData) Message() *xi18n.Message {
	return &xi18n.Message{
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
	xi18n.LocalisableError
}

func NewInterlaceError(value, acceptable string) InvalidInterlaceError {
	return InvalidInterlaceError{
		LocalisableError: xi18n.LocalisableError{
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

func (td ShrinkCmdOutputPathDoesNotExistTemplData) Message() *xi18n.Message {
	return &xi18n.Message{
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
	xi18n.LocalisableError
}

func NewOutputPathDoesNotExistError(path string) OutputPathDoesNotExistError {
	return OutputPathDoesNotExistError{
		LocalisableError: xi18n.LocalisableError{
			Data: ShrinkCmdOutputPathDoesNotExistTemplData{
				Path: path,
			},
		},
	}
}
