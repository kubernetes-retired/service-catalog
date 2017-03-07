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

package rest

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	servicecatalogv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/binding"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/broker"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/instance"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/serviceclass"
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/etcd"
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/tpr"
	"k8s.io/kubernetes/pkg/api/rest"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/genericapiserver"
	"k8s.io/kubernetes/pkg/registry"
	"k8s.io/kubernetes/pkg/storage"
)

// StorageProvider provides a factory method to create a new APIGroupInfo for
// the servicecatalog API group. It implements (./pkg/apiserver).RESTStorageProvider
type StorageProvider struct {
	DefaultNamespace string
	StorageType      server.StorageType
	Client           clientset.Interface
}

// NewRESTStorage is a factory method to make a new APIGroupInfo for the
// servicecatalog API group.
func (p StorageProvider) NewRESTStorage(
	apiResourceConfigSource genericapiserver.APIResourceConfigSource,
	restOptionsGetter registry.RESTOptionsGetter,
) (*genericapiserver.APIGroupInfo, error) {

	storage, err := p.v1alpha1Storage(apiResourceConfigSource, restOptionsGetter)
	if err != nil {
		return nil, err
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(servicecatalog.GroupName)
	apiGroupInfo.GroupMeta.GroupVersion = servicecatalogv1alpha1.SchemeGroupVersion

	apiGroupInfo.VersionedResourcesStorageMap = map[string]map[string]rest.Storage{
		servicecatalogv1alpha1.SchemeGroupVersion.Version: storage,
	}

	return &apiGroupInfo, nil
}

func (p StorageProvider) v1alpha1Storage(
	apiResourceConfigSource genericapiserver.APIResourceConfigSource,
	restOptionsGetter registry.RESTOptionsGetter,
) (map[string]rest.Storage, error) {
	brokerOpts := server.NewOptions(
		etcd.Options{
			RESTOptions:   restOptionsGetter(servicecatalog.Resource("brokers")),
			Capacity:      1000,
			ObjectType:    broker.EmptyObject(),
			ScopeStrategy: broker.NewScopeStrategy(),
			NewListFunc:   broker.NewList,
			GetAttrsFunc:  broker.GetAttrs,
			Trigger:       storage.NoTriggerPublisher,
		},
		tpr.Options{
			HasNamespace:     false,
			RESTOptions:      restOptionsGetter(servicecatalog.Resource("brokers")),
			DefaultNamespace: p.DefaultNamespace,
			Client:           p.Client,
			SingularKind:     tpr.ServiceBrokerKind,
			NewSingularFunc:  broker.NewSingular,
			ListKind:         tpr.ServiceBrokerListKind,
			NewListFunc:      broker.NewList,
			CheckObjectFunc:  broker.CheckObject,
			DestroyFunc:      func() {},
			Keyer: tpr.Keyer{
				DefaultNamespace: p.DefaultNamespace,
				ResourceName:     tpr.ServiceBrokerKind.String(),
				Separator:        "/",
			},
		},
		p.StorageType,
	)

	serviceClassOpts := server.NewOptions(
		etcd.Options{
			RESTOptions:   restOptionsGetter(servicecatalog.Resource("serviceclasses")),
			Capacity:      1000,
			ObjectType:    serviceclass.EmptyObject(),
			ScopeStrategy: serviceclass.NewScopeStrategy(),
			NewListFunc:   serviceclass.NewList,
			GetAttrsFunc:  serviceclass.GetAttrs,
			Trigger:       storage.NoTriggerPublisher,
		},
		tpr.Options{
			HasNamespace:     false,
			RESTOptions:      restOptionsGetter(servicecatalog.Resource("serviceclasses")),
			DefaultNamespace: p.DefaultNamespace,
			Client:           p.Client,
			SingularKind:     tpr.ServiceClassKind,
			NewSingularFunc:  serviceclass.NewSingular,
			ListKind:         tpr.ServiceClassListKind,
			NewListFunc:      serviceclass.NewList,
			CheckObjectFunc:  serviceclass.CheckObject,
			DestroyFunc:      func() {},
			Keyer: tpr.Keyer{
				DefaultNamespace: p.DefaultNamespace,
				ResourceName:     tpr.ServiceClassKind.String(),
				Separator:        "/",
			},
		},
		p.StorageType,
	)

	instanceOpts := server.NewOptions(
		etcd.Options{
			RESTOptions:   restOptionsGetter(servicecatalog.Resource("instances")),
			Capacity:      1000,
			ObjectType:    instance.EmptyObject(),
			ScopeStrategy: instance.NewScopeStrategy(),
			NewListFunc:   instance.NewList,
			GetAttrsFunc:  instance.GetAttrs,
			Trigger:       storage.NoTriggerPublisher,
		},
		tpr.Options{
			HasNamespace:     true,
			RESTOptions:      restOptionsGetter(servicecatalog.Resource("instances")),
			DefaultNamespace: p.DefaultNamespace,
			Client:           p.Client,
			SingularKind:     tpr.ServiceInstanceKind,
			NewSingularFunc:  instance.NewSingular,
			ListKind:         tpr.ServiceInstanceListKind,
			NewListFunc:      instance.NewList,
			CheckObjectFunc:  instance.CheckObject,
			DestroyFunc:      func() {},
			Keyer: tpr.Keyer{
				DefaultNamespace: p.DefaultNamespace,
				ResourceName:     tpr.ServiceInstanceKind.String(),
				Separator:        "/",
			},
		},
		p.StorageType,
	)

	bindingsOpts := server.NewOptions(
		etcd.Options{
			RESTOptions:   restOptionsGetter(servicecatalog.Resource("bindings")),
			Capacity:      1000,
			ObjectType:    binding.EmptyObject(),
			ScopeStrategy: binding.NewScopeStrategy(),
			NewListFunc:   binding.NewList,
			GetAttrsFunc:  binding.GetAttrs,
			Trigger:       storage.NoTriggerPublisher,
		},
		tpr.Options{
			HasNamespace:     true,
			RESTOptions:      restOptionsGetter(servicecatalog.Resource("bindings")),
			DefaultNamespace: p.DefaultNamespace,
			Client:           p.Client,
			SingularKind:     tpr.ServiceBindingKind,
			NewSingularFunc:  binding.NewSingular,
			ListKind:         tpr.ServiceBindingListKind,
			NewListFunc:      binding.NewList,
			CheckObjectFunc:  binding.CheckObject,
			DestroyFunc:      func() {},
			Keyer: tpr.Keyer{
				DefaultNamespace: p.DefaultNamespace,
				ResourceName:     tpr.ServiceBindingKind.String(),
				Separator:        "/",
			},
		},
		p.StorageType,
	)

	brokerStorage, brokerStatusStorage := broker.NewStorage(*brokerOpts)
	serviceClassStorage := serviceclass.NewStorage(*serviceClassOpts)
	instanceStorage, instanceStatusStorage := instance.NewStorage(*instanceOpts)
	bindingStorage, bindingStatusStorage, err := binding.NewStorage(*bindingsOpts)
	if err != nil {
		return nil, err
	}
	return map[string]rest.Storage{
		"brokers":          brokerStorage,
		"brokers/status":   brokerStatusStorage,
		"serviceclasses":   serviceClassStorage,
		"instances":        instanceStorage,
		"instances/status": instanceStatusStorage,
		"bindings":         bindingStorage,
		"bindings/status":  bindingStatusStorage,
	}, nil
}

// GroupName returns the API group name.
func (p StorageProvider) GroupName() string {
	return servicecatalog.GroupName
}
