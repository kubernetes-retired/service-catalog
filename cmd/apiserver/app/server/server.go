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

package server

import (
	"flag"
	"fmt"
	"io"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/genericapiserver"
	genericserveroptions "k8s.io/kubernetes/pkg/genericapiserver/options"
	"k8s.io/kubernetes/pkg/util/wait"

	"github.com/kubernetes-incubator/service-catalog/pkg/apiserver"
)

// ServiceCatalogServerOptions contains the aggregation of configuration structs for
// the service-catalog server. The theory here is that any future user
// of this server will be able to use this options object as a sub
// options of its own.
type ServiceCatalogServerOptions struct {
	// the runtime configuration of our server
	GenericServerRunOptions *genericserveroptions.ServerRunOptions
	// the https configuration. certs, etc
	SecureServingOptions *genericserveroptions.SecureServingOptions
	// storage with etcd
	EtcdOptions *genericserveroptions.EtcdOptions
	// authn
	AuthenticationOptions *genericserveroptions.DelegatingAuthenticationOptions
	// authz
	AuthorizationOptions *genericserveroptions.DelegatingAuthorizationOptions
	// InsecureOptions are options for serving insecurely.
	InsecureServingOptions *genericserveroptions.ServingOptions
}

const (
	// Store generated SSL certificates in a place that won't collide with the
	// k8s core API server.
	certDirectory = "/var/run/kubernetes-service-catalog"

	// I made this up to match some existing paths. I am not sure if there
	// are any restrictions on the format or structure beyond text
	// separated by slashes.
	etcdPathPrefix = "/k8s.io/service-catalog"

	// GroupName I made this up. Maybe we'll need it.
	GroupName = "service-catalog.k8s.io"
)

// NewCommandServer creates a new cobra command to run our server.
func NewCommandServer(out io.Writer) *cobra.Command {
	// initalize our sub options.
	options := &ServiceCatalogServerOptions{
		GenericServerRunOptions: genericserveroptions.NewServerRunOptions(),
		SecureServingOptions:    genericserveroptions.NewSecureServingOptions(),
		EtcdOptions:             genericserveroptions.NewEtcdOptions(),
		AuthenticationOptions:   genericserveroptions.NewDelegatingAuthenticationOptions(),
		AuthorizationOptions:    genericserveroptions.NewDelegatingAuthorizationOptions(),
		InsecureServingOptions:  genericserveroptions.NewInsecureServingOptions(),
	}

	// Store resources in etcd under our special prefix
	options.EtcdOptions.StorageConfig.Prefix = etcdPathPrefix

	// Set generated SSL cert path correctly
	options.SecureServingOptions.ServerCert.CertDirectory = certDirectory

	// Create the command that runs the API server
	cmd := &cobra.Command{
		Short: "run a service-catalog server",
		Run: func(c *cobra.Command, args []string) {
			options.RunServer(wait.NeverStop)
		},
	}

	// We pass flags object to sub option structs to have them configure
	// themselves. Each options adds its own command line flags
	// in addition to the flags that are defined above.
	flags := cmd.Flags()
	// TODO consider an AddFlags() method on our options
	// struct. Will need to import pflag.
	//
	// repeated pattern seems like it should be refactored if all
	// options were of an interface type that specified AddFlags.
	flags.AddGoFlagSet(flag.CommandLine)
	options.GenericServerRunOptions.AddUniversalFlags(flags)
	options.SecureServingOptions.AddFlags(flags)
	options.EtcdOptions.AddFlags(flags)
	options.AuthenticationOptions.AddFlags(flags)
	options.AuthorizationOptions.AddFlags(flags)
	options.InsecureServingOptions.AddFlags(flags)

	return cmd
}

// RunServer is a method on the options for composition. Allows embedding in a
// higher level options as we do the etcd and serving options.
func (serverOptions ServiceCatalogServerOptions) RunServer(stopCh <-chan struct{}) error {
	glog.V(4).Infoln("Preparing to run API server")
	// options
	// runtime options
	if err := serverOptions.GenericServerRunOptions.DefaultExternalAddress(serverOptions.SecureServingOptions, nil); err != nil {
		return err
	}

	// server configuration options
	glog.V(4).Infoln("Setting up secure serving options")
	if err := serverOptions.SecureServingOptions.MaybeDefaultWithSelfSignedCerts(serverOptions.GenericServerRunOptions.AdvertiseAddress.String()); err != nil {
		glog.Errorf("Error creating self-signed certificates: %v", err)
		return err
	}

	// etcd options
	if errs := serverOptions.EtcdOptions.Validate(); len(errs) > 0 {
		glog.Errorln("Error validating etcd options, do you have `--etcd-servers localhost` set?")
		return errs[0]
	}

	// config
	glog.V(4).Infoln("Configuring generic API server")
	genericconfig := genericapiserver.NewConfig().ApplyOptions(serverOptions.GenericServerRunOptions)
	// these are all mutators of each specific suboption in serverOptions object.
	// this repeated pattern seems like we could refactor
	if _, err := genericconfig.ApplySecureServingOptions(serverOptions.SecureServingOptions); err != nil {
		glog.Errorln(err)
		return err
	}

	genericconfig.ApplyInsecureServingOptions(serverOptions.InsecureServingOptions)

	glog.V(4).Info("Setting up authn (disabled)")
	// need to figure out what's throwing the `missing clientCA file` err
	/*
		if _, err := genericconfig.ApplyDelegatingAuthenticationOptions(serverOptions.AuthenticationOptions); err != nil {
			glog.Infoln(err)
			return err
		}
	*/

	glog.V(4).Infoln("Setting up authz (disabled)")
	// having this enabled causes the server to crash for any call
	/*
		if _, err := genericconfig.ApplyDelegatingAuthorizationOptions(serverOptions.AuthorizationOptions); err != nil {
			glog.Infoln(err)
			return err
		}
	*/

	glog.V(4).Infoln("Creating storage factory")
	// The API server stores objects using a particular API version for each
	// group, regardless of API version of the object when it was created.
	//
	// storageGroupsToEncodingVersion holds a map of API group to version that
	// the API server uses to store that group.
	storageGroupsToEncodingVersion, err := serverOptions.GenericServerRunOptions.StorageGroupsToEncodingVersion()
	if err != nil {
		return fmt.Errorf("error generating storage version map: %s", err)
	}

	// Build the default storage factory.
	//
	// The default storage factory returns the storage interface for a
	// particular GroupResource (an (api-group, resource) tuple).
	storageFactory, err := genericapiserver.BuildDefaultStorageFactory(
		serverOptions.EtcdOptions.StorageConfig,
		serverOptions.GenericServerRunOptions.DefaultStorageMediaType,
		api.Codecs,
		genericapiserver.NewDefaultResourceEncodingConfig(),
		storageGroupsToEncodingVersion,
		nil, /* group storage version overrides */
		apiserver.DefaultAPIResourceConfigSource(),
		serverOptions.GenericServerRunOptions.RuntimeConfig)
	if err != nil {
		glog.Errorf("error creating storage factory: %v", err)
		return err
	}

	// Set the finalized generic and storage configs
	config := apiserver.Config{
		GenericConfig:  genericconfig,
		StorageFactory: storageFactory,
	}

	// Fill in defaults not already set in the config
	completedconfig := config.Complete()

	// make the server
	glog.V(4).Infoln("Completing API server configuration")
	server, err := completedconfig.New()
	if err != nil {
		return fmt.Errorf("error completing API server configuration: %v", err)
	}

	// I don't like this. We're reaching in too far to call things.
	preparedserver := server.GenericAPIServer.PrepareRun() // post api installation setup? We should have set up the api already?

	glog.Infoln("Running the API server")
	preparedserver.Run(stopCh)

	return nil
}
