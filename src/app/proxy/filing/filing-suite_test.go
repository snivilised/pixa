package filing_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // foo
	. "github.com/onsi/gomega"    //nolint:revive // foo
)

func TestFiling(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Filing Suite")
}
