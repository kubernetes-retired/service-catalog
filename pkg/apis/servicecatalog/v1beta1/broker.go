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

// GetName returns the broker's name.
func (b *ClusterServiceBroker) GetName() string {
	return b.Name
}

// GetName returns the broker's name.
func (b *ServiceBroker) GetName() string {
	return b.Name
}

// GetNamespace always returns "", because it's cluster-scoped.
func (b *ClusterServiceBroker) GetNamespace() string {
	return ""
}

// GetNamespace returns the broker's namespace.
func (b *ServiceBroker) GetNamespace() string {
	return b.Namespace
}

// GetURL returns the broker's endpoint URL.
func (b *ClusterServiceBroker) GetURL() string {
	return b.Spec.URL
}

// GetURL returns the broker's endpoint URL.
func (b *ServiceBroker) GetURL() string {
	return b.Spec.URL
}

// GetStatus returns the broker status.
func (b *ClusterServiceBroker) GetStatus() CommonServiceBrokerStatus {
	return b.Status.CommonServiceBrokerStatus
}

// GetStatus returns the broker status.
func (b *ServiceBroker) GetStatus() CommonServiceBrokerStatus {
	return b.Status.CommonServiceBrokerStatus
}
