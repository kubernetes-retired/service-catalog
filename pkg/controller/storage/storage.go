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
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
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
	// List returns a list of all brokers
	List() ([]*servicecatalog.Broker, error)

	// Get gets a broker by name. Returns error if broker does not exist.
	Get(name string) (*servicecatalog.Broker, error)

	// Create adds a new broker. Returns an error if a broker with the given name already exists
	// a broker of this name already exists.
	Create(*servicecatalog.Broker) (*servicecatalog.Broker, error)

	// Update updates an existing broker. Returns error if the broker doesn't exist
	Update(*servicecatalog.Broker) (*servicecatalog.Broker, error)

	// Delete deletes an existing broker by name. Returns error if broker does not exist.
	Delete(name string) error
}

// ServiceClassStorage defines the interface to manage service classes.
type ServiceClassStorage interface {
	// List returns all service classes
	List() ([]*servicecatalog.ServiceClass, error)

	// Get returns a service class by name. Returns error if the class doesn't exist
	Get(name string) (*servicecatalog.ServiceClass, error)

	// Create adds a new service class. Returns error if a service class of this
	// name already exists.
	Create(*servicecatalog.ServiceClass) (*servicecatalog.ServiceClass, error)
}

// InstanceStorage defines the interface to manage service instances.
type InstanceStorage interface {
	// ListServiceInstances returns all service instances
	List() ([]*servicecatalog.Instance, error)

	// Get fetches a service instance by name. Returns error if instance does not exist
	Get(name string) (*servicecatalog.Instance, error)

	// Create adds a new service instance. Returns error if an instance of this name already exists.
	Create(*servicecatalog.Instance) (*servicecatalog.Instance, error)

	// Update updates an existing service instance. Returns error if instance does not exist.
	Update(*servicecatalog.Instance) (*servicecatalog.Instance, error)

	// Delete deletes an existing service instance. Returns error if instance does not exist.
	Delete(name string) error
}

// BindingStorage defines the interface manage service bindings.
type BindingStorage interface {
	// List returns all bindings.
	List() ([]*servicecatalog.Binding, error)

	// Get gets a binding by name. Returns error if binding does not exist.
	Get(name string) (*servicecatalog.Binding, error)

	// Create adds a new binding. Returns error if a binding of this name already exists.
	Create(*servicecatalog.Binding) (*servicecatalog.Binding, error)

	// Update updates an existing binding. Returns error if binding does not exist.
	Update(*servicecatalog.Binding) (*servicecatalog.Binding, error)

	// Delete deletes an existing binding. Returns error if binding does not exist.
	Delete(name string) error
}

// Storage defines the interface to manage service brokers, types, instances, and bindings.
type Storage interface {
	Brokers() BrokerStorage
	ServiceClasses() ServiceClassStorage
	Instances(string) InstanceStorage
	Bindings(string) BindingStorage
}
