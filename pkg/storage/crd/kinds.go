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

package crd

// Kind represents the kind of a custom resource. This type implements fmt.Stringer
type Kind string

// String is the fmt.Stringer interface implementation
func (k Kind) String() string {
	return string(k)
}

// ResourcePlural represents the plural name of a custom resource. This type implements fmt.Stringer
type ResourcePlural string

// String is the fmt.Stringer interface implementation
func (k ResourcePlural) String() string {
	return string(k)
}

const (
	// ServiceBrokerKind is the name of a Service Broker resource, a Kubernetes third party resource.
	ServiceBrokerKind Kind = "ServiceBroker"

	// ServiceBrokerListKind is the name of a list of Service Broker resources
	ServiceBrokerListKind Kind = "ServiceBrokerList"

	// ServiceBrokerResourcePlural is the plural name of a Service Broker resource, a Kubernetes third party resource.
	// TODO (nilebox): Given that now we have this flexibility, shall we prefix all CRD-backed resources with something to avoid clashing in `kubectl`
	// through API aggregator? For example, crd-servicebrokers
	// Unfortunately, `kubectl` doesn't support custom "plurals" correctly yet, see https://github.com/kubernetes/kubernetes/issues/51639
	ServiceBrokerResourcePlural ResourcePlural = "servicebrokers"

	// ServiceInstanceCredentialKind is the name of a Service Instance
	// Credential resource, a Kubernetes third party resource.
	ServiceInstanceCredentialKind Kind = "ServiceInstanceCredential"

	// ServiceInstanceCredentialListKind is the name for lists of Service
	// Instance Credentials
	ServiceInstanceCredentialListKind Kind = "ServiceInstanceCredentialList"

	// ServiceInstanceCredentialResourcePlural is the plural name of a Service Instance
	// Credential resource, a Kubernetes third party resource.
	ServiceInstanceCredentialResourcePlural ResourcePlural = "serviceinstancecredentials"

	// ServiceClassKind is the name of a Service Class resource, a Kubernetes third party resource.
	ServiceClassKind Kind = "ServiceClass"

	// ServiceClassListKind is the name of a list of service class resources
	ServiceClassListKind Kind = "ServiceClassList"

	// ServiceClassResourcePlural is the plural name of a Service Class resource, a Kubernetes third party resource.
	ServiceClassResourcePlural ResourcePlural = "serviceclasses"

	// ServiceInstanceKind is the name of a Service Instance resource, a Kubernetes third party resource.
	ServiceInstanceKind Kind = "ServiceInstance"

	// ServiceInstanceListKind is the name of a list of service instance resources
	ServiceInstanceListKind Kind = "ServiceInstanceList"

	// ServiceInstanceResourcePlural is the plural name of a Service Instance resource, a Kubernetes third party resource.
	ServiceInstanceResourcePlural ResourcePlural = "serviceinstances"
)
