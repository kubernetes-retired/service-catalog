/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package filter

import (
	"k8s.io/apimachinery/pkg/labels"
)

// Want to filter the allowed labels.
// want to use only
//const (
//	DoesNotExist Operator = "!"
//	Equals       Operator = "="
//	DoubleEquals Operator = "=="
//	In           Operator = "in"
//	NotEquals    Operator = "!="
//	NotIn        Operator = "notin"
//)

// Predicate is used to test if this rule accepts the properties given.
// A Predicate wraps a label.Selector allowing us to use selectors.
type Predicate interface {
	// Accepts returns true if this predicate accepts the given set of properties.
	Accepts(Properties) bool

	// Empty returns true if this predicate does not restrict the acceptance space.
	Empty() bool

	// String returns a human readable string that represents this predicate.
	String() string
}

// NewPredicate returns a empty predicate
func NewPredicate() Predicate {
	return internalPredicate{}
}

// internalPredicate is our internal representation of Predicate. It will be
// implemented as a wrapper around labels.Selector to leverage the label
// selector work.
type internalPredicate struct {
	selector labels.Selector
}

// Accepts tests to see if the given properties are allowed for this
// predicate. If there is no predicate, then it is
func (ip internalPredicate) Accepts(p Properties) bool {
	if ip.Empty() {
		return true
	}
	return ip.selector.Matches(p)
}

// Empty returns true if this predicate does not restrict the acceptance space.
func (ip internalPredicate) Empty() bool {
	if ip.selector == nil {
		return true
	}
	return ip.selector.Empty()
}

// String returns a human-readable version of the selector.
func (ip internalPredicate) String() string {
	return ip.selector.String()
}
