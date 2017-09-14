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
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/storage"
	restclient "k8s.io/client-go/rest"
)

// crdConfig is the configuration needed to run the API server in CRD storage mode
type crdConfig struct {
	restClient     restclient.Interface
	genericConfig  *genericapiserver.Config
	storageFactory storage.StorageFactory
}

// NewCRDConfig returns a new Config for a server that is backed by CRD storage
func NewCRDConfig(
	restClient restclient.Interface,
	genericCfg *genericapiserver.Config,
	factory storage.StorageFactory,
) Config {
	return &crdConfig{
		restClient:     restClient,
		genericConfig:  genericCfg,
		storageFactory: factory,
	}
}

// Complete fills in the remaining fields of t and returns a completed config
func (t *crdConfig) Complete() CompletedConfig {
	completeGenericConfig(t.genericConfig)
	return &completedCRDConfig{
		restClient: t.restClient,
		crdConfig:  t,
		// Not every API group compiled in is necessarily enabled by the operator
		// at runtime.
		//
		// Install the API resource config source, which describes versions of
		// which API groups are enabled.
		apiResourceConfigSource: DefaultAPIResourceConfigSource(),
		factory:                 t.storageFactory,
	}
}

// CompletedCRDConfig is the completed version of the CRD config. It can be used to create a
// new server, ready to be run
type completedCRDConfig struct {
	restClient restclient.Interface
	*crdConfig
	apiResourceConfigSource storage.APIResourceConfigSource
	factory                 storage.StorageFactory
}

// NewServer returns a new service catalog server, that is ready for execution
func (c *completedCRDConfig) NewServer() (*ServiceCatalogAPIServer, error) {
	s, err := createSkeletonServer(c.crdConfig.genericConfig)
	if err != nil {
		return nil, err
	}
	glog.V(4).Infoln("Created skeleton API server. Installing API groups")

	roFactory := crdRESTOptionsFactory{
		storageFactory: c.factory,
	}

	providers := restStorageProviders(metav1.NamespaceDefault, server.StorageTypeCRD, c.restClient)
	for _, provider := range providers {
		groupInfo, err := provider.NewRESTStorage(
			c.apiResourceConfigSource, // genericapiserver.APIResourceConfigSource
			roFactory,                 // registry.RESTOptionsGetter
		)
		if IsErrAPIGroupDisabled(err) {
			glog.Warningf("Skipping API group %v because it is not enabled", provider.GroupName())
			continue
		} else if err != nil {
			return nil, err
		}
		glog.V(4).Infof("Installing API group %v", provider.GroupName())
		if err := s.GenericAPIServer.InstallAPIGroup(groupInfo); err != nil {
			glog.Fatalf("Error installing API group %v: %v", provider.GroupName(), err)
		}
	}
	glog.Infoln("Finished installing API groups")
	return s, nil
}
