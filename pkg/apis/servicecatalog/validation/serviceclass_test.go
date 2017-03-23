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

package validation

import (
	"testing"

	kapi "k8s.io/kubernetes/pkg/api"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

func TestValidateServiceClass(t *testing.T) {
	cases := []struct {
		name         string
		serviceClass *servicecatalog.ServiceClass
		valid        bool
	}{
		{
			name: "valid serviceClass",
			serviceClass: &servicecatalog.ServiceClass{
				ObjectMeta: kapi.ObjectMeta{
					Name: "test-serviceclass",
				},
				Bindable:   true,
				BrokerName: "test-broker",
				OSBGUID:    "1234-4354a-49b",
				Plans: []servicecatalog.ServicePlan{
					{
						Name:    "test-plan",
						OSBGUID: "40d-0983-1b89",
					},
				},
			},
			valid: true,
		},
		{
			name: "valid serviceClass - plan with underscore in name",
			serviceClass: &servicecatalog.ServiceClass{
				ObjectMeta: kapi.ObjectMeta{
					Name: "test-serviceclass",
				},
				Bindable:   true,
				BrokerName: "test-broker",
				OSBGUID:    "1234-4354a-49b",
				Plans: []servicecatalog.ServicePlan{
					{
						Name:    "test_plan",
						OSBGUID: "40d-0983-1b89",
					},
				},
			},
			valid: true,
		},
		{
			name: "valid serviceClass - uppercase in GUID",
			serviceClass: &servicecatalog.ServiceClass{
				ObjectMeta: kapi.ObjectMeta{
					Name: "test-serviceclass",
				},
				Bindable:   true,
				BrokerName: "test-broker",
				OSBGUID:    "1234-4354a-49b",
				Plans: []servicecatalog.ServicePlan{
					{
						Name:    "test-plan",
						OSBGUID: "40D-0983-1b89",
					},
				},
			},
			valid: true,
		},
		{
			name: "invalid serviceClass - has namespace",
			serviceClass: &servicecatalog.ServiceClass{
				ObjectMeta: kapi.ObjectMeta{
					Name:      "test-serviceclass",
					Namespace: "test-ns",
				},
				Bindable:   true,
				BrokerName: "test-broker",
				OSBGUID:    "1234-4354a-49b",
				Plans: []servicecatalog.ServicePlan{
					{
						Name:    "test-plan",
						OSBGUID: "40d-0983-1b89",
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid serviceClass - missing guid",
			serviceClass: &servicecatalog.ServiceClass{
				ObjectMeta: kapi.ObjectMeta{
					Name: "test-serviceclass",
				},
				Bindable:   true,
				BrokerName: "test-broker",
				Plans: []servicecatalog.ServicePlan{
					{
						Name:    "test-plan",
						OSBGUID: "40d-0983-1b89",
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid serviceClass - invalid guid",
			serviceClass: &servicecatalog.ServiceClass{
				ObjectMeta: kapi.ObjectMeta{
					Name: "test-serviceclass",
				},
				Bindable:   true,
				BrokerName: "test-broker",
				OSBGUID:    "1234-4354a\\%-49b",
				Plans: []servicecatalog.ServicePlan{
					{
						Name:    "test-plan",
						OSBGUID: "40d-0983-1b89",
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid serviceClass - invalid plan name",
			serviceClass: &servicecatalog.ServiceClass{
				ObjectMeta: kapi.ObjectMeta{
					Name: "test-serviceclass",
				},
				Bindable:   true,
				BrokerName: "test-broker",
				OSBGUID:    "1234-4354a-49b",
				Plans: []servicecatalog.ServicePlan{
					{
						Name:    "test-plan.oops",
						OSBGUID: "40d-0983-1b89",
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid serviceClass - invalid plan guid",
			serviceClass: &servicecatalog.ServiceClass{
				ObjectMeta: kapi.ObjectMeta{
					Name: "test-serviceclass",
				},
				Bindable:   true,
				BrokerName: "test-broker",
				OSBGUID:    "1234-4354a-49b",
				Plans: []servicecatalog.ServicePlan{
					{
						Name:    "test-plan",
						OSBGUID: "40d-0983-1b89-★",
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid serviceClass - missing plan guid",
			serviceClass: &servicecatalog.ServiceClass{
				ObjectMeta: kapi.ObjectMeta{
					Name: "test-serviceclass",
				},
				Bindable:   true,
				BrokerName: "test-broker",
				OSBGUID:    "1234-4354a-49b",
				Plans: []servicecatalog.ServicePlan{
					{
						Name: "test-plan",
					},
				},
			},
			valid: false,
		},
	}

	for _, tc := range cases {
		errs := ValidateServiceClass(tc.serviceClass)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}
}
