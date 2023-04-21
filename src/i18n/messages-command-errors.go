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

// ShrinkCmdMirrorPathDoesNotExistTemplData
// ❌
type ShrinkCmdMirrorPathDoesNotExistTemplData struct {
	pixaTemplData
	Path string
}

func (td ShrinkCmdMirrorPathDoesNotExistTemplData) Message() *xi18n.Message {
	return &xi18n.Message{
		ID:          "shrink-cmd-mirror-path-does-not-exist.error",
		Description: "shrink command mirror path does not exist validation",
		Other:       "mirror path: {{.Path}}, does not exist",
	}
}

// InvalidInterlaceErrorBehaviourQuery used to query if an error is:
// "invalid interlace value"
type MirrorPathDoesNotExistBehaviourQuery interface {
	MirrorPathValidationFailure() bool
}

type MirrorPathDoesNotExistError struct {
	xi18n.LocalisableError
}

func NewMirrorPathDoesNotExistError(path string) InvalidInterlaceError {
	return InvalidInterlaceError{
		LocalisableError: xi18n.LocalisableError{
			Data: ShrinkCmdMirrorPathDoesNotExistTemplData{
				Path: path,
			},
		},
	}
}

// ShrinkCmdModeInvalidTemplData
// ❌
type ShrinkCmdModeInvalidTemplData struct {
	pixaTemplData
	Value      string
	Acceptable string
}

func (td ShrinkCmdModeInvalidTemplData) Message() *xi18n.Message {
	return &xi18n.Message{
		ID:          "shrink-cmd-mode-invalid.error",
		Description: "shrink command mode failed validation",
		Other:       "invalid mode value: {{.Value}}, acceptable: {{.Acceptable}}",
	}
}

// InvalidModeErrorBehaviourQuery used to query if an error is:
// "invalid mode value"
type InvalidModeErrorBehaviourQuery interface {
	ModeValidationFailure() bool
}

type InvalidModeError struct {
	xi18n.LocalisableError
}

func NewModeError(value, acceptable string) InvalidModeError {
	return InvalidModeError{
		LocalisableError: xi18n.LocalisableError{
			Data: ShrinkCmdModeInvalidTemplData{
				Value:      value,
				Acceptable: acceptable,
			},
		},
	}
}
