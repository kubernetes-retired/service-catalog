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
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetBroker is a convenience function to get a *v1alpha1.Broker for use in tests
func GetBroker() *v1alpha1.Broker {
	return &v1alpha1.Broker{
		ObjectMeta: metav1.ObjectMeta{Name: BrokerName},
		Spec: v1alpha1.BrokerSpec{
			URL: BrokerURL,
		},
	}
}

// GetServiceClass returns a ServiceClass that can be used by tests
func GetServiceClass() *v1alpha1.ServiceClass {
	return &v1alpha1.ServiceClass{
		ObjectMeta:  metav1.ObjectMeta{Name: ServiceClassName},
		BrokerName:  BrokerName,
		Description: "a test service",
		ExternalID:  ServiceClassGUID,
		Bindable:    true,
		Plans: []v1alpha1.ServicePlan{
			{
				Name:        PlanName,
				Description: "a test plan",
				Free:        true,
				ExternalID:  PlanGUID,
			},
			{
				Name:        NonBindablePlanName,
				Description: "a test plan",
				Free:        true,
				ExternalID:  NonBindablePlanGUID,
				Bindable:    boolPtr(false),
			},
		},
	}

}

// GetInstance returns an Instance that can be used by tests
func GetInstance() *v1alpha1.Instance {
	return &v1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{Name: InstanceName, Namespace: Namespace},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: ServiceClassName,
			PlanName:         PlanName,
			ExternalID:       InstanceGUID,
		},
	}
}
