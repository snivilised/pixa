package helpers

import (
	"bytes"

	"github.com/spf13/cobra"
)

// CommandTester eases creation of unit tests for cobra command
type CommandTester struct {
	Args []string
	Root *cobra.Command
}

func (ch *CommandTester) Execute() (string, error) {
	_, output, err := ch.executeC()
	return output, err
}

func (ch *CommandTester) executeC() (*cobra.Command, string, error) {
	buf := new(bytes.Buffer)
	ch.Root.SetOut(buf)
	ch.Root.SetErr(buf)
	ch.Root.SetArgs(ch.Args)

	c, err := ch.Root.ExecuteC()

	return c, buf.String(), err
}
