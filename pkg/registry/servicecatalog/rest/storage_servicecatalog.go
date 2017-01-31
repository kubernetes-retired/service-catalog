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
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/genericapiserver"
	"k8s.io/kubernetes/pkg/registry"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	servicecatalogv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/binding"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/broker"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/instance"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/serviceclass"
)

// StorageProvider provides a factory method to create a new APIGroupInfo for
// the servicecatalog API group.
type StorageProvider struct{}

// NewRESTStorage is a factory method to make a new APIGroupInfo for the
// servicecatalog API group.
func (p StorageProvider) NewRESTStorage(apiResourceConfigSource genericapiserver.APIResourceConfigSource, restOptionsGetter registry.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(servicecatalog.GroupName)
	apiGroupInfo.GroupMeta.GroupVersion = servicecatalogv1alpha1.SchemeGroupVersion
	apiGroupInfo.VersionedResourcesStorageMap = map[string]map[string]rest.Storage{
		servicecatalogv1alpha1.SchemeGroupVersion.Version: p.v1alpha1Storage(apiResourceConfigSource, restOptionsGetter),
	}

	return apiGroupInfo, true
}

func (p StorageProvider) v1alpha1Storage(apiResourceConfigSource genericapiserver.APIResourceConfigSource, restOptionsGetter registry.RESTOptionsGetter) map[string]rest.Storage {
	brokers, brokersStatus := broker.NewStorage(restOptionsGetter(servicecatalog.Resource("brokers")))
	return map[string]rest.Storage{
		"brokers":        brokers,
		"brokers/status": brokersStatus,
		"serviceclasses": serviceclass.NewStorage(restOptionsGetter(servicecatalog.Resource("serviceclasses"))),
		"instances":      instance.NewStorage(restOptionsGetter(servicecatalog.Resource("instances"))),
		"bindings":       binding.NewStorage(restOptionsGetter(servicecatalog.Resource("bindings"))),
	}
}

// GroupName returns the API group name.
func (p StorageProvider) GroupName() string {
	return servicecatalog.GroupName
}
