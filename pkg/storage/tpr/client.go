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
	"k8s.io/client-go/dynamic"
)

type errUnsupportedResource struct {
	kind Kind
}

func (e errUnsupportedResource) Error() string {
	return fmt.Sprintf("unsupported resource %s", e.kind)
}

// GetResourceClient returns the *dynamic.ResourceClient for a given resource type
func GetResourceClient(cl *dynamic.Client, kind Kind, namespace string) (*dynamic.ResourceClient, error) {
	switch kind {
	case ServiceCatalogInstanceKind, ServiceCatalogInstanceListKind:
		return cl.Resource(&ServiceCatalogInstanceResource, namespace), nil
	case ServiceCatalogBindingKind, ServiceCatalogBindingListKind:
		return cl.Resource(&ServiceCatalogBindingResource, namespace), nil
	case ServiceCatalogBrokerKind, ServiceCatalogBrokerListKind:
		return cl.Resource(&ServiceCatalogBrokerResource, namespace), nil
	case ServiceCatalogServiceClassKind, ServiceCatalogServiceClassListKind:
		return cl.Resource(&ServiceCatalogServiceClassResource, namespace), nil
	default:
		return nil, errUnsupportedResource{kind: kind}
	}
}
