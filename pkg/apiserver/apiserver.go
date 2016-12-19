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
	//"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/genericapiserver"
	"k8s.io/kubernetes/pkg/version"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
)

// ServiceCatalogAPIServer contains base GenericAPIServer along with
// other configured runtime confiuration
type ServiceCatalogAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

// Config contains our base generic Config along with config specific
// to the service catalog.
type Config struct {
	GenericConfig *genericapiserver.Config
}

// CompletedConfig is an internal type to take advantage of
// typechecking in the type system. mhb does not like it.
type CompletedConfig struct {
	*Config
}

// Complete fills in any fields not set that are required to have
// valid data and can be derived from other fields.
func (c *Config) Complete() CompletedConfig {
	c.GenericConfig.Complete()

	version := version.Get()
	// Setting this var enables the version resource. We should
	// populate the fields of the object from above if we wish to
	// have our own output. Or establish our own version object
	// somewhere else.
	c.GenericConfig.Version = &version

	return CompletedConfig{c}
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

	glog.Infoln("Creating the server to run")

	s := &ServiceCatalogAPIServer{
		GenericAPIServer: genericServer,
	}

	glog.Infoln("make and install the apis")

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(servicecatalog.GroupName)
	// giving it v1alpha1 version
	apiGroupInfo.GroupMeta.GroupVersion = v1alpha1.SchemeGroupVersion

	// TODO make storage work
	//v1alpha1storage := map[string]rest.Storage{}
	//
	//v1alpha1storage["servicecatalog"] = apiservice.NewREST(c.RESTOptionsGetter.NewFor(apiregistration.Resource("servicecatalog")))

	//apiGroupInfo.VersionedResourcesStorageMap[v1alpha1.SchemeGroupVersion.Version] = v1alpha1storage
	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	glog.Infoln("returning the server")
	return s, nil
}
