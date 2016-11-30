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

package util

import (
	model "github.com/kubernetes-incubator/service-catalog/model/service_broker"
)

// BrokerClient defines the interface for interacting with a broker for catalog
// retrieval, service instance management, and service binding management.
type BrokerClient interface {
	CatalogClient
	InstanceClient
	BindingClient
}

// CatalogClient defines the interface for catalog interaction with a broker.
type CatalogClient interface {
	GetCatalog() (*model.Catalog, error)
}

// InstanceClient defines the interface for managing service instances with a
// broker.
type InstanceClient interface {
	// TODO: these should return appropriate response objects.
	CreateServiceInstance(ID string, req *model.ServiceInstanceRequest) (*model.ServiceInstance, error)
	UpdateServiceInstance(ID string, req *model.ServiceInstanceRequest) (*model.ServiceInstance, error)

	DeleteServiceInstance(ID string) error
}

// BindingClient defines the interface for managing service bindings with a
// broker.
type BindingClient interface {
	CreateServiceBinding(sID, bID string, req *model.BindingRequest) (*model.CreateServiceBindingResponse, error)
	DeleteServiceBinding(sID, bID string) error
}
