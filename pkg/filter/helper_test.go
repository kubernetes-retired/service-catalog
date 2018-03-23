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
	"testing"
)

func TestCreatePredicate(t *testing.T) {
	cases := []struct {
		name         string
		restrictions []string
		error        bool
		predicate    string
	}{
		{
			name: "no restrictions",
		},
		{
			name: "invalid class restrictions",
			restrictions: []string{
				"this throws an error",
			},
			error: true,
		},
		{
			name: "valid class restriction",
			restrictions: []string{
				"name in (Foo, Bar)",
			},
			predicate: "name in (Bar,Foo)",
		},
		{
			name: "valid class double restriction and wacky spacing",
			restrictions: []string{
				"name   in      (Foo,   Bar)",
				"name   notin   (Baz,   Barf)",
			},
			predicate: "name in (Bar,Foo),name notin (Barf,Baz)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			predicate, err := CreatePredicate(tc.restrictions)

			if err != nil {
				if tc.error {
					return
				}
				t.Fatalf("Unexpected error from CreatePredicateForServiceClassesFromRestrictions: %v", err)
			}

			if predicate == nil {
				t.Fatalf("Failed to create predicate from restrictions: %+v", tc.restrictions)
			}

			if tc.restrictions == nil && !predicate.Empty() {
				t.Fatalf("Failed to create predicate an empty prediate from nil restrictions.")
			}

			// test the predicate is what we expected.
			ps := predicate.String()
			if ps != tc.predicate {
				t.Fatalf("Failed to create expected predicate, \n\texpected: \t%q,\n \tgot: \t\t%q", tc.predicate, ps)
			}
		})
	}
}
