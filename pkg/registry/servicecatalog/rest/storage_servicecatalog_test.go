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

package rest

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/genericapiserver"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/storage"
	"k8s.io/kubernetes/pkg/storage/storagebackend"
	"k8s.io/kubernetes/pkg/storage/storagebackend/factory"
)

func testRESTOptionsGetter(
	retStorageInterface storage.Interface,
	retDestroyFunc func(),
) registry.RESTOptionsGetter {
	return func(resource schema.GroupResource) generic.RESTOptions {
		return generic.RESTOptions{
			StorageConfig: &storagebackend.Config{},
			Decorator: generic.StorageDecorator(func(
				config *storagebackend.Config,
				capacity int,
				objectType runtime.Object,
				resourcePrefix string,
				scopeStrategy rest.NamespaceScopedStrategy,
				newListFunc func() runtime.Object,
				getAttrsFunc func(
					runtime.Object,
				) (labels.Set, fields.Set, error),
				trigger storage.TriggerPublisherFunc,
			) (storage.Interface, factory.DestroyFunc) {
				return retStorageInterface, retDestroyFunc
			}),
		}
	}
}

func TestV1Alpha1Storage(t *testing.T) {
	provider := StorageProvider{
		DefaultNamespace: "test-default",
		StorageType:      server.StorageTypeTPR,
		RESTClient:       nil,
	}
	configSource := genericapiserver.NewResourceConfig()
	roGetter := testRESTOptionsGetter(nil, func() {})
	storageMap, err := provider.v1alpha1Storage(configSource, roGetter)
	if err != nil {
		t.Fatalf("error getting v1alpha1 storage (%s)", err)
	}
	_, brokerStorageExists := storageMap["brokers"]
	if !brokerStorageExists {
		t.Fatalf("no broker storage found")
	}
	// TODO: do stuff with broker storage
	_, brokerStatusStorageExists := storageMap["brokers/status"]
	if !brokerStatusStorageExists {
		t.Fatalf("no broker status storage found")
	}
	// TODO: do stuff with broker status storage

	_, serviceClassStorageExists := storageMap["serviceclasses"]
	if !serviceClassStorageExists {
		t.Fatalf("no service class storage found")
	}
	// TODO: do stuff with service class storage

	_, instanceStorageExists := storageMap["instances"]
	if !instanceStorageExists {
		t.Fatalf("no instance storage found")
	}
	// TODO: do stuff with instance storage

	_, bindingStorageExists := storageMap["bindings"]
	if !bindingStorageExists {
		t.Fatalf("no binding storage found")
	}
	// TODO: do stuff with binding storage

}
