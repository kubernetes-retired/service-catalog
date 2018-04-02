/*
Copyright 2017 The Kubernetes Authors.

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

package v1beta1

import (
	"encoding/json"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

type conversionFunc func(string, string) (string, string, error)

type testcase struct {
	name          string
	inLabel       string
	inValue       string
	outLabel      string
	outValue      string
	success       bool
	expectedError string
}

func TestClusterServicePlanFieldLabelConversionFunc(t *testing.T) {
	cases := []testcase{
		{
			name:     "spec.externalName works",
			inLabel:  "spec.externalName",
			inValue:  "somenamehere",
			outLabel: "spec.externalName",
			outValue: "somenamehere",
			success:  true,
		},
		{
			name:     "spec.clusterServiceClassRef.name works",
			inLabel:  "spec.clusterServiceClassRef.name",
			inValue:  "someref",
			outLabel: "spec.clusterServiceClassRef.name",
			outValue: "someref",
			success:  true,
		},
		{
			name:     "spec.clusterServiceBrokerName works",
			inLabel:  "spec.clusterServiceBrokerName",
			inValue:  "somebroker",
			outLabel: "spec.clusterServiceBrokerName",
			outValue: "somebroker",
			success:  true,
		},
		{
			name:     "spec.externalID works",
			inLabel:  "spec.externalID",
			inValue:  "externalid",
			outLabel: "spec.externalID",
			outValue: "externalid",
			success:  true,
		},
		{
			name:          "random fails",
			inLabel:       "spec.random",
			inValue:       "randomvalue",
			outLabel:      "",
			outValue:      "",
			success:       false,
			expectedError: "field label not supported: spec.random",
		},
	}

	runTestCases(t, cases, "ClusterServicePlanFieldLabelConversionFunc", ClusterServicePlanFieldLabelConversionFunc)
}

func TestServicePlanFieldLabelConversionFunc(t *testing.T) {
	cases := []testcase{
		{
			name:     "spec.externalName works",
			inLabel:  "spec.externalName",
			inValue:  "somenamehere",
			outLabel: "spec.externalName",
			outValue: "somenamehere",
			success:  true,
		},
		{
			name:     "spec.serviceClassRef.name works",
			inLabel:  "spec.serviceClassRef.name",
			inValue:  "someref",
			outLabel: "spec.serviceClassRef.name",
			outValue: "someref",
			success:  true,
		},
		{
			name:     "spec.serviceBrokerName works",
			inLabel:  "spec.serviceBrokerName",
			inValue:  "somebroker",
			outLabel: "spec.serviceBrokerName",
			outValue: "somebroker",
			success:  true,
		},
		{
			name:     "spec.externalID works",
			inLabel:  "spec.externalID",
			inValue:  "externalid",
			outLabel: "spec.externalID",
			outValue: "externalid",
			success:  true,
		},
		{
			name:          "random fails",
			inLabel:       "spec.random",
			inValue:       "randomvalue",
			outLabel:      "",
			outValue:      "",
			success:       false,
			expectedError: "field label not supported: spec.random",
		},
	}

	runTestCases(t, cases, "ServicePlanFieldLabelConversionFunc", ServicePlanFieldLabelConversionFunc)
}

func TestClusterServiceClassFieldLabelConversionFunc(t *testing.T) {
	cases := []testcase{
		{
			name:     "spec.externalName works",
			inLabel:  "spec.externalName",
			inValue:  "somenamehere",
			outLabel: "spec.externalName",
			outValue: "somenamehere",
			success:  true,
		},
		{
			name:          "spec.clusterServiceClassRef.name fails",
			inLabel:       "spec.clusterServiceClassRef.name",
			inValue:       "someref",
			outLabel:      "",
			outValue:      "",
			success:       false,
			expectedError: "field label not supported: spec.clusterServiceClassRef.name",
		},
		{
			name:     "spec.clusterServiceBrokerName works",
			inLabel:  "spec.clusterServiceBrokerName",
			inValue:  "somebroker",
			outLabel: "spec.clusterServiceBrokerName",
			outValue: "somebroker",
			success:  true,
		},
		{
			name:     "spec.externalID works",
			inLabel:  "spec.externalID",
			inValue:  "externalid",
			outLabel: "spec.externalID",
			outValue: "externalid",
			success:  true,
		},
		{
			name:          "random fails",
			inLabel:       "spec.random",
			inValue:       "randomvalue",
			outLabel:      "",
			outValue:      "",
			success:       false,
			expectedError: "field label not supported: spec.random",
		},
	}
	runTestCases(t, cases, "ClusterServiceClassFieldLabelConversionFunc", ClusterServiceClassFieldLabelConversionFunc)

}

func TestServiceInstanceFieldLabelConversionFunc(t *testing.T) {
	cases := []testcase{
		{
			name:     "spec.clusterServiceClassRef.name works",
			inLabel:  "spec.clusterServiceClassRef.name",
			inValue:  "someref",
			outLabel: "spec.clusterServiceClassRef.name",
			outValue: "someref",
			success:  true,
		},
		{
			name:     "spec.clusterServicePlanRef.name works",
			inLabel:  "spec.clusterServicePlanRef.name",
			inValue:  "someref",
			outLabel: "spec.clusterServicePlanRef.name",
			outValue: "someref",
			success:  true,
		},
		{
			name:          "random fails",
			inLabel:       "spec.random",
			inValue:       "randomvalue",
			outLabel:      "",
			outValue:      "",
			success:       false,
			expectedError: "field label not supported: spec.random",
		},
	}
	runTestCases(t, cases, "ServiceInstanceFieldLabelConversionFunc", ServiceInstanceFieldLabelConversionFunc)

}

func runTestCases(t *testing.T, cases []testcase, testFuncName string, testFunc conversionFunc) {
	for _, tc := range cases {
		outLabel, outValue, err := testFunc(tc.inLabel, tc.inValue)
		if tc.success {
			if err != nil {
				t.Errorf("%s:%s -- unexpected failure : %q", testFuncName, tc.name, err.Error())
			} else {
				if a, e := outLabel, tc.outLabel; a != e {
					t.Errorf("%s:%s -- label mismatch, expected %q got %q", testFuncName, tc.name, e, a)
				}
				if a, e := outValue, tc.outValue; a != e {
					t.Errorf("%s:%s -- value mismatch, expected %q got %q", testFuncName, tc.name, e, a)
				}
			}
		} else {
			if err == nil {
				t.Errorf("%s:%s -- unexpected success, expected: %q", testFuncName, tc.name, tc.expectedError)
			} else {
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("%s:%s -- did not find expected error %q got %q", testFuncName, tc.name, tc.expectedError, err)
				}
			}
		}
	}
}

func TestConvert_v1beta1_CatalogRestrictions_To_servicecatalog_CatalogRestrictions_AndBack(t *testing.T) {
	originalIn := CatalogRestrictions{
		ServiceClass: []string{"name not in (foo)"},
		ServicePlan:  []string{"name == bar", "externalName==baz"},
	}
	var originalOut servicecatalog.CatalogRestrictions

	Convert_v1beta1_CatalogRestrictions_To_servicecatalog_CatalogRestrictions(&originalIn, &originalOut, nil)

	var convertedOut CatalogRestrictions

	Convert_servicecatalog_CatalogRestrictions_To_v1beta1_CatalogRestrictions(&originalOut, &convertedOut, nil)

	// original in and converted out should match, but string formatting and order modifications are  allowed.

	for _, r := range convertedOut.ServiceClass {
		if !findInRequirementsIgnoreSpaces(r, originalIn.ServiceClass) {
			t.Fail()
		}
	}

	for _, r := range convertedOut.ServicePlan {
		if !findInRequirementsIgnoreSpaces(r, originalIn.ServicePlan) {
			t.Fail()
		}
	}
}

func findInRequirementsIgnoreSpaces(requirement string, requirements []string) bool {
	find := strings.Replace(requirement, " ", "", -1)
	for _, r := range requirements {
		found := strings.Replace(r, " ", "", -1)
		if find == found {
			return true
		}
	}
	return false
}

func TestConvertClusterServiceClassToProperties(t *testing.T) {
	cases := []struct {
		name string
		sc   *ClusterServiceClass
		json string
	}{
		{
			name: "nil object",
			json: "{}",
		},
		{
			name: "normal object",
			sc: &ClusterServiceClass{
				ObjectMeta: metav1.ObjectMeta{Name: "service-class"},
				Spec: ClusterServiceClassSpec{
					CommonServiceClassSpec: CommonServiceClassSpec{
						ExternalName: "external-class-name",
						ExternalID:   "external-id",
					},
				},
			},
			json: "{\"name\":\"service-class\",\"spec.externalID\":\"external-id\",\"spec.externalName\":\"external-class-name\"}",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := ConvertClusterServiceClassToProperties(tc.sc)
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

func TestConvertClusterServicePlanToProperties(t *testing.T) {
	cases := []struct {
		name string
		sp   *ClusterServicePlan
		json string
	}{
		{
			name: "nil object",
			json: "{}",
		},
		{
			name: "normal object",
			sp: &ClusterServicePlan{
				ObjectMeta: metav1.ObjectMeta{Name: "service-plan"},
				Spec: ClusterServicePlanSpec{
					CommonServicePlanSpec: CommonServicePlanSpec{
						ExternalName: "external-plan-name",
						ExternalID:   "external-id",
					},
					ClusterServiceClassRef: ClusterObjectReference{
						Name: "cluster-service-class-name",
					},
				},
			},
			json: "{\"name\":\"service-plan\",\"spec.clusterServiceClass.name\":\"cluster-service-class-name\",\"spec.externalID\":\"external-id\",\"spec.externalName\":\"external-plan-name\"}",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := ConvertClusterServicePlanToProperties(tc.sp)
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
