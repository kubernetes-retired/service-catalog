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

package client

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetBinding retrieves  a binding by external name in a given namespace
func (c *PluginClient) GetBinding(bindingName, namespace string) (*v1beta1.ServiceBinding, error) {
	binding, err := c.ScClient.ServicecatalogV1beta1().ServiceBindings(namespace).Get(bindingName, v1.GetOptions{})
	return binding, err
}

// ListBindings returns a list of all bindings in a namespace
func (c *PluginClient) ListBindings(namespace string) (*v1beta1.ServiceBindingList, error) {
	bindings, err := c.ScClient.ServicecatalogV1beta1().ServiceBindings(namespace).List(v1.ListOptions{})
	return bindings, err
}
