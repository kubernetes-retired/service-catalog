package tpr

import (
	"testing"
)

func TestTPRName(t *testing.T) {
	testCases := []struct {
		before string
		after  string
	}{
		{before: "ServiceClass", after: "service-class"},
		{before: "ThisIsAThing", after: "this-is-a-thing"},
		{before: "thisIsAThing", after: "this-is-a-thing"},
		{before: "Binding", after: "binding"},
	}
	for _, testCase := range testCases {
		kind := Kind(testCase.before)
		if kind.TPRName() != testCase.after {
			t.Errorf("expected %s, got %s", testCase.after, kind.TPRName())
		}
	}
}
