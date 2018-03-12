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
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
)

// CreatePredicateForServiceClasses creates the Predicate that will be used to
// test if acceptance is allowed for service classes.
func createPredicateForServiceClasses(requirements string) (Predicate, error) {
	selector, err := labels.Parse(requirements)
	if err != nil {
		return nil, err
	}
	predicate := internalPredicate{selector: selector}
	return predicate, nil
}

func CreatePredicateForServiceClassesFromRestrictions(restrictions *v1beta1.ServiceClassCatalogRestrictions) (Predicate, error) {
	if restrictions != nil && len(restrictions.ServiceClass) > 0 {
		// Flatten the requirements into a selector string.
		requirements := string(restrictions.ServiceClass[0])
		for i := 1; i < len(restrictions.ServiceClass); i++ {
			requirements = fmt.Sprintf("%s, %s", requirements, string(restrictions.ServiceClass[i]))
		}
		return createPredicateForServiceClasses(requirements)
	} else {
		return createPredicateForServiceClasses("")
	}
}

// CreatePredicateForServicePlans creates the Predicate that will be used to
// test if acceptance is allowed for service plans.
func createPredicateForServicePlans(requirements string) (Predicate, error) {
	selector, err := labels.Parse(string(requirements))
	if err != nil {
		return nil, err
	}
	predicate := internalPredicate{selector: selector}
	return predicate, nil
}

func CreatePredicateForServicePlansFromRestrictions(restrictions *v1beta1.ServiceClassCatalogRestrictions) (Predicate, error) {
	if restrictions != nil && len(restrictions.ServicePlan) > 0 {
		// Flatten the requirements into a selector string.
		requirements := string(restrictions.ServicePlan[0])
		for i := 1; i < len(restrictions.ServicePlan); i++ {
			requirements = fmt.Sprintf("%s, %s", requirements, string(restrictions.ServicePlan[i]))
		}
		return createPredicateForServicePlans(requirements)
	} else {
		return createPredicateForServicePlans("")
	}
}

// ConvertServiceClassToProperties takes a Service Class and pulls out the
// properties we support for filtering, converting them into a map in the
// expected format.
func ConvertServiceClassToProperties(serviceClass *v1beta1.ClusterServiceClass) Properties {
	if serviceClass == nil {
		return labels.Set{}
	}
	return labels.Set{
		Name:             serviceClass.Name,
		SpecExternalName: serviceClass.Spec.ExternalName,
		SpecExternalID:   serviceClass.Spec.ExternalID,
	}
}

// ConvertServicePlanToProperties takes a Service Plan and pulls out the
// properties we support for filtering, converting them into a map in the
// expected format.
func ConvertServicePlanToProperties(servicePlan *v1beta1.ClusterServicePlan) Properties {
	if servicePlan == nil {
		return labels.Set{}
	}
	return labels.Set{
		Name:                        servicePlan.Name,
		SpecExternalName:            servicePlan.Spec.ExternalName,
		SpecExternalID:              servicePlan.Spec.ExternalID,
		SpecClusterServiceClassName: servicePlan.Spec.ClusterServiceClassRef.Name,
	}
}
