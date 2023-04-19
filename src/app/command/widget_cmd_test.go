package command_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/arcadia/src/app/command"
	"github.com/snivilised/arcadia/src/internal/helpers"
)

var _ = Describe("WidgetCmd", func() {

	When("specified flags are valid", func() {
		It("🧪 should: execute without error", func() {
			bootstrap := command.Bootstrap{
				Detector: &DetectorStub{},
			}
			tester := helpers.CommandTester{
				Args: []string{"widget", "-p", "P?<date>", "-t", "42"},
				Root: bootstrap.Root(),
			}
			_, err := tester.Execute()
			Expect(err).Error().To(BeNil(),
				"should pass validation due to all flag being valid",
			)
		})
	})

	When("specified flags are valid", func() {
		It("🧪 should: execute without error", func() {
			bootstrap := command.Bootstrap{
				Detector: &DetectorStub{},
			}
			tester := helpers.CommandTester{
				Args: []string{"widget", "-p", "P?<date>", "-t", "99"},
				Root: bootstrap.Root(),
			}
			_, err := tester.Execute()
			Expect(err).Error().NotTo(BeNil(),
				"expected validation failure due to -t being within out of range",
			)
		})
	})
})
