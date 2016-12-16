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
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

// ServiceController defines the interface that either the HTTP server
// or the native kubernetes API handler will call to actually operate
// on the underlying resources.
type ServiceController interface {
	// CreateServiceInstance takes in a (possibly incomplete)
	// ServiceInstance and will either create or update an
	// existing one.
	CreateServiceInstance(*servicecatalog.Instance) (*servicecatalog.Instance, error)

	// CreateServiceBinding takes in a (possibly incomplete)
	// ServiceBinding and will either create or update an
	// existing one.
	CreateServiceBinding(*servicecatalog.Binding) (*servicecatalog.Binding, error)

	// CreateServiceBroker takes in a (possibly incomplete)
	// ServiceBroker and will either create or update an
	// existing one.
	CreateServiceBroker(*servicecatalog.Broker) (*servicecatalog.Broker, error)
}
