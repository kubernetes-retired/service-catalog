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

package server

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apiserver"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/tpr"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/genericapiserver"
)

// RunServer runs an API server with configuration according to opts
func RunServer(opts *ServiceCatalogServerOptions) error {
	storageType, err := opts.StorageType()
	if err != nil {
		return err
	}
	if storageType == server.StorageTypeTPR {
		return runTPRServer(opts)
	}
	return runEtcdServer(opts)
}

func runTPRServer(opts *ServiceCatalogServerOptions) error {
	tprOpts := opts.TPROptions
	glog.Infoln("Installing TPR types to the cluster")
	tprInstaller := tpr.NewInstaller(tprOpts.clIface.Extensions().ThirdPartyResources())
	if err := tprInstaller.InstallTypes(); err != nil {
		glog.V(4).Infof("Installing TPR types failed, continuing anyway (%s)", err)
	}
	glog.V(4).Infoln("Preparing to run API server")
	genericConfig, err := setupBasicServer(opts)
	if err != nil {
		return err
	}

	config := apiserver.NewTPRConfig(
		tprOpts.clIface,
		genericConfig,
		tprOpts.globalNamespace,
		tprOpts.storageFactory(),
	)
	completed := config.Complete()
	// make the server
	glog.V(4).Infoln("Completing API server configuration")
	server, err := completed.NewServer()
	if err != nil {
		return fmt.Errorf("error completing API server configuration: %v", err)
	}

	glog.Infoln("Running the API server")
	stop := make(chan struct{})
	server.GenericAPIServer.PrepareRun().Run(stop)

	return nil
}

func runEtcdServer(opts *ServiceCatalogServerOptions) error {
	etcdOpts := opts.EtcdOptions
	glog.V(4).Infoln("Preparing to run API server")
	genericConfig, err := setupBasicServer(opts)
	if err != nil {
		return err
	}

	// etcd options
	if errs := etcdOpts.Validate(); len(errs) > 0 {
		glog.Errorln("Error validating etcd options, do you have `--etcd-servers localhost` set?")
		return errs[0]
	}

	glog.V(4).Infoln("Creating storage factory")
	// The API server stores objects using a particular API version for each
	// group, regardless of API version of the object when it was created.
	//
	// storageGroupsToEncodingVersion holds a map of API group to version that
	// the API server uses to store that group.
	storageGroupsToEncodingVersion, err := opts.GenericServerRunOptions.StorageGroupsToEncodingVersion()
	if err != nil {
		return fmt.Errorf("error generating storage version map: %s", err)
	}

	// Build the default storage factory.
	//
	// The default storage factory returns the storage interface for a
	// particular GroupResource (an (api-group, resource) tuple).
	storageFactory, err := genericapiserver.BuildDefaultStorageFactory(
		etcdOpts.StorageConfig,
		opts.GenericServerRunOptions.DefaultStorageMediaType,
		api.Codecs,
		genericapiserver.NewDefaultResourceEncodingConfig(),
		storageGroupsToEncodingVersion,
		nil, /* group storage version overrides */
		apiserver.DefaultAPIResourceConfigSource(),
		opts.GenericServerRunOptions.RuntimeConfig,
	)
	if err != nil {
		glog.Errorf("error creating storage factory: %v", err)
		return err
	}

	// Set the finalized generic and storage configs
	config := apiserver.NewEtcdConfig(genericConfig, 0 /* deleteCollectionWorkers */, storageFactory)

	// Fill in defaults not already set in the config
	completed := config.Complete()

	// make the server
	glog.V(4).Infoln("Completing API server configuration")
	server, err := completed.NewServer()
	if err != nil {
		return fmt.Errorf("error completing API server configuration: %v", err)
	}

	// do we need to do any post api installation setup? We should have set up the api already?
	glog.Infoln("Running the API server")
	stop := make(chan struct{})
	server.PrepareRun().Run(stop)

	return nil
}
