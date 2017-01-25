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

package mem

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient"
)

type apiClient struct {
	namespaces     []string
	brokers        *brokerClient
	serviceClasses *serviceClassClient
	instances      map[string]apiclient.InstanceClient
	bindings       map[string]apiclient.BindingClient
}

// NewAPIClient creates an instance of APIClient interface, backed by memory.
func NewAPIClient() apiclient.APIClient {
	return &apiClient{
		namespaces:     nil,
		brokers:        newBrokerClient(),
		serviceClasses: newServiceClassClient(),
		instances:      make(map[string]apiclient.InstanceClient),
		bindings:       make(map[string]apiclient.BindingClient),
	}
}

// NewPopulatedAPIClient is the equivalent of NewAPIClient, except
// pre-populataes the underlying in-memory storage with brokers and service
// classes
func NewPopulatedAPIClient(
	namespaces []string,
	brokers map[string]*servicecatalog.Broker,
	serviceClasses map[string]*servicecatalog.ServiceClass,
	instances map[string]apiclient.InstanceClient,
	bindings map[string]apiclient.BindingClient,
) apiclient.APIClient {
	brokerClient := newBrokerClient()
	brokerClient.brokers = brokers
	serviceClassClient := newServiceClassClient()
	serviceClassClient.classes = serviceClasses
	return &apiClient{
		namespaces:     namespaces,
		brokers:        brokerClient,
		serviceClasses: serviceClassClient,
		instances:      instances,
		bindings:       bindings,
	}
}

func (c *apiClient) Namespaces() ([]string, error) {
	return c.namespaces, nil
}

func (c *apiClient) Brokers() apiclient.BrokerClient {
	return c.brokers
}

func (c *apiClient) ServiceClasses() apiclient.ServiceClassClient {
	return c.serviceClasses
}

func (c *apiClient) Instances(ns string) apiclient.InstanceClient {
	ret, ok := c.instances[ns]
	if !ok {
		ret = newInstanceClient()
		c.instances[ns] = ret
	}
	return ret
}

func (c *apiClient) Bindings(ns string) apiclient.BindingClient {
	ret, ok := c.bindings[ns]
	if !ok {
		ret = newBindingClient()
		c.bindings[ns] = ret
	}
	return ret
}
