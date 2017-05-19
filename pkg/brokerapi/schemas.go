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
	ServiceInstances *Schema `json:"service_instances,omitempty"`
	ServiceBindings  *Schema `json:"service_bindings,omitempty"`
}

// Schema represents a plan's schemas for a create and update of an API
// resource.
type Schema struct {
	Create interface{} `json:"create,omitempty"`
	Update interface{} `json:"update,omitempty"`
}
