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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
	utilfeature "k8s.io/apiserver/pkg/util/feature"

	scfeatures "github.com/kubernetes-incubator/service-catalog/pkg/features"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/binding"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/clusterservicebroker"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/clusterserviceclass"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/clusterserviceplan"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/instance"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/servicebroker"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/serviceclass"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/serviceplan"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Testing with Ginkgo", func() {
	It("checks v1beta1 standard storage", func() {

		defer utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.NamespacedServiceBroker))
		err := utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.NamespacedServiceBroker))
		if err != nil {
			GinkgoT().Fatal(err)
		}
		provider := StorageProvider{
			DefaultNamespace: "test-default",
			StorageType:      server.StorageTypeEtcd,
			RESTClient:       nil,
		}
		configSource := serverstorage.NewResourceConfig()
		roGetter := testRESTOptionsGetter(nil, func() {})
		storageMap, err := provider.v1beta1Storage(configSource, roGetter)
		if err != nil {
			GinkgoT().Fatalf("error getting v1beta1 storage (%s)", err)
		}

		storages := [...]string{
			"clusterservicebrokers",
			"clusterserviceclasses",
			"clusterserviceplans",
			"serviceinstances",
			"servicebindings",
			"serviceclasses",
			"serviceplans",
			"servicebrokers",
		}

		for _, storage := range storages {
			s, storageExists := storageMap[storage]
			if !storageExists {
				GinkgoT().Fatalf("no %q storage found", storage)
			}
			checkStorageType(GinkgoT(), s)
		}
	})

	// TestCheckStatusRESTTypes ensures that our Status storage types fulfill the
	// specific interfaces that are expected and no more. This is similar to what is
	// done internally to the apiserver when it is deciding what http verbs to
	// expose on each resource. For status, we only want to support GET and a form
	// of update like PATCH. This could partly be done by type var type-assertions
	// at the site of declaration, but because we want to explicitly determine that
	// an object does NOT implement some interface, it has to be done at runtime.
	It("checks v1beta1 StatusREST storage", func() {
		checkStatusStorageType(GinkgoT(), &clusterservicebroker.StatusREST{})
		checkStatusStorageType(GinkgoT(), &servicebroker.StatusREST{})
		checkStatusStorageType(GinkgoT(), &clusterserviceclass.StatusREST{})
		checkStatusStorageType(GinkgoT(), &serviceclass.StatusREST{})
		checkStatusStorageType(GinkgoT(), &clusterserviceplan.StatusREST{})
		checkStatusStorageType(GinkgoT(), &serviceplan.StatusREST{})
		checkStatusStorageType(GinkgoT(), &instance.StatusREST{})
		checkStatusStorageType(GinkgoT(), &binding.StatusREST{})
	})
})

type GetRESTOptionsHelper struct {
	retStorageInterface storage.Interface
	retDestroyFunc      func()
}

func (g GetRESTOptionsHelper) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
	return generic.RESTOptions{
		ResourcePrefix: resource.Group + "/" + resource.Resource,
		StorageConfig:  &storagebackend.Config{},
		Decorator: generic.StorageDecorator(func(
			config *storagebackend.Config,
			objectType runtime.Object,
			resourcePrefix string,
			keyFunc func(obj runtime.Object) (string, error),
			newListFunc func() runtime.Object,
			getAttrsFunc storage.AttrFunc,
			trigger storage.TriggerPublisherFunc,
		) (storage.Interface, factory.DestroyFunc) {
			return g.retStorageInterface, g.retDestroyFunc
		})}, nil
}

func testRESTOptionsGetter(
	retStorageInterface storage.Interface,
	retDestroyFunc func(),
) generic.RESTOptionsGetter {
	return GetRESTOptionsHelper{retStorageInterface, retDestroyFunc}
}

func checkStorageType(t GinkgoTInterface, s rest.Storage) {
	// Our normal stores are all of these things
	if _, isStorageType := s.(rest.Storage); !isStorageType {
		t.Errorf("%q not compliant to storage interface", s)
	}
	if _, isStorageType := s.(rest.Updater); !isStorageType {
		t.Errorf("%q not compliant to updater interface", s)
	}
	if _, isStorageType := s.(rest.Getter); !isStorageType {
		t.Errorf("%q not compliant to getter interface", s)
	}
	if _, isStorageType := s.(rest.Lister); !isStorageType {
		t.Errorf("%q not compliant to lister interface", s)
	}
	if _, isStorageType := s.(rest.Creater); !isStorageType {
		t.Errorf("%q not compliant to creater interface", s)
	}
	if _, isStorageType := s.(rest.GracefulDeleter); !isStorageType {
		t.Errorf("%q not compliant to GracefulDeleter interface", s)
	}
	if _, isStorageType := s.(rest.CollectionDeleter); !isStorageType {
		t.Errorf("%q not compliant to CollectionDeleter interface", s)
	}
	if _, isStorageType := s.(rest.Watcher); !isStorageType {
		t.Errorf("%q not compliant to watcher interface", s)
	}
	if _, isStorageType := s.(rest.StandardStorage); !isStorageType {
		t.Errorf("%q not compliant to StandardStorage interface", s)
	}
}

func checkStatusStorageType(t GinkgoTInterface, s rest.Storage) {
	// Status is New & Get & Update ONLY
	if _, isStandardStorage := s.(rest.Storage); !isStandardStorage {
		t.Errorf("not compliant to storage interface for %q", s)
	}
	if _, isStandardStorage := s.(rest.Updater); !isStandardStorage {
		t.Errorf("not compliant to updaterer interface for %q", s)
	}
	if _, isStandardStorage := s.(rest.Getter); !isStandardStorage {
		t.Errorf("not compliant to getter interface for %q", s)
	}
	// NONE of these things
	if _, isStandardStorage := s.(rest.Lister); isStandardStorage {
		t.Errorf("%q was a lister but should not be", s)
	}
	if _, isStandardStorage := s.(rest.Creater); isStandardStorage {
		t.Errorf("%q was a creater but should not be", s)
	}
	if _, isStandardStorage := s.(rest.GracefulDeleter); isStandardStorage {
		t.Errorf("%q was a graceful delete but should not be", s)
	}
	if _, isStandardStorage := s.(rest.CollectionDeleter); isStandardStorage {
		t.Errorf("%q was a collection deleter but should not be", s)
	}
	if _, isStandardStorage := s.(rest.Watcher); isStandardStorage {
		t.Errorf("%q was a watcher but should not be", s)
	}
	if _, isStandardStorage := s.(rest.StandardStorage); isStandardStorage {
		t.Errorf("%q was a StandardStorage but should not be", s)
	}
}
