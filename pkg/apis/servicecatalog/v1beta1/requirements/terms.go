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

type Term string

const (
	Name         Term = "name"
	ExternalName Term = "externalName"
	ExternalID   Term = "externalID"
)

type Terms interface {
	Has(term Term) (exists bool)
	Get(term Term) (value string)
}

// Set is a map of term:value. It implements Terms.
type Set map[Term]string

// String returns all terms listed as a human readable string.
func (s Set) String() string {
	terms := make([]string, 0, len(s))
	for term, value := range s {
		terms = append(terms, string(term)+"="+value)
	}
	sort.StringSlice(terms).Sort()
	return strings.Join(terms, ",")
}

// Has returns whether the provided term exists in the map.
func (s Set) Has(term Term) bool {
	_, exists := s[term]
	return exists
}

// Get returns the value in the map for the provided term.
func (s Set) Get(term Term) string {
	return s[term]
}

//
//func ConvertClusterServicePlanToTerms(plan *v1beta1.ClusterServicePlan) Terms {
//	return Set{
//		Name:         plan.Name,
//		ExternalName: plan.Spec.ExternalName,
//		ExternalID:   plan.Spec.ExternalID,
//	}
//}
//
//func ConvertClusterServiceClassToTerms(class *v1beta1.ClusterServiceClass) Terms {
//	return Set{
//		Name:         class.Name,
//		ExternalName: class.Spec.ExternalName,
//		ExternalID:   class.Spec.ExternalID,
//	}
//}
