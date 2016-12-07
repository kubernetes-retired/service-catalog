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

package storage

import (
	model "github.com/kubernetes-incubator/service-catalog/model/service_controller"
)

// BindingDirection is an integer type used for expressing direction of the
// service binding (from, to, or both).
type BindingDirection int

const (
	// To represents the 'to' direction of the service binding.
	To BindingDirection = iota
	// From represents the 'from' direction of the service binding.
	From
	// Both represents both directions of the service binding.
	Both
)

// BrokerStorage defines the interface to manage brokers.
type BrokerStorage interface {
	// ListBrokers returns all brokers.
	ListBrokers() ([]*model.ServiceBroker, error)

	// GetBroker gets a broker by name. Returns error if broker does not exist.
	GetBroker(name string) (*model.ServiceBroker, error)

	// AddBroker adds a new broker with its associated catalog. Returns error if
	// a broker of this name already exists.
	AddBroker(*model.ServiceBroker, *model.Catalog) error

	// UpdateBroker updates an existing broker with its associated catalog.
	// Returns error if broker does not exist.
	UpdateBroker(*model.ServiceBroker, *model.Catalog) error

	// DeleteBroker deletes an existing broker by name. Returns error if broker
	// does not exist.
	DeleteBroker(name string) error
}

// ClassStorage defines the interface to manage service classes.
type ClassStorage interface {
	// GetInventory returns the aggregate catalog of service classes across all
	// brokers.
	GetInventory() (*model.Catalog, error)

	// GetServiceClass returns the definition of a service class by name. Returns
	// error if service class does not exist.
	GetServiceClass(name string) (*model.Service, error)
}

// InstanceStorage defines the interface to manage service instances.
type InstanceStorage interface {
	// ListServiceInstances returns all service instances.
	ListServiceInstances(namespace string) ([]*model.ServiceInstance, error)

	// GetServiceInstance gets a service instance by name. Returns error if instance does
	// not exist.
	GetServiceInstance(namespace string, name string) (*model.ServiceInstance, error)

	// ServiceInstanceExists returns whether a service instance exists.
	ServiceInstanceExists(namespace string, name string) bool

	// AddServiceInstance adds a new service instance. Returns error if an instance of
	// this name already exists.
	AddServiceInstance(*model.ServiceInstance) error

	// UpdateServiceInstance updates an existing service instance. Returns error if instance
	// does not exist.
	UpdateServiceInstance(*model.ServiceInstance) error

	// DeleteServiceInstance deletes an existing service instance. Returns error if
	// instance does not exist.
	DeleteServiceInstance(name string) error
}

// BindingStorage defines the interface manage service bindings.
type BindingStorage interface {
	// ListServiceBindings returns all service bindings.
	ListServiceBindings() ([]*model.ServiceBinding, error)

	// GetServiceBinding gets a service binding by name. Returns error if binding
	// does not exist.
	GetServiceBinding(name string) (*model.ServiceBinding, error)

	// AddServiceBinding adds a new service binding with its associated
	// credentials. Returns error if a binding of this name already exists.
	AddServiceBinding(*model.ServiceBinding, *model.Credential) error

	// UpdateServiceBinding updates an existing service binding. Returns error if
	// binding does not exist.
	UpdateServiceBinding(*model.ServiceBinding) error

	// DeleteServiceBinding deletes an existing service binding. Returns error if
	// binding does not exist.
	DeleteServiceBinding(name string) error
}

// Storage defines the interface to manage service brokers, types, instances,
// and bindings.
type Storage interface {
	BrokerStorage
	ClassStorage
	InstanceStorage
	BindingStorage
}
