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

package requirements

import (
	"sort"
	"strings"
)

// A Predicate is the representation of all the Requirements.
type Predicate interface {
	Matches(Terms) bool

	// Empty returns true if this predicate does not restrict the selection space.
	Empty() bool

	// String returns a human readable string that represents this selector.
	String() string

	// Add adds requirements to the ListablePredicate
	Add(r ...Requirement) Predicate

	// Requirements converts this interface into Requirements to expose
	// more detailed selection information.
	// If there are querying parameters, it will return converted requirements and selectable=true.
	// If this selector doesn't want to select anything, it will return selectable=false.
	Requirements() (requirements Requirements, selectable bool)
}

// Everything returns a predicate that matches everythingPredicate.
func Everything() Predicate {
	return internalPredicate{}
}

type nothingPredicate struct{}

func (n nothingPredicate) Matches(_ Terms) bool               { return false }
func (n nothingPredicate) Empty() bool                        { return false }
func (n nothingPredicate) String() string                     { return "" }
func (n nothingPredicate) Add(_ ...Requirement) Predicate     { return n }
func (n nothingPredicate) Requirements() (Requirements, bool) { return nil, false }

// Nothing returns a selector that matches no labels
func Nothing() Predicate {
	return nothingPredicate{}
}

// NewSelector returns a nil selector
func NewPredicate() Predicate {
	return internalPredicate(nil)
}

type internalPredicate []Requirement

// Matches for a internalSelector returns true if all
// its Requirements match the input Labels. If any
// Requirement does not match, false is returned.
func (lsel internalPredicate) Matches(l Terms) bool {
	for ix := range lsel {
		if matches := lsel[ix].Matches(l); !matches {
			return false
		}
	}
	return true
}

// Empty returns true if the internalSelector doesn't restrict selection space
func (lsel internalPredicate) Empty() bool {
	if lsel == nil {
		return true
	}
	return len(lsel) == 0
}

// Add adds requirements to the selector. It copies the current selector returning a new one
func (lsel internalPredicate) Add(reqs ...Requirement) Predicate {
	var sel internalPredicate
	for ix := range lsel {
		sel = append(sel, lsel[ix])
	}
	for _, r := range reqs {
		sel = append(sel, r)
	}
	sort.Sort(ByKey(sel))
	return sel
}

func (lsel internalPredicate) Requirements() (Requirements, bool) { return Requirements(lsel), true }

// String returns a comma-separated string of all
// the internalSelector Requirements' human-readable strings.
func (lsel internalPredicate) String() string {
	var reqs []string
	for ix := range lsel {
		reqs = append(reqs, lsel[ix].String())
	}
	return strings.Join(reqs, ",")
}
