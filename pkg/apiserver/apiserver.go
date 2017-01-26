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

package apiserver

import (
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/genericapiserver"
	"k8s.io/kubernetes/pkg/registry"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/version"

	servicecatalogv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	servicecatalogrest "github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/rest"
)

// ServiceCatalogAPIServer contains the base GenericAPIServer along with other
// configured runtime configuration
type ServiceCatalogAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

// Config contains a generic API server Config along with config specific to
// the service catalog API server.
type Config struct {
	GenericConfig *genericapiserver.Config

	// BABYNETES: cargo culted from master.go

	APIResourceConfigSource genericapiserver.APIResourceConfigSource
	DeleteCollectionWorkers int
	StorageFactory          genericapiserver.StorageFactory
}

// CompletedConfig is an internal type to take advantage of typechecking in
// the type system. mhb does not like it.
type CompletedConfig struct {
	*Config
}

// Complete fills in any fields not set that are required to have valid data
// and can be derived from other fields.
func (c *Config) Complete() CompletedConfig {
	c.GenericConfig.Complete()

	version := version.Get()
	// Setting this var enables the version resource. We should populate the
	// fields of the object from above if we wish to have our own output. Or
	// establish our own version object somewhere else.
	c.GenericConfig.Version = &version

	return CompletedConfig{c}
}

// RESTStorageProvider is a local interface describing a REST storage factory.
// It can report the name of the API group and create a new storage interface
// for it.
type RESTStorageProvider interface {
	// GroupName returns the API group name
	GroupName() string
	// NewRESTStorage returns a new
	NewRESTStorage(apiResourceConfigSource genericapiserver.APIResourceConfigSource, restOptionsGetter registry.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool)
}

// New creates the server to run.
func (c CompletedConfig) New() (*ServiceCatalogAPIServer, error) {
	// we need to call new on a "completed" config, which we
	// should already have, as this is a 'CompletedConfig' and the
	// only way to get here from there is by Complete()'ing. Thus
	// we skip the complete on the underlying config and go
	// straight to running it's New() method.
	genericServer, err := c.Config.GenericConfig.SkipComplete().New()
	if err != nil {
		return nil, err
	}

	glog.V(4).Infoln("Creating API server")

	s := &ServiceCatalogAPIServer{
		GenericAPIServer: genericServer,
	}

	// Not every API group compiled in is necessarily enabled by the operator
	// at runtime.
	//
	// Install the API resource config source, which describes versions of
	// which API groups are enabled.
	c.APIResourceConfigSource = DefaultAPIResourceConfigSource()

	restStorageProviders := []RESTStorageProvider{
		servicecatalogrest.StorageProvider{},
	}

	restOptionsFactory := restOptionsFactory{
		deleteCollectionWorkers: c.DeleteCollectionWorkers,
		enableGarbageCollection: c.GenericConfig.EnableGarbageCollection,
		storageFactory:          c.StorageFactory,
		storageDecorator:        generic.UndecoratedStorage,
	}

	glog.V(4).Infoln("Installing API groups")
	for _, provider := range restStorageProviders {
		groupInfo, enabled := provider.NewRESTStorage(c.Config.APIResourceConfigSource, restOptionsFactory.NewFor)
		if !enabled {
			glog.Warningf("Skipping API group %v because it is not enabled", provider.GroupName())
		}

		glog.V(4).Infof("Installing API group %v", provider.GroupName())
		if err := s.GenericAPIServer.InstallAPIGroup(&groupInfo); err != nil {
			glog.Fatalf("Error installing API group %v: %v", provider.GroupName(), err)
		}
	}

	glog.Infoln("Finished installing API groups")

	return s, nil
}

// BABYNETES: had to be lifted from pkg/master/master.go

// restOptionsFactory is an object that provides a factory method for getting
// the REST options for a particular GroupResource.
type restOptionsFactory struct {
	deleteCollectionWorkers int
	enableGarbageCollection bool
	storageFactory          genericapiserver.StorageFactory
	storageDecorator        generic.StorageDecorator
}

// NewFor returns the RESTOptions for a particular GroupResource.
func (f restOptionsFactory) NewFor(resource schema.GroupResource) generic.RESTOptions {
	storageConfig, err := f.storageFactory.NewConfig(resource)
	if err != nil {
		glog.Fatalf("Unable to find storage destination for %v, due to %v", resource, err.Error())
	}

	return generic.RESTOptions{
		StorageConfig:           storageConfig,
		Decorator:               f.storageDecorator,
		DeleteCollectionWorkers: f.deleteCollectionWorkers,
		EnableGarbageCollection: f.enableGarbageCollection,
		ResourcePrefix:          f.storageFactory.ResourcePrefix(resource),
	}
}

// DefaultAPIResourceConfigSource returns a default API Resource config source
func DefaultAPIResourceConfigSource() *genericapiserver.ResourceConfig {
	ret := genericapiserver.NewResourceConfig()
	ret.EnableVersions(
		servicecatalogv1alpha1.SchemeGroupVersion,
	)

	return ret
}
