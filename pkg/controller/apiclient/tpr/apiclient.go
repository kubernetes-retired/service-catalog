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

package tpr

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/watch"
	// Need this for gcp auth
	_ "k8s.io/client-go/1.5/kubernetes"
)

type apiClient struct {
	watcher *watch.Watcher
}

// NewAPIClient creates an instance of APIClient backed by Kubernetes
// third-party resources.
func NewAPIClient(w *watch.Watcher) apiclient.APIClient {
	return &apiClient{
		watcher: w,
	}
}

func (c *apiClient) Brokers() apiclient.BrokerClient {
	return newBrokerClient(c.watcher)
}

func (c *apiClient) ServiceClasses() apiclient.ServiceClassClient {
	return newServiceClassClient(c.watcher)
}

func (c *apiClient) Instances(ns string) apiclient.InstanceClient {
	return newInstanceClient(c.watcher, ns)
}

func (c *apiClient) Bindings(ns string) apiclient.BindingClient {
	return newBindingClient(c.watcher, ns)
}
