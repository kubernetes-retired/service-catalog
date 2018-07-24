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
	"github.com/kubernetes-incubator/service-catalog/pkg/api"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	servicecatalogv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/binding"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/clusterservicebroker"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/clusterserviceclass"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/clusterserviceplan"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/instance"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/servicebroker"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/serviceclass"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/serviceplan"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"

	scfeatures "github.com/kubernetes-incubator/service-catalog/pkg/features"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
)

// StorageProvider provides a factory method to create a new APIGroupInfo for
// the servicecatalog API group. It implements (./pkg/apiserver).RESTStorageProvider
type StorageProvider struct {
}

// NewRESTStorage is a factory method to make a new APIGroupInfo for the
// servicecatalog API group.
func (p StorageProvider) NewRESTStorage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (*genericapiserver.APIGroupInfo, error) {

	storage, err := p.v1beta1Storage(apiResourceConfigSource, restOptionsGetter)
	if err != nil {
		return nil, err
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(servicecatalog.GroupName, api.Scheme, api.ParameterCodec, api.Codecs)

	apiGroupInfo.VersionedResourcesStorageMap = map[string]map[string]rest.Storage{
		servicecatalogv1beta1.SchemeGroupVersion.Version: storage,
	}

	return &apiGroupInfo, nil
}

func (p StorageProvider) v1beta1Storage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (map[string]rest.Storage, error) {
	clusterServiceBrokerRESTOptions, err := restOptionsGetter.GetRESTOptions(servicecatalog.Resource("clusterservicebrokers"))
	if err != nil {
		return nil, err
	}
	clusterServiceBrokerStorage, clusterServiceBrokerStatusStorage, err := clusterservicebroker.NewStorage(clusterServiceBrokerRESTOptions)
	if err != nil {
		return nil, err
	}

	clusterServiceClassRESTOptions, err := restOptionsGetter.GetRESTOptions(servicecatalog.Resource("clusterserviceclasses"))
	if err != nil {
		return nil, err
	}
	clusterServiceClassStorage, clusterServiceClassStatusStorage, err := clusterserviceclass.NewStorage(clusterServiceClassRESTOptions)
	if err != nil {
		return nil, err
	}

	clusterServicePlanRESTOptions, err := restOptionsGetter.GetRESTOptions(servicecatalog.Resource("clusterserviceplans"))
	if err != nil {
		return nil, err
	}
	clusterServicePlanStorage, clusterServicePlanStatusStorage, err := clusterserviceplan.NewStorage(clusterServicePlanRESTOptions)
	if err != nil {
		return nil, err
	}

	instanceClassRESTOptions, err := restOptionsGetter.GetRESTOptions(servicecatalog.Resource("serviceinstances"))
	if err != nil {
		return nil, err
	}
	instanceStorage, instanceStatusStorage, instanceReferencesStorage, err := instance.NewStorage(instanceClassRESTOptions)
	if err != nil {
		return nil, err
	}

	bindingClassRESTOptions, err := restOptionsGetter.GetRESTOptions(servicecatalog.Resource("servicebindings"))
	if err != nil {
		return nil, err
	}
	bindingStorage, bindingStatusStorage, err := binding.NewStorage(bindingClassRESTOptions)
	if err != nil {
		return nil, err
	}

	storageMap := map[string]rest.Storage{
		"clusterservicebrokers":        clusterServiceBrokerStorage,
		"clusterservicebrokers/status": clusterServiceBrokerStatusStorage,
		"clusterserviceclasses":        clusterServiceClassStorage,
		"clusterserviceclasses/status": clusterServiceClassStatusStorage,
		"clusterserviceplans":          clusterServicePlanStorage,
		"clusterserviceplans/status":   clusterServicePlanStatusStorage,
		"serviceinstances":             instanceStorage,
		"serviceinstances/status":      instanceStatusStorage,
		"serviceinstances/reference":   instanceReferencesStorage,
		"servicebindings":              bindingStorage,
		"servicebindings/status":       bindingStatusStorage,
	}

	if utilfeature.DefaultFeatureGate.Enabled(scfeatures.NamespacedServiceBroker) {
		serviceClassRESTOptions, err := restOptionsGetter.GetRESTOptions(servicecatalog.Resource("serviceclasses"))
		if err != nil {
			return nil, err
		}
		serviceClassStorage, serviceClassStatusStorage, err := serviceclass.NewStorage(serviceClassRESTOptions)
		if err != nil {
			return nil, err
		}

		serviceBrokerRESTOptions, err := restOptionsGetter.GetRESTOptions(servicecatalog.Resource("servicebrokers"))
		if err != nil {
			return nil, err
		}
		serviceBrokerStorage, serviceBrokerStatusStorage, err := servicebroker.NewStorage(serviceBrokerRESTOptions)
		if err != nil {
			return nil, err
		}

		servicePlanRESTOptions, err := restOptionsGetter.GetRESTOptions(servicecatalog.Resource("serviceplans"))
		if err != nil {
			return nil, err
		}
		servicePlanStorage, servicePlanStatusStorage, err := serviceplan.NewStorage(servicePlanRESTOptions)
		if err != nil {
			return nil, err
		}

		storageMap["serviceclasses"] = serviceClassStorage
		storageMap["serviceclasses/status"] = serviceClassStatusStorage
		storageMap["serviceplans"] = servicePlanStorage
		storageMap["serviceplans/status"] = servicePlanStatusStorage
		storageMap["servicebrokers"] = serviceBrokerStorage
		storageMap["servicebrokers/status"] = serviceBrokerStatusStorage
	}

	return storageMap, nil
}

// GroupName returns the API group name.
func (p StorageProvider) GroupName() string {
	return servicecatalog.GroupName
}
