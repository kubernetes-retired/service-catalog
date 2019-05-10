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

const (
	statusActive     = "Active"
	statusDeprecated = "Deprecated"
)

// GetName returns the class's name.
func (c *ClusterServiceClass) GetName() string {
	return c.Name
}

// GetName returns the class's name.
func (c *ServiceClass) GetName() string {
	return c.Name
}

// GetNamespace for cluster-scoped classes always returns "".
func (c *ClusterServiceClass) GetNamespace() string {
	return ""
}

// GetNamespace returns the class's namespace.
func (c *ServiceClass) GetNamespace() string {
	return c.Namespace
}

// GetExternalName returns the class's external name.
func (c *ClusterServiceClass) GetExternalName() string {
	return c.Spec.ExternalName
}

// GetExternalName returns the class's external name.
func (c *ServiceClass) GetExternalName() string {
	return c.Spec.ExternalName
}

// GetDescription returns the class description.
func (c *ClusterServiceClass) GetDescription() string {
	return c.Spec.Description
}

// GetDescription returns the class description.
func (c *ServiceClass) GetDescription() string {
	return c.Spec.Description
}

// GetSpec returns the spec for the class.
func (c *ServiceClass) GetSpec() CommonServiceClassSpec {
	return c.Spec.CommonServiceClassSpec
}

// GetSpec returns the spec for the class.
func (c *ClusterServiceClass) GetSpec() CommonServiceClassSpec {
	return c.Spec.CommonServiceClassSpec
}

// GetServiceBrokerName returns the name of the service broker for the class.
func (c *ServiceClass) GetServiceBrokerName() string {
	return c.Spec.ServiceBrokerName
}

// GetServiceBrokerName returns the name of the service broker for the class.
func (c *ClusterServiceClass) GetServiceBrokerName() string {
	return c.Spec.ClusterServiceBrokerName
}

// GetStatusText returns the status of the class.
func (c *ServiceClass) GetStatusText() string {
	return c.Status.GetStatusText()
}

// GetStatusText returns the status of the class.
func (c *ClusterServiceClass) GetStatusText() string {
	return c.Status.GetStatusText()
}

// GetStatusText returns the status based on the CommonServiceClassStatus.
func (c *CommonServiceClassStatus) GetStatusText() string {
	if c.RemovedFromBrokerCatalog {
		return statusDeprecated
	}
	return statusActive
}

// IsClusterServiceClass returns true for ClusterServiceClasses
func (c *ClusterServiceClass) IsClusterServiceClass() bool {
	return true
}

// IsClusterServiceClass returns false for ServiceClasses
func (c *ServiceClass) IsClusterServiceClass() bool {
	return false
}
