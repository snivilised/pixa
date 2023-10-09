package magick_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMagick(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Magick Suite")
}
