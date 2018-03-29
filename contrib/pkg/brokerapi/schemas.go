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

package brokerapi

// Schemas represents a plan's schemas for service instance and binding create
// and update.
type Schemas struct {
	ServiceInstance *ServiceInstanceSchema `json:"service_instance,omitempty"`
	ServiceBinding  *ServiceBindingSchema  `json:"service_binding,omitempty"`
}

// ServiceInstanceSchema represents a plan's schemas for a create and update
// of a service instance.
type ServiceInstanceSchema struct {
	Create *InputParametersSchema `json:"create,omitempty"`
	Update *InputParametersSchema `json:"update,omitempty"`
}

// ServiceBindingSchema represents a plan's schemas for the parameters
// accepted for binding creation.
type ServiceBindingSchema struct {
	Create *RequestResponseSchema `json:"create,omitempty"`
}

// InputParametersSchema represents a schema for input parameters for
// creation or update of an API resource.
type InputParametersSchema struct {
	Parameters interface{} `json:"parameters,omitempty"`
}

// RequestResponseSchema represents a schema for both input parameters and
// the broker's response to the binding request
type RequestResponseSchema struct {
	InputParametersSchema
	Response interface{} `json:"response,omitempty"`
}
