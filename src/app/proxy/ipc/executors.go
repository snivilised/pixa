package ipc

import (
	"fmt"
	"os/exec"
	"strings"
)

type ProgramExecutor struct {
	Name string
}

func (e *ProgramExecutor) ProgName() string {
	return e.Name
}

func (e *ProgramExecutor) Look() (string, error) {
	return exec.LookPath(e.Name)
}

func (e *ProgramExecutor) Execute(args ...string) error {
	fmt.Printf("✨ executing: '%v %v'\n",
		e.Name,
		strings.Join(args, " "),
	)

	// #nosec G204 // prog(e.Name) is pre-vetted
	cmd := exec.Command(e.Name, args...)
	err := cmd.Start()

	if err != nil {
		return err
	}

	return cmd.Wait()
}

type DummyExecutor struct {
	Name string
}

func (e *DummyExecutor) ProgName() string {
	return e.Name
}

func (e *DummyExecutor) Look() (string, error) {
	return "", nil
}

func (e *DummyExecutor) Execute(args ...string) error {
	fmt.Printf("✨ dummy:executing: '%v %v'\n",
		e.Name,
		strings.Join(args, " "),
	)

	return nil
}
