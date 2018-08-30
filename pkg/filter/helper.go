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
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/labels"
)

var conditionalsRegex = regexp.MustCompile("=|==|!=| in | notin ")

// CreatePredicate creates the Predicate that will be used to
// test if acceptance is allowed for service classes.
func CreatePredicate(restrictions []string) (Predicate, error) {
	// default is no requirements
	requirements := ""
	if len(restrictions) > 0 {
		requirements = string(restrictions[0])

		for i := 1; i < len(restrictions); i++ {
			requirements = fmt.Sprintf("%s, %s", requirements, string(restrictions[i]))
		}
	}

	selector, err := labels.Parse(requirements)
	if err != nil {
		return nil, err
	}
	predicate := internalPredicate{selector: selector}
	return predicate, nil
}

// ConvertToSelector converts Predicate to a labels.Selector
func ConvertToSelector(p Predicate) (labels.Selector, error) {
	return labels.Parse(p.String())
}

// ExtractProperty extracts the property from the given restriction
// E.g., for the restriction "spec.externalName=foo", the function
// returns "spec.externalName"
func ExtractProperty(restriction string) string {
	return conditionalsRegex.Split(restriction, 2)[0]
}
