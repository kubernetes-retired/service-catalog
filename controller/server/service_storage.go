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

package server

import (
	model "github.com/kubernetes-incubator/service-catalog/model/service_controller"
)

type BindingDirection int

const (
	To BindingDirection = iota
	From
	Both
)

// The Broker interface provides functions to deal with brokers.
type Broker interface {
	ListBrokers() ([]*model.ServiceBroker, error)
	GetBroker(string) (*model.ServiceBroker, error)
	GetBrokerByService(string) (*model.ServiceBroker, error)
	GetInventory() (*model.Catalog, error)
	AddBroker(*model.ServiceBroker, *model.Catalog) error
	UpdateBroker(*model.ServiceBroker, *model.Catalog) error
	DeleteBroker(string) error
}

type ServiceTyper interface {
	GetServiceType(string) (*model.Service, error)
}

// The Instancer interface provides functions to deal with service instances.
type Instancer interface {
	ListServices(string) ([]*model.ServiceInstance, error)
	GetService(string, string) (*model.ServiceInstance, error)
	ServiceExists(string, string) bool
	AddService(*model.ServiceInstance) error
	SetService(*model.ServiceInstance) error
	DeleteService(string) error
}

// The Binder interface provides functions to deal with service
// bindings.
type Binder interface {
	ListServiceBindings() ([]*model.ServiceBinding, error)
	GetServiceBinding(string) (*model.ServiceBinding, error)
	AddServiceBinding(*model.ServiceBinding, *model.Credential) error
	UpdateServiceBinding(*model.ServiceBinding) error
	DeleteServiceBinding(string) error
	// GetBindingsForService returns bindings for a given service instance and direction.
	GetBindingsForService(string, BindingDirection) ([]*model.ServiceBinding, error)
}

// The ServiceStorage interface provides a comprehensive combined
// resource for end to end dealings with service brokers, service instances,
// and service bindings.
type ServiceStorage interface {
	Broker
	ServiceTyper
	Instancer
	Binder
}
