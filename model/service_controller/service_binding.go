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

package model

import (
	"k8s.io/client-go/1.5/pkg/runtime"
)

// ServiceBinding defines a binding/link between an app and a service instance.
type ServiceBinding struct {
	Name       string                 `json:"name"`
	ID         string                 `json:"id"`
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	Credentials Credential `json:"credentials"`

	// For k8s object completeness
	runtime.TypeMeta `json:",inline"`
}

// CreateServiceBindingRequest defines the payload of the HTTP request
// to create a new binding.
type CreateServiceBindingRequest struct {
	Name       string                 `json:"name"`
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// CreateServiceBindingResponse defines the payload of the HTTP response
// to create a new binding.
type CreateServiceBindingResponse struct {
	Name        string     `json:"name"`
	Credentials Credential `json:"credentials"`
}

// Credential defines some common fields that might make up the credentials
// of a binding response.
type Credential struct {
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}
