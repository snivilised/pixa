package ipc

import "errors"

// Internal errors are those which are supposed to be handled
// internally and are of no significance to the user directly, which
// means they also don't need to be i18n error messages.

var ErrUsingDummyExecutor = errors.New("using dummy executor")
