package matchers

import (
	"fmt"

	"github.com/onsi/gomega/types"
	"github.com/snivilised/traverse/lfs"
)

type PathExistsMatcher struct {
	FS interface{}
}

type AsDirectory string
type AsFile string

func ExistInFS(fs interface{}) types.GomegaMatcher {
	return &PathExistsMatcher{
		FS: fs,
	}
}

func (m *PathExistsMatcher) Match(actual interface{}) (bool, error) {
	FS, fileSystemOK := m.FS.(lfs.MkDirAllFS)
	if !fileSystemOK {
		return false, fmt.Errorf("❌ matcher expected a VirtualFS instance (%T)", FS)
	}

	if actualPath, dirOK := actual.(AsDirectory); dirOK {
		return FS.DirectoryExists(string(actualPath)), nil
	}

	if actualPath, fileOK := actual.(AsFile); fileOK {
		return FS.FileExists(string(actualPath)), nil
	}

	return false, fmt.Errorf("❌ matcher expected an AsDirectory or AsFile instance (%T)", actual)
}

func (m *PathExistsMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("🔥 Expected\n\t%v\npath to exist", actual)
}

func (m *PathExistsMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("🔥 Expected\n\t%v\npath NOT to exist\n", actual)
}
