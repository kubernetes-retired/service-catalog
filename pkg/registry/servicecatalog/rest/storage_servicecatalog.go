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
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	servicecatalogv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/binding"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/broker"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/instance"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/serviceclass"
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/crd"
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/etcd"
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/tpr"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/client-go/pkg/api"
	restclient "k8s.io/client-go/rest"
)

// StorageProvider provides a factory method to create a new APIGroupInfo for
// the servicecatalog API group. It implements (./pkg/apiserver).RESTStorageProvider
type StorageProvider struct {
	DefaultNamespace string
	StorageType      server.StorageType
	RESTClient       restclient.Interface
}

// NewRESTStorage is a factory method to make a new APIGroupInfo for the
// servicecatalog API group.
func (p StorageProvider) NewRESTStorage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (*genericapiserver.APIGroupInfo, error) {

	storage, err := p.v1alpha1Storage(apiResourceConfigSource, restOptionsGetter)
	if err != nil {
		return nil, err
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(servicecatalog.GroupName, api.Registry, server.Scheme, server.ParameterCodec, server.Codecs)
	apiGroupInfo.GroupMeta.GroupVersion = servicecatalogv1alpha1.SchemeGroupVersion

	apiGroupInfo.VersionedResourcesStorageMap = map[string]map[string]rest.Storage{
		servicecatalogv1alpha1.SchemeGroupVersion.Version: storage,
	}

	return &apiGroupInfo, nil
}

func (p StorageProvider) v1alpha1Storage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (map[string]rest.Storage, error) {
	brokerOpts, err := p.getServerOptions(restOptionsGetter, &resourceMetaDelegate{
		hasNamespaceVal:      false,
		etcdResourceVal:      "servicebrokers",
		tprResourceVal:       "servicebrokers",
		tprKindVal:           tpr.ServiceBrokerKind,
		tprListKindVal:       tpr.ServiceBrokerListKind,
		crdResourceVal:       crd.ServiceBrokerResourcePlural,
		crdKindVal:           crd.ServiceBrokerKind,
		crdListKindVal:       crd.ServiceBrokerListKind,
		emptyObjectFunc:      broker.EmptyObject,
		newScopeStrategyFunc: broker.NewScopeStrategy,
		newListFunc:          broker.NewList,
		newSingularFunc:      broker.NewSingular,
		getAttrsFunc:         broker.GetAttrs,
		checkObjectFunc:      broker.CheckObject,
	})
	if err != nil {
		return nil, err
	}

	serviceClassOpts, err := p.getServerOptions(restOptionsGetter, &resourceMetaDelegate{
		hasNamespaceVal:      false,
		etcdResourceVal:      "serviceclasses",
		tprResourceVal:       "serviceclasses",
		tprKindVal:           tpr.ServiceClassKind,
		tprListKindVal:       tpr.ServiceClassListKind,
		crdResourceVal:       crd.ServiceClassResourcePlural,
		crdKindVal:           crd.ServiceBrokerKind,
		crdListKindVal:       crd.ServiceBrokerListKind,
		emptyObjectFunc:      serviceclass.EmptyObject,
		newScopeStrategyFunc: serviceclass.NewScopeStrategy,
		newListFunc:          serviceclass.NewList,
		newSingularFunc:      serviceclass.NewSingular,
		getAttrsFunc:         serviceclass.GetAttrs,
		checkObjectFunc:      serviceclass.CheckObject,
		hardDeleteVal:        true,
	})
	if err != nil {
		return nil, err
	}

	instanceOpts, err := p.getServerOptions(restOptionsGetter, &resourceMetaDelegate{
		hasNamespaceVal:      true,
		etcdResourceVal:      "serviceinstances",
		tprResourceVal:       "serviceinstances",
		tprKindVal:           tpr.ServiceInstanceKind,
		tprListKindVal:       tpr.ServiceInstanceListKind,
		crdResourceVal:       crd.ServiceInstanceResourcePlural,
		crdKindVal:           crd.ServiceInstanceKind,
		crdListKindVal:       crd.ServiceInstanceListKind,
		emptyObjectFunc:      instance.EmptyObject,
		newScopeStrategyFunc: instance.NewScopeStrategy,
		newListFunc:          instance.NewList,
		newSingularFunc:      instance.NewSingular,
		getAttrsFunc:         instance.GetAttrs,
		checkObjectFunc:      instance.CheckObject,
	})
	if err != nil {
		return nil, err
	}

	bindingsOpts, err := p.getServerOptions(restOptionsGetter, &resourceMetaDelegate{
		hasNamespaceVal:      true,
		etcdResourceVal:      "serviceinstancecredentials",
		tprResourceVal:       "serviceinstancecredentials",
		tprKindVal:           tpr.ServiceInstanceCredentialKind,
		tprListKindVal:       tpr.ServiceInstanceCredentialListKind,
		crdResourceVal:       crd.ServiceInstanceCredentialResourcePlural,
		crdKindVal:           crd.ServiceInstanceCredentialKind,
		crdListKindVal:       crd.ServiceInstanceCredentialListKind,
		emptyObjectFunc:      binding.EmptyObject,
		newScopeStrategyFunc: binding.NewScopeStrategy,
		newListFunc:          binding.NewList,
		newSingularFunc:      binding.NewSingular,
		getAttrsFunc:         binding.GetAttrs,
		checkObjectFunc:      binding.CheckObject,
	})
	if err != nil {
		return nil, err
	}

	brokerStorage, brokerStatusStorage := broker.NewStorage(*brokerOpts)
	serviceClassStorage := serviceclass.NewStorage(*serviceClassOpts)
	instanceStorage, instanceStatusStorage := instance.NewStorage(*instanceOpts)
	bindingStorage, bindingStatusStorage, err := binding.NewStorage(*bindingsOpts)
	if err != nil {
		return nil, err
	}
	return map[string]rest.Storage{
		"servicebrokers":                    brokerStorage,
		"servicebrokers/status":             brokerStatusStorage,
		"serviceclasses":                    serviceClassStorage,
		"serviceinstances":                  instanceStorage,
		"serviceinstances/status":           instanceStatusStorage,
		"serviceinstancecredentials":        bindingStorage,
		"serviceinstancecredentials/status": bindingStatusStorage,
	}, nil
}

func (p StorageProvider) getServerOptions(restOptionsGetter generic.RESTOptionsGetter, m ResourceMeta) (*server.Options, error) {
	etcdOpts := etcd.Options{}
	crdOpts := crd.Options{}
	tprOpts := tpr.Options{}

	restOptions, err := p.getRESTOptions(restOptionsGetter, m)
	if err != nil {
		return nil, err
	}
	switch p.StorageType {
	case server.StorageTypeEtcd:
		etcdOpts = etcd.Options{
			RESTOptions:   restOptions,
			Capacity:      1000,
			ObjectType:    m.EmptyObject(),
			ScopeStrategy: m.NewScopeStrategy(),
			NewListFunc:   m.NewList,
			GetAttrsFunc:  m.GetAttrs,
			Trigger:       storage.NoTriggerPublisher,
		}
	case server.StorageTypeCRD:
		crdOpts = crd.Options{
			HasNamespace:     m.HasNamespace(),
			RESTOptions:      restOptions,
			Copier:           api.Scheme,
			DefaultNamespace: p.DefaultNamespace,
			RESTClient:       p.RESTClient,
			ResourcePlural:   m.CrdResource(),
			NewSingularFunc:  m.NewSingular,
			NewListFunc:      m.NewList,
			CheckObjectFunc:  m.CheckObject,
			DestroyFunc:      func() {},
			Keyer: crd.Keyer{
				DefaultNamespace: p.DefaultNamespace,
				ResourceName:     m.CrdKind().String(),
				Separator:        "/",
			},
			HardDelete: m.HardDelete(),
		}
	case server.StorageTypeTPR:
		tprOpts = tpr.Options{
			HasNamespace:     m.HasNamespace(),
			RESTOptions:      restOptions,
			DefaultNamespace: p.DefaultNamespace,
			RESTClient:       p.RESTClient,
			SingularKind:     m.TprKind(),
			NewSingularFunc:  m.NewSingular,
			ListKind:         m.TprListKind(),
			NewListFunc:      m.NewList,
			CheckObjectFunc:  m.CheckObject,
			DestroyFunc:      func() {},
			Keyer: tpr.Keyer{
				DefaultNamespace: p.DefaultNamespace,
				ResourceName:     m.TprKind().String(),
				Separator:        "/",
			},
			HardDelete: m.HardDelete(),
		}
	}

	return server.NewOptions(etcdOpts, crdOpts, tprOpts, p.StorageType), nil
}

func (p StorageProvider) getRESTOptions(restOptionsGetter generic.RESTOptionsGetter, m ResourceMeta) (generic.RESTOptions, error) {
	switch p.StorageType {
	case server.StorageTypeEtcd:
		return restOptionsGetter.GetRESTOptions(servicecatalog.Resource(m.EtcdResource()))
	case server.StorageTypeCRD:
		return restOptionsGetter.GetRESTOptions(crd.InternalResource(m.CrdResource().String()))
	case server.StorageTypeTPR:
		return restOptionsGetter.GetRESTOptions(servicecatalog.Resource(m.TprResource()))
	default:
		return generic.RESTOptions{}, fmt.Errorf("Unexpected storage type: %s", p.StorageType)
	}
}

// GroupName returns the API group name.
func (p StorageProvider) GroupName() string {
	return servicecatalog.GroupName
}
