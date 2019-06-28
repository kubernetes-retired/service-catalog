/*
Copyright 2019 The Kubernetes Authors.

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

package clientutil

import (
	"fmt"
	sc "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ClientStorage stores all required clients in upgrade tests
type ClientStorage struct {
	client   kubernetes.Interface
	scClient sc.Interface
}

// NewClientStorage returns pointer to new ClientStorage struct
func NewClientStorage(k8sKubeconfig *rest.Config) (*ClientStorage, error) {
	clientk8s, err := kubernetes.NewForConfig(k8sKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes client: %v", err)
	}
	serviceCatalogClient, err := sc.NewForConfig(k8sKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get ServiceCatalog client: %v", err)
	}

	return &ClientStorage{
		client:   clientk8s,
		scClient: serviceCatalogClient,
	}, nil
}

// KubernetesClient returns kubernetes clientset
func (cs *ClientStorage) KubernetesClient() kubernetes.Interface {
	return cs.client
}

// ServiceCatalogClient returns ServiceCatalog clientset
func (cs *ClientStorage) ServiceCatalogClient() sc.Interface {
	return cs.scClient
}
