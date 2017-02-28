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
		{before: "ThisIsAAThing", after: "this-is-a-a-thing"},
	}
	for _, testCase := range testCases {
		kind := Kind(testCase.before)
		if kind.TPRName() != testCase.after {
			t.Errorf("expected %s, got %s", testCase.after, kind.TPRName())
		}
	}
}

func TestURLName(t *testing.T) {
	testCases := []struct {
		before string
		after  string
	}{
		{before: "ServiceClass", after: "serviceclasses"},
		{before: "ThisIsAThing", after: "thisisathings"},
		{before: "thisIsAThing", after: "thisisathings"},
		{before: "Binding", after: "bindings"},
	}

	for _, testCase := range testCases {
		kind := Kind(testCase.before)
		if kind.URLName() != testCase.after {
			t.Errorf("expected %s, got %s", testCase.after, kind.URLName())
		}
	}
}
