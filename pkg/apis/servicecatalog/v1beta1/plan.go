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

// GetName returns the plan's name.
func (p *ClusterServicePlan) GetName() string {
	return p.Name
}

// GetName returns the plan's name.
func (p *ServicePlan) GetName() string {
	return p.Name
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
