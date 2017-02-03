/*
Copyright 2017 The Kubernetes Authors.

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

package tpr

import (
	"fmt"
	"strings"
)

const (
	groupName = "servicecatalog.k8s.io"
)

func withGroupName(name string) string {
	return fmt.Sprintf("%s.%s", name, groupName)
}

// Kind represents the kind of a third party resource. This type implements fmt.Stringer
type Kind string

// String is the fmt.Stringer interface implementation
func (k Kind) String() string {
	return string(k)
}

// TPRName returns the lowercase name, suitable for fetching resources of this kind
func (k Kind) TPRName() string {
	return strings.ToLower(k.String())
}

const (
	// ServiceBrokerKind is the name of a Service Broker resource, a Kubernetes third party resource.
	ServiceBrokerKind Kind = "Broker"

	// ServiceBrokerListKind is the name of a list of Service Broker resources
	ServiceBrokerListKind Kind = "BrokerList"

	// ServiceBindingKind is the name of a Service Binding resource, a Kubernetes third party resource.
	ServiceBindingKind Kind = "Binding"

	// ServiceBindingListKind is the name for lists of Service Bindings
	ServiceBindingListKind Kind = "BindingList"

	// ServiceClassKind is the name of a Service Class resource, a Kubernetes third party resource.
	ServiceClassKind Kind = "ServiceClass"

	// ServiceClassListKind is the name of a list of service class resources
	ServiceClassListKind Kind = "ServiceClassList"

	// ServiceInstanceKind is the name of a Service Instance resource, a Kubernetes third party resource.
	ServiceInstanceKind Kind = "Instance"

	// ServiceInstanceListKind is the name of a list of service instance resources
	ServiceInstanceListKind Kind = "InstanceList"
)
