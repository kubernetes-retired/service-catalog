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

package fake

import (
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

const (
	// ServiceClassName is the static name for service classes in test data
	ServiceClassName = "testserviceclass"
	// ServiceClassGUID is the static guid for service classes in test data
	ServiceClassGUID = "testserviceclassGUID"
	// PlanName is the static name for plans in test data
	PlanName = "testplan"
	// PlanGUID is the static name for plan GUIDs in test data
	PlanGUID = "testPlanGUID"
	// NonBindablePlanName is the static name for non-bindable plans in test data
	NonBindablePlanName = "testNonBindablePlan"
	// NonBindablePlanGUID is the static GUID for non-bindable plans in test data
	NonBindablePlanGUID = "testNonBinablePlanGUID"
)

func boolPtr(b bool) *bool {
	return &b
}

// GetTestCatalog returns a static osb.CatalogResponse for use in testing
func GetTestCatalog() *osb.CatalogResponse {
	return &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:        ServiceClassName,
				ID:          ServiceClassGUID,
				Description: "a test service",
				Bindable:    true,
				Plans: []osb.Plan{
					{
						Name:        PlanName,
						Free:        boolPtr(true),
						ID:          PlanGUID,
						Description: "a test plan",
					},
					{
						Name:        NonBindablePlanName,
						Free:        boolPtr(true),
						ID:          NonBindablePlanGUID,
						Description: "a test plan",
						Bindable:    boolPtr(false),
					},
				},
			},
		},
	}
}
