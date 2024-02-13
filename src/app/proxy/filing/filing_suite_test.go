package filing_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFiling(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Filing Suite")
}
