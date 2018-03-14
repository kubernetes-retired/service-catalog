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
	"encoding/json"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCreatePredicateForServiceClassesFromRestrictions(t *testing.T) {
	cases := []struct {
		name         string
		restrictions *v1beta1.ServiceCatalogRestrictions
		error        bool
		predicate    string
	}{
		{
			name: "no restrictions",
		},
		{
			name: "invalid class restrictions",
			restrictions: &v1beta1.ServiceCatalogRestrictions{
				ServiceClass: []v1beta1.ClusterServiceClassRequirement{
					"this throws an error",
				},
			},
			error: true,
		},
		{
			name: "valid class restriction",
			restrictions: &v1beta1.ServiceCatalogRestrictions{
				ServiceClass: []v1beta1.ClusterServiceClassRequirement{
					"name in (Foo, Bar)",
				},
			},
			predicate: "name in (Bar,Foo)",
		},
		{
			name: "valid class restriction, ignores plan restriction",
			restrictions: &v1beta1.ServiceCatalogRestrictions{
				ServiceClass: []v1beta1.ClusterServiceClassRequirement{
					"name in (Foo, Bar)",
				},
				ServicePlan: []v1beta1.ClusterServicePlanRequirement{
					"name in (Boo, Far)",
				},
			},
			predicate: "name in (Bar,Foo)",
		},
		{
			name: "valid class double restriction and wacky spacing",
			restrictions: &v1beta1.ServiceCatalogRestrictions{
				ServiceClass: []v1beta1.ClusterServiceClassRequirement{
					"name   in      (Foo,   Bar)",
					"name   notin   (Baz,   Barf)",
				},
			},
			predicate: "name in (Bar,Foo),name notin (Barf,Baz)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			predicate, err := CreatePredicateForServiceClassesFromRestrictions(tc.restrictions)

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

func TestCreatePredicateForServicePlansFromRestrictions(t *testing.T) {
	cases := []struct {
		name         string
		restrictions *v1beta1.ServiceCatalogRestrictions
		error        bool
		predicate    string
	}{
		{
			name: "no restrictions",
		},
		{
			name: "invalid plan restrictions",
			restrictions: &v1beta1.ServiceCatalogRestrictions{
				ServicePlan: []v1beta1.ClusterServicePlanRequirement{
					"this throws an error",
				},
			},
			error: true,
		},
		{
			name: "valid plan restriction",
			restrictions: &v1beta1.ServiceCatalogRestrictions{
				ServicePlan: []v1beta1.ClusterServicePlanRequirement{
					"name in (Foo, Bar)",
				},
			},
			predicate: "name in (Bar,Foo)",
		},
		{
			name: "valid plan restriction, ignores class restriction",
			restrictions: &v1beta1.ServiceCatalogRestrictions{
				ServiceClass: []v1beta1.ClusterServiceClassRequirement{
					"name in (Foo, Bar)",
				},
				ServicePlan: []v1beta1.ClusterServicePlanRequirement{
					"name in (Bar, Foo)",
				},
			},
			predicate: "name in (Bar,Foo)",
		},
		{
			name: "valid plan double restriction and wacky spacing",
			restrictions: &v1beta1.ServiceCatalogRestrictions{
				ServicePlan: []v1beta1.ClusterServicePlanRequirement{
					"name   notin   (Baz,   Barf)",
					"name   in      (Foo,   Bar)",
				},
			},
			predicate: "name notin (Barf,Baz),name in (Bar,Foo)",
		},
		{
			name: "valid plan tripple restriction and wacky spacing and kinda using array",
			restrictions: &v1beta1.ServiceCatalogRestrictions{
				ServicePlan: []v1beta1.ClusterServicePlanRequirement{
					"name   notin   (Baz,   Barf), name=Taz",
					"name   in      (Foo,   Bar)",
				},
			},
			predicate: "name notin (Barf,Baz),name=Taz,name in (Bar,Foo)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			predicate, err := CreatePredicateForServicePlansFromRestrictions(tc.restrictions)

			if err != nil {
				if tc.error {
					return
				}
				t.Fatalf("Unexpected error from CreatePredicateForServicePlansFromRestrictions: %v", err)
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

func TestConvertServiceClassToProperties(t *testing.T) {
	cases := []struct {
		name string
		sc   *v1beta1.ClusterServiceClass
		json string
	}{
		{
			name: "nil object",
			json: "{}",
		},
		{
			name: "normal object",
			sc: &v1beta1.ClusterServiceClass{
				ObjectMeta: metav1.ObjectMeta{Name: "service-class"},
				Spec: v1beta1.ClusterServiceClassSpec{
					ExternalName: "external-class-name",
					ExternalID:   "external-id",
				},
			},
			json: "{\"name\":\"service-class\",\"spec.externalID\":\"external-id\",\"spec.externalName\":\"external-class-name\"}",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := ConvertServiceClassToProperties(tc.sc)
			if p == nil {
				t.Fatalf("Failed to create Properties object from %+v", tc.sc)
			}
			b, err := json.Marshal(p)
			if err != nil {
				t.Fatalf("Unexpected error with json marchal, %v", err)
			}
			js := string(b)
			if js != tc.json {
				t.Fatalf("Failed to create expected Properties object,\n\texpected: \t%q,\n \tgot: \t\t%q", tc.json, js)
			}
		})
	}
}

func TestConvertServicePlanToProperties(t *testing.T) {
	cases := []struct {
		name string
		sp   *v1beta1.ClusterServicePlan
		json string
	}{
		{
			name: "nil object",
			json: "{}",
		},
		{
			name: "normal object",
			sp: &v1beta1.ClusterServicePlan{
				ObjectMeta: metav1.ObjectMeta{Name: "service-plan"},
				Spec: v1beta1.ClusterServicePlanSpec{
					ExternalName: "external-plan-name",
					ExternalID:   "external-id",
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: "cluster-service-class-name",
					},
				},
			},
			json: "{\"name\":\"service-plan\",\"spec.clusterServiceClassName\":\"cluster-service-class-name\",\"spec.externalID\":\"external-id\",\"spec.externalName\":\"external-plan-name\"}",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := ConvertServicePlanToProperties(tc.sp)
			if p == nil {
				t.Fatalf("Failed to create Properties object from %+v", tc.sp)
			}
			b, err := json.Marshal(p)
			if err != nil {
				t.Fatalf("Unexpected error with json marchal, %v", err)
			}
			js := string(b)
			if js != tc.json {
				t.Fatalf("Failed to create expected Properties object,\n\texpected: \t%q,\n \tgot: \t\t%q", tc.json, js)
			}
		})
	}
}
