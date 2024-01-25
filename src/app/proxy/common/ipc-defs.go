package common

import "github.com/snivilised/cobrass/src/clif"

type (
	ExecutionAgent interface {
		// IsInstalled determines whether the underlying program is installed
		IsInstalled() bool

		// Invoke returns the command line args required for the executor to
		// run correctly. Only the source and destination are required because
		// they are the dynamic args that change for each invocation. All the
		// other flags that come directly from the command line/config possibly
		// via a profile are static, which means the concrete agent will already
		// have those and merely has to formulate the complete command line in
		// the correct order required by the third party program.
		Invoke(thirdPartyCL clif.ThirdPartyCommandLine, source, destination string) error
	}

	Executor interface {
		ProgName() string
		Look() (string, error)
		Execute(args ...string) error
	}
)
