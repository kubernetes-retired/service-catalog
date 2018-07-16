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
	api "github.com/kubernetes-incubator/service-catalog/pkg/api"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/settings"
	settingsapiv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/settings/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/settings/podpreset"

	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
)

// StorageProvider provides a factory method to create a new APIGroupInfo for
// the servicecatalog API group. It implements (./pkg/apiserver).RESTStorageProvider
type StorageProvider struct {
}

// NewRESTStorage is a factory method to make a new APIGroupInfo for the
// settings API group.
func (p StorageProvider) NewRESTStorage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (*genericapiserver.APIGroupInfo, error) {

	storage, err := p.v1alpha1Storage(apiResourceConfigSource, restOptionsGetter)
	if err != nil {
		return nil, err
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(settings.GroupName, api.Scheme, api.ParameterCodec, api.Codecs)

	if apiResourceConfigSource.AnyVersionForGroupEnabled(settingsapiv1alpha1.SchemeGroupVersion.Group) {
		apiGroupInfo.VersionedResourcesStorageMap = map[string]map[string]rest.Storage{
			settingsapiv1alpha1.SchemeGroupVersion.Version: storage,
		}
	}

	return &apiGroupInfo, nil
}

func (p StorageProvider) v1alpha1Storage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter) (map[string]rest.Storage, error) {

	version := settingsapiv1alpha1.SchemeGroupVersion

	storage := map[string]rest.Storage{}
	if apiResourceConfigSource.VersionEnabled(version) {
		podPresetRESTOptions, err := restOptionsGetter.GetRESTOptions(settings.Resource("podpresets"))
		if err != nil {
			return nil, err
		}
		podPresetStorage, err := podpreset.NewStorage(podPresetRESTOptions)
		if err != nil {
			return nil, err
		}
		storage["podpresets"] = podPresetStorage
	}
	return storage, nil
}

// GroupName returns the API group name.
func (p StorageProvider) GroupName() string {
	return settings.GroupName
}
