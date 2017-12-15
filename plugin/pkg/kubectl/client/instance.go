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

package client

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetInstance retrieves a service instance by external name in a given namespace
func (c *PluginClient) GetInstance(instanceName, namespace string) (*v1beta1.ServiceInstance, error) {
	instance, err := c.ScClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(instanceName, v1.GetOptions{})
	return instance, err
}

//ListInstances returns all service instances in a given namespace
func (c *PluginClient) ListInstances(namespace string) (*v1beta1.ServiceInstanceList, error) {
	instances, err := c.ScClient.ServicecatalogV1beta1().ServiceInstances(namespace).List(v1.ListOptions{})
	return instances, err
}
