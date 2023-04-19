package i18n

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	xi18n "github.com/snivilised/extendio/i18n"
)

// ❌ FooBar

// FooBarTemplData - TODO: this is a none existent error that should be
// replaced by the client. Its just defined here to illustrate the pattern
// that should be used to implement i18n with extendio. Also note,
// that this message has been removed from the translation files, so
// it is not useable at run time.
type FooBarTemplData struct {
	arcadiaTemplData
	Path   string
	Reason error
}

// the ID should use spp/library specific code, so replace arcadia with the
// name of the library implementing this template project.
func (td FooBarTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "foo-bar.arcadia.nav",
		Description: "Foo Bar description",
		Other:       "foo bar failure '{{.Path}}' (reason: {{.Reason}})",
	}
}

// FooBarErrorBehaviourQuery used to query if an error is:
// "Failed to read directory contents from the path specified"
type FooBarErrorBehaviourQuery interface {
	FooBar() bool
}

type FooBarError struct {
	xi18n.LocalisableError
}

// FooBar enables the client to check if error is FooBarError
// via FooBarErrorBehaviourQuery
func (e FooBarError) FooBar() bool {
	return true
}

// NewFooBarError creates a FooBarError
func NewFooBarError(path string, reason error) FooBarError {
	return FooBarError{
		LocalisableError: xi18n.LocalisableError{
			Data: FooBarTemplData{
				Path:   path,
				Reason: reason,
			},
		},
	}
}
