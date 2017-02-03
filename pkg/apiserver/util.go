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
	servicecatalogrest "github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/rest"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/genericapiserver"
	"k8s.io/kubernetes/pkg/version"
)

func restStorageProviders(
	defaultNamespace string,
	storageType server.StorageType,
	client clientset.Interface,
) []RESTStorageProvider {
	return []RESTStorageProvider{
		servicecatalogrest.StorageProvider{
			DefaultNamespace: defaultNamespace,
			StorageType:      storageType,
			Client:           client,
		},
	}
}

func completeGenericConfig(cfg *genericapiserver.Config) {
	cfg.Complete()

	version := version.Get()
	// Setting this var enables the version resource. We should populate the
	// fields of the object from above if we wish to have our own output. Or
	// establish our own version object somewhere else.
	cfg.Version = &version
}

func createSkeletonServer(genericCfg *genericapiserver.Config) (*ServiceCatalogAPIServer, error) {
	// we need to call new on a "completed" config, which we
	// should already have, as this is a 'CompletedConfig' and the
	// only way to get here from there is by Complete()'ing. Thus
	// we skip the complete on the underlying config and go
	// straight to running it's New() method.
	genericServer, err := genericCfg.SkipComplete().New()
	if err != nil {
		return nil, err
	}

	return &ServiceCatalogAPIServer{
		GenericAPIServer: genericServer,
	}, nil
}
