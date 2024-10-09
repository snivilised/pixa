package lab

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"

	. "github.com/onsi/gomega/types" //nolint:stylecheck,revive // ok
	"github.com/samber/lo"
	"github.com/snivilised/traverse/core"
	"github.com/snivilised/traverse/enums"
)

type DirectoryContentsMatcher struct {
	expected      interface{}
	expectedNames []string
	actualNames   []string
}

func HaveDirectoryContents(expected interface{}) GomegaMatcher {
	return &DirectoryContentsMatcher{
		expected: expected,
	}
}

func (m *DirectoryContentsMatcher) Match(actual interface{}) (bool, error) {
	entries, entriesOk := actual.([]fs.DirEntry)
	if !entriesOk {
		return false, fmt.Errorf("🔥 matcher expected []fs.DirEntry (%T)", entries)
	}

	m.actualNames = lo.Map(entries, func(entry fs.DirEntry, _ int) string {
		return entry.Name()
	})

	expected, expectedOk := m.expected.([]string)
	if !expectedOk {
		return false, fmt.Errorf("🔥 matcher expected []string (%T)", expected)
	}
	m.expectedNames = expected

	return slices.Compare(m.actualNames, m.expectedNames) == 0, nil
}

func (m *DirectoryContentsMatcher) FailureMessage(_ interface{}) string {
	return fmt.Sprintf(
		"❌ DirectoryContentsMatcher Expected\n\t%v\nto match contents\n\t%v\n",
		strings.Join(m.expectedNames, ", "), strings.Join(m.actualNames, ", "),
	)
}

func (m *DirectoryContentsMatcher) NegatedFailureMessage(_ interface{}) string {
	return fmt.Sprintf(
		"❌ DirectoryContentsMatcher Expected\n\t%v\nNOT to match contents\n\t%v\n",
		strings.Join(m.expectedNames, ", "), strings.Join(m.actualNames, ", "),
	)
}

type InvokeNodeMatcher struct {
	expected  interface{}
	mandatory string
}

func HaveInvokedNode(expected interface{}) GomegaMatcher {
	return &InvokeNodeMatcher{
		expected: expected,
	}
}

func (m *InvokeNodeMatcher) Match(actual interface{}) (bool, error) {
	recording, ok := actual.(RecordingMap)
	if !ok {
		return false, fmt.Errorf(
			"InvokeNodeMatcher expected actual to be a RecordingMap (%T)",
			actual,
		)
	}

	mandatory, ok := m.expected.(string)
	if !ok {
		return false, fmt.Errorf("InvokeNodeMatcher expected string (%T)", actual)
	}
	m.mandatory = mandatory

	_, found := recording[m.mandatory]

	return found, nil
}

func (m *InvokeNodeMatcher) FailureMessage(_ interface{}) string {
	return fmt.Sprintf("❌ Expected\n\t%v\nnode to be invoked\n",
		m.mandatory,
	)
}

func (m *InvokeNodeMatcher) NegatedFailureMessage(_ interface{}) string {
	return fmt.Sprintf("❌ Expected\n\t%v\nnode NOT to be invoked\n",
		m.mandatory,
	)
}

type NotInvokeNodeMatcher struct {
	expected  interface{}
	mandatory string
}

func HaveNotInvokedNode(expected interface{}) GomegaMatcher {
	return &NotInvokeNodeMatcher{
		expected: expected,
	}
}

func (m *NotInvokeNodeMatcher) Match(actual interface{}) (bool, error) {
	recording, ok := actual.(RecordingMap)
	if !ok {
		return false, fmt.Errorf("matcher expected actual to be a RecordingMap (%T)", actual)
	}

	mandatory, ok := m.expected.(string)
	if !ok {
		return false, fmt.Errorf("matcher expected string (%T)", actual)
	}
	m.mandatory = mandatory

	_, found := recording[m.mandatory]

	return !found, nil
}

func (m *NotInvokeNodeMatcher) FailureMessage(_ interface{}) string {
	return fmt.Sprintf("❌ Expected\n\t%v\nnode to NOT be invoked\n",
		m.mandatory,
	)
}

func (m *NotInvokeNodeMatcher) NegatedFailureMessage(_ interface{}) string {
	return fmt.Sprintf("❌ Expected\n\t%v\nnode to be invoked\n",
		m.mandatory,
	)
}

type (
	ExpectedCount struct {
		Name  string
		Count int
	}

	ChildCountMatcher struct {
		expected    interface{}
		expectation MatcherExpectation[uint]
		name        string
	}
)

func HaveChildCountOf(expected interface{}) GomegaMatcher {
	return &ChildCountMatcher{
		expected: expected,
	}
}

func (m *ChildCountMatcher) Match(actual interface{}) (bool, error) {
	recording, ok := actual.(RecordingMap)
	if !ok {
		return false, fmt.Errorf("ChildCountMatcher expected actual to be a RecordingMap (%T)", actual)
	}

	expected, ok := m.expected.(ExpectedCount)
	if !ok {
		return false, fmt.Errorf("ChildCountMatcher expected ExpectedCount (%T)", actual)
	}

	count, ok := recording[expected.Name]
	if !ok {
		return false, fmt.Errorf("🔥 not found: '%v'", expected.Name)
	}

	m.expectation = MatcherExpectation[uint]{
		Expected: uint(expected.Count),
		Actual:   uint(count),
	}
	m.name = expected.Name

	return m.expectation.IsEqual(), nil
}

func (m *ChildCountMatcher) FailureMessage(_ interface{}) string {
	return fmt.Sprintf(
		"❌ Expected child count for node: '%v' to be equal; expected: '%v', actual: '%v'\n",
		m.name, m.expectation.Expected, m.expectation.Actual,
	)
}

func (m *ChildCountMatcher) NegatedFailureMessage(_ interface{}) string {
	return fmt.Sprintf(
		"❌ Expected child count for node: '%v' NOT to be equal; expected: '%v', actual: '%v'\n",
		m.name, m.expectation.Expected, m.expectation.Actual,
	)
}

type (
	ExpectedMetric struct {
		Type  enums.Metric
		Count uint
	}

	MetricMatcher struct {
		expected    interface{}
		expectation MatcherExpectation[uint]
		typ         enums.Metric
	}
)

func HaveMetricCountOf(expected interface{}) GomegaMatcher {
	return &MetricMatcher{
		expected: expected,
	}
}

func (m *MetricMatcher) Match(actual interface{}) (bool, error) {
	result, ok := actual.(core.TraverseResult)
	if !ok {
		return false, fmt.Errorf(
			"🔥 MetricMatcher expected actual to be a core.TraverseResult (%T)",
			actual,
		)
	}

	expected, ok := m.expected.(ExpectedMetric)
	if !ok {
		return false, fmt.Errorf("🔥 MetricMatcher expected ExpectedMetric (%T)", actual)
	}

	m.expectation = MatcherExpectation[uint]{
		Expected: expected.Count,
		Actual:   result.Metrics().Count(expected.Type),
	}
	m.typ = expected.Type

	return m.expectation.IsEqual(), nil
}

func (m *MetricMatcher) FailureMessage(_ interface{}) string {
	return fmt.Sprintf(
		"❌ Expected metric '%v' to be equal; expected:'%v', actual: '%v'\n",
		m.typ.String(), m.expectation.Expected, m.expectation.Actual,
	)
}

func (m *MetricMatcher) NegatedFailureMessage(_ interface{}) string {
	return fmt.Sprintf(
		"❌ Expected metric '%v' NOT to be equal; expected:'%v', actual: '%v'\n",
		m.typ.String(), m.expectation.Expected, m.expectation.Actual,
	)
}
