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
