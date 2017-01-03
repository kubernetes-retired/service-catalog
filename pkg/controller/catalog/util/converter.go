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
	sbmodel "github.com/kubernetes-incubator/service-catalog/model/service_broker"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

func convertServicePlans(plans []sbmodel.ServicePlan) []servicecatalog.ServicePlan {
	ret := make([]servicecatalog.ServicePlan, len(plans))
	for i, plan := range plans {
		ret[i] = servicecatalog.ServicePlan{
			Name:       plan.Name,
			CFGUID:     plan.ID,
			CFMetadata: plan.Metadata,
			CFFree:     plan.Free,
		}
	}
	return ret
}

// ConvertCatalog converts a service broker catalog into an array of ServiceClasses
func ConvertCatalog(in *sbmodel.Catalog) ([]*servicecatalog.ServiceClass, error) {
	ret := make([]*servicecatalog.ServiceClass, len(in.Services))
	for i, svc := range in.Services {
		plans := convertServicePlans(svc.Plans)
		ret[i] = &servicecatalog.ServiceClass{
			Bindable:      svc.Bindable,
			Plans:         plans,
			PlanUpdatable: svc.PlanUpdateable,
			CFGUID:        svc.ID,
			CFTags:        svc.Tags,
			CFRequires:    svc.Requires,
			CFMetadata:    svc.Metadata,
		}
	}
	return ret, nil
}
