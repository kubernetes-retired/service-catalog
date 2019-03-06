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

package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// GetName returns the plan's name.
func (p *ClusterServicePlan) GetName() string {
	return p.Name
}

// GetName returns the plan's name.
func (p *ServicePlan) GetName() string {
	return p.Name
}

// GetNamespace for cluster-scoped plans always returns "".
func (p *ClusterServicePlan) GetNamespace() string {
	return ""
}

// GetNamespace returns the plan's namespace.
func (p *ServicePlan) GetNamespace() string {
	return p.Namespace
}

// GetShortStatus returns the plan's status.
func (p *ClusterServicePlan) GetShortStatus() string {
	if p.Status.RemovedFromBrokerCatalog {
		return "Deprecated"
	}
	return "Active"
}

// GetShortStatus returns the plan's status.
func (p *ServicePlan) GetShortStatus() string {
	if p.Status.RemovedFromBrokerCatalog {
		return "Deprecated"
	}
	return "Active"
}

// GetExternalName returns the plan's external name.
func (p *ClusterServicePlan) GetExternalName() string {
	return p.Spec.ExternalName
}

// GetExternalName returns the plan's external name.
func (p *ServicePlan) GetExternalName() string {
	return p.Spec.ExternalName
}

// GetDescription returns the plan description.
func (p *ClusterServicePlan) GetDescription() string {
	return p.Spec.Description
}

// GetDescription returns the plan description.
func (p *ServicePlan) GetDescription() string {
	return p.Spec.Description
}

// GetFree returns if the plan is free.
func (p *ClusterServicePlan) GetFree() bool {
	return p.Spec.Free
}

// GetFree returns if the plan is free.
func (p *ServicePlan) GetFree() bool {
	return p.Spec.Free
}

// GetClassID returns the class name from plan.
func (p *ClusterServicePlan) GetClassID() string {
	return p.Spec.ClusterServiceClassRef.Name
}

// GetClassID returns the class name from plan.
func (p *ServicePlan) GetClassID() string {
	return p.Spec.ServiceClassRef.Name
}

// GetDefaultProvisionParameters returns the default provision parameters from plan.
func (p *ClusterServicePlan) GetDefaultProvisionParameters() *runtime.RawExtension {
	return p.Spec.DefaultProvisionParameters
}

// GetDefaultProvisionParameters returns the default provision parameters from plan.
func (p *ServicePlan) GetDefaultProvisionParameters() *runtime.RawExtension {
	return p.Spec.DefaultProvisionParameters
}

// GetInstanceCreateSchema returns the instance create schema from plan.
func (p *ClusterServicePlan) GetInstanceCreateSchema() *runtime.RawExtension {
	return p.Spec.InstanceCreateParameterSchema
}

// GetInstanceCreateSchema returns the instance create schema from plan.
func (p *ServicePlan) GetInstanceCreateSchema() *runtime.RawExtension {
	return p.Spec.InstanceCreateParameterSchema
}

// GetInstanceUpdateSchema returns the instance update schema from plan.
func (p *ClusterServicePlan) GetInstanceUpdateSchema() *runtime.RawExtension {
	return p.Spec.InstanceUpdateParameterSchema
}

// GetInstanceUpdateSchema returns the instance update schema from plan.
func (p *ServicePlan) GetInstanceUpdateSchema() *runtime.RawExtension {
	return p.Spec.InstanceUpdateParameterSchema
}

// GetBindingCreateSchema returns the instance create schema from plan.
func (p *ClusterServicePlan) GetBindingCreateSchema() *runtime.RawExtension {
	return p.Spec.ServiceBindingCreateParameterSchema
}

// GetBindingCreateSchema returns the instance create schema from plan.
func (p *ServicePlan) GetBindingCreateSchema() *runtime.RawExtension {
	return p.Spec.ServiceBindingCreateParameterSchema
}
