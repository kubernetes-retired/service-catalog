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
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/types"
)

// The ServiceInstanceExpansion interface allows setting the References
// to ServiceClasses and ServicePlans.
type ServiceInstanceExpansion interface {
	UpdateReferences(serviceInstance *v1beta1.ServiceInstance) (*v1beta1.ServiceInstance, error)
}

func (c *serviceInstances) UpdateReferences(serviceInstance *v1beta1.ServiceInstance) (result *v1beta1.ServiceInstance, err error) {
	result = &v1beta1.ServiceInstance{}

	// TODO(mszostok): replace the subresource "resource" with custom patch
	// This is a temporary fix, to make the POC running  - https://github.com/kyma-project/kyma/issues/2836
	type serviceInstanceSpecRefPatch struct {
		ClusterServiceClassRef *v1beta1.ClusterObjectReference `json:"clusterServiceClassRef,omitempty"`
		ClusterServicePlanRef  *v1beta1.ClusterObjectReference `json:"clusterServicePlanRef,omitempty"`
		ServiceClassRef        *v1beta1.LocalObjectReference   `json:"serviceClassRef,omitempty"`
		ServicePlanRef         *v1beta1.LocalObjectReference   `json:"servicePlanRef,omitempty"`
	}
	type serviceInstanceRefPatch struct {
		Spec serviceInstanceSpecRefPatch `json:"spec"`
	}

	patchedSvc := serviceInstanceRefPatch{
		Spec: serviceInstanceSpecRefPatch{

			serviceInstance.Spec.ClusterServiceClassRef,
			serviceInstance.Spec.ClusterServicePlanRef,
			serviceInstance.Spec.ServiceClassRef,
			serviceInstance.Spec.ServicePlanRef,
		},
	}

	encoded, err := json.Marshal(patchedSvc)
	if err != nil {
		return result, err
	}

	err = c.client.Patch(types.MergePatchType).
		Namespace(serviceInstance.Namespace).
		Resource("serviceinstances").
		Name(serviceInstance.Name).
		Body(encoded).
		Do().
		Into(result)

	return
}
