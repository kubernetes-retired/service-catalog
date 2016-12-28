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

package controller

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
)

// Reifier is an interface which manipulates deployment resources to achieve
// a desired final state.
type Reifier interface {
	// Catalog returns all the available Services that can be instantiated
	Catalog() ([]*brokerapi.Service, error)

	// RemoveServiceInstance removes an existing Service Instance
	RemoveServiceInstance(instanceID string) error

	// CreateServiceInstance creates a new Service Instance
	CreateServiceInstance(instanceID string, template string, sir *brokerapi.ServiceInstanceRequest) (*brokerapi.CreateServiceInstanceResponse, error)

	// CreateBinding creates a new Service Binding for a given instanceId
	CreateServiceBinding(instanceID string, sir *brokerapi.BindingRequest) (*brokerapi.CreateServiceBindingResponse, error)

	// RemoveServiceBinding removes an existing Service Binding
	RemoveServiceBinding(instanceID string) error
}
