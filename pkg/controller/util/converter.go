/*
Copyright 2016 The Kubernetes Authors.

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

package util

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
)

// TODO: uncomment metadata fields once those fields are corrected and made round-trippable.

func convertServicePlans(plans []brokerapi.ServicePlan) []servicecatalog.ServicePlan {
	ret := make([]servicecatalog.ServicePlan, len(plans))
	for i, plan := range plans {
		ret[i] = servicecatalog.ServicePlan{
			Name:    plan.Name,
			OSBGUID: plan.ID,
			// OSBMetadata: plan.Metadata,
			OSBFree: plan.Free,
		}
	}
	return ret
}

// ConvertCatalog converts a service broker catalog into an array of ServiceClasses
func ConvertCatalog(in *brokerapi.Catalog) ([]*servicecatalog.ServiceClass, error) {
	ret := make([]*servicecatalog.ServiceClass, len(in.Services))
	for i, svc := range in.Services {
		plans := convertServicePlans(svc.Plans)
		ret[i] = &servicecatalog.ServiceClass{
			Bindable:      svc.Bindable,
			Plans:         plans,
			PlanUpdatable: svc.PlanUpdateable,
			OSBGUID:       svc.ID,
			OSBTags:       svc.Tags,
			OSBRequires:   svc.Requires,
			// OSBMetadata:   svc.Metadata,
		}
	}
	return ret, nil
}
