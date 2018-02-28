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
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
)

func CreatePredicateForServiceClasses(requirements v1beta1.ClusterServiceClassRequirements) (Predicate, error) {
	selector, err := labels.Parse(string(requirements))
	if err != nil {
		return nil, err
	}
	predicate := internalPredicate{selector: selector}
	return predicate, nil
}

func CreatePredicateForServicePlans(requirements v1beta1.ClusterServicePlanRequirements) (Predicate, error) {
	selector, err := labels.Parse(string(requirements))
	if err != nil {
		return nil, err
	}
	predicate := internalPredicate{selector: selector}
	return predicate, nil
}

func ConvertServiceClassToProperties(serviceClass *v1beta1.ClusterServiceClass) Properties {
	return labels.Set{
		Name:         serviceClass.Name,
		ExternalName: serviceClass.Spec.ExternalName,
		ExternalID:   serviceClass.Spec.ExternalID,
	}
}

func ConvertServicePlanToProperties(servicePlan *v1beta1.ClusterServicePlan) Properties {
	return labels.Set{
		Name:                    servicePlan.Name,
		ExternalName:            servicePlan.Spec.ExternalName,
		ExternalID:              servicePlan.Spec.ExternalID,
		ClusterServiceClassName: servicePlan.Spec.ClusterServiceClassRef.Name,
	}
}
