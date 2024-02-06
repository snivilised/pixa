package cfg_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCfg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cfg Suite")
}
