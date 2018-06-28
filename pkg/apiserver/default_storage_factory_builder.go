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

package apiserver

import (
	"fmt"
	"strconv"

	"github.com/kubernetes-incubator/service-catalog/pkg/api"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/server/resourceconfig"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	utilflag "k8s.io/apiserver/pkg/util/flag"
)

// NewStorageFactory builds the DefaultStorageFactory.
// Merges defaultResourceConfig with the user specified overrides and merges
// defaultAPIResourceConfig with the corresponding user specified overrides as well.
func NewStorageFactory(storageConfig storagebackend.Config, defaultMediaType string, serializer runtime.StorageSerializer,
	defaultResourceEncoding *serverstorage.DefaultResourceEncodingConfig, storageEncodingOverrides map[string]schema.GroupVersion, resourceEncodingOverrides []schema.GroupVersionResource,
	defaultAPIResourceConfig *serverstorage.ResourceConfig, resourceConfigOverrides utilflag.ConfigurationMap) (*serverstorage.DefaultStorageFactory, error) {

	resourceEncodingConfig := mergeGroupEncodingConfigs(defaultResourceEncoding, storageEncodingOverrides)
	resourceEncodingConfig = mergeResourceEncodingConfigs(resourceEncodingConfig, resourceEncodingOverrides)
	apiResourceConfig, err := resourceconfig.MergeAPIResourceConfigs(defaultAPIResourceConfig, resourceConfigOverrides, api.Scheme)
	if err != nil {
		return nil, err
	}
	return serverstorage.NewDefaultStorageFactory(storageConfig, defaultMediaType, serializer, resourceEncodingConfig, apiResourceConfig, nil), nil
}

// Merges the given defaultResourceConfig with specifc GroupvVersionResource overrides.
func mergeResourceEncodingConfigs(defaultResourceEncoding *serverstorage.DefaultResourceEncodingConfig, resourceEncodingOverrides []schema.GroupVersionResource) *serverstorage.DefaultResourceEncodingConfig {
	resourceEncodingConfig := defaultResourceEncoding
	for _, gvr := range resourceEncodingOverrides {
		resourceEncodingConfig.SetResourceEncoding(gvr.GroupResource(), gvr.GroupVersion(),
			schema.GroupVersion{Group: gvr.Group, Version: runtime.APIVersionInternal})
	}
	return resourceEncodingConfig
}

// Merges the given defaultResourceConfig with specifc GroupVersion overrides.
func mergeGroupEncodingConfigs(defaultResourceEncoding *serverstorage.DefaultResourceEncodingConfig, storageEncodingOverrides map[string]schema.GroupVersion) *serverstorage.DefaultResourceEncodingConfig {
	resourceEncodingConfig := defaultResourceEncoding
	for group, storageEncodingVersion := range storageEncodingOverrides {
		resourceEncodingConfig.SetVersionEncoding(group, storageEncodingVersion, schema.GroupVersion{Group: group, Version: runtime.APIVersionInternal})
	}
	return resourceEncodingConfig
}

func getRuntimeConfigValue(overrides utilflag.ConfigurationMap, apiKey string, defaultValue bool) (bool, error) {
	flagValue, ok := overrides[apiKey]
	if ok {
		if flagValue == "" {
			return true, nil
		}
		boolValue, err := strconv.ParseBool(flagValue)
		if err != nil {
			return false, fmt.Errorf("invalid value of %s: %s, err: %v", apiKey, flagValue, err)
		}
		return boolValue, nil
	}
	return defaultValue, nil
}
