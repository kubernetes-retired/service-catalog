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

package server

import (
	pkgbrokerapi "github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/pivotal-cf/brokerapi"
)

// ConvertCatalog converts a (github.com/kubernetes-incubator/service-catalog/pkg/brokerapi).Catalog
// to an array of brokerapi.Services
func ConvertCatalog(cat *pkgbrokerapi.Catalog) []brokerapi.Service {
	ret := make([]brokerapi.Service, len(cat.Services))
	for i, svc := range cat.Services {
		ret[i] = convertService(svc)
	}
	return ret
}

func convertService(svc *pkgbrokerapi.Service) brokerapi.Service {
	return brokerapi.Service{
		ID:            svc.ID,
		Name:          svc.Name,
		Description:   svc.Description,
		Bindable:      svc.Bindable,
		Tags:          svc.Tags,
		PlanUpdatable: svc.PlanUpdateable,
		Plans:         convertPlans(svc.Plans),
		// TODO: convert Requires, Metadata, DashboardClient
	}
}

func convertPlans(plans []pkgbrokerapi.ServicePlan) []brokerapi.ServicePlan {
	ret := make([]brokerapi.ServicePlan, len(plans))
	for i, plan := range plans {
		ret[i] = brokerapi.ServicePlan{
			ID:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
			Free:        &plan.Free,
			Bindable:    plan.Bindable,
			// TODO: convert Metadata
		}
	}
	return ret
}
