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
	"io"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/model"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/apiserver"
	"github.com/kubernetes-incubator/service-catalog/util"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/genericapiserver"
	genericserveroptions "k8s.io/kubernetes/pkg/genericapiserver/options"
)

// ServiceCatalogServerOptions contains the aggregation of configuration structs for
// the service-catalog server. The theory here is that any future user
// of this server will be able to use this options object as a sub
// options of it's own.
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
}

const (
	// I made this up to match some existing paths. I am not sure if there
	// are any restrictions on the format or structure beyond text
	// separated by slashes.
	etcdPathPrefix = "/k8s.io/incubator/service-catalog"
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
	}

	// when our etcd resources are created, put them in a location
	// specific to us
	options.EtcdOptions.StorageConfig.Prefix = etcdPathPrefix

	// I have no idea what this line does but I keep seeing it. Do
	// we have our own `GroupName`? Is it like the pathPrefix for
	// etcd?
	//
	// options.EtcdOptions.StorageConfig.Codec = api.Codecs.LegacyCodec(registered.EnabledVersionsForGroup(api.GroupName)...)

	// this is the one thing that this program does. It runs the apiserver.
	cmd := &cobra.Command{
		Short: "run a service-catalog server",
		Run: func(c *cobra.Command, args []string) {
			options.runServer()
		},
	}

	// We pass flags object to sub option structs to have them configure
	// themselves. Each options adds it's own command line flags
	// in addition to the flags that are defined above.
	flags := cmd.Flags()
	// TODO consider an AddFlags() method on our options
	// struct. Will need to import pflag.
	//
	// repeated pattern seems like it should be refactored if all
	// options were of an interface type that specified AddFlags.
	options.GenericServerRunOptions.AddUniversalFlags(flags)
	options.SecureServingOptions.AddFlags(flags)
	options.EtcdOptions.AddFlags(flags)
	options.AuthenticationOptions.AddFlags(flags)
	options.AuthorizationOptions.AddFlags(flags)

	return cmd
}

// runServer is a method on the options for composition. allows embedding in a higher level options as we do the etcd and serving options.
func (serverOptions ServiceCatalogServerOptions) runServer() error {
	glog.Infoln("set up the server")
	// options
	// runtime options
	if err := serverOptions.GenericServerRunOptions.DefaultExternalAddress(serverOptions.SecureServingOptions, nil); err != nil {
		return err
	}
	// server configuration options
	glog.Infoln("set up serving options")
	if err := serverOptions.SecureServingOptions.MaybeDefaultWithSelfSignedCerts(serverOptions.GenericServerRunOptions.AdvertiseAddress.String()); err != nil {
		glog.Errorf("Error creating self-signed certificates: %v", err)
		return err
	}

	// etcd options
	if errs := serverOptions.EtcdOptions.Validate(); len(errs) > 0 {
		glog.Errorln("set up etcd options, do you have `--etcd-servers localhost` set?")
		return errs[0]
	}

	// config
	glog.Infoln("set up config object")
	genericconfig := genericapiserver.NewConfig().ApplyOptions(serverOptions.GenericServerRunOptions)
	// these are all mutators of each specific suboption in serverOptions object.
	// this repeated pattern seems like we could refactor
	if _, err := genericconfig.ApplySecureServingOptions(serverOptions.SecureServingOptions); err != nil {
		glog.Errorln(err)
		return err
	}

	// need to figure out what's throwing the `missing clientCA file` err
	/*
		if _, err := genericconfig.ApplyDelegatingAuthenticationOptions(serverOptions.AuthenticationOptions); err != nil {
			glog.Infoln(err)
			return err
		}
	*/
	// having this enabled causes the server to crash for any call
	/*
		if _, err := genericconfig.ApplyDelegatingAuthorizationOptions(serverOptions.AuthorizationOptions); err != nil {
			glog.Infoln(err)
			return err
		}
	*/

	// configure our own apiserver using the preconfigured genericApiServer
	config := apiserver.Config{
		GenericConfig: genericconfig,
	}

	// finish config
	completedconfig := config.Complete()

	// make the server
	glog.Infoln("make the server")
	server, err := completedconfig.New()
	if err != nil {
		return err
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(v1alpha1.GroupNameString)
	apiGroupInfo.GroupMeta.GroupVersion = v1alpha1.GroupVersion
	apiGroupInfo.VersionedResourcesStorageMap = map[string]map[string]rest.Storage{
		v1alpha1.GroupVersion.Version: map[string]rest.Storage{
			util.FormatResourceKind(model.ServiceBrokerKind):   v1alpha1.NewServiceBrokerStorage(),
			util.FormatResourceKind(model.ServiceBindingKind):  v1alpha1.NewServiceBindingStorage(),
			util.FormatResourceKind(model.ServiceInstanceKind): v1alpha1.NewServiceInstanceStorage(),
		},
	}
	// TODO: do more API group setup before installing it
	// apiGroupInfo.GroupMeta.GroupVersion = projectapiv1.SchemeGroupVersion
	if err := server.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return err
	}

	// I don't like this. We're reaching in too far to call things.
	preparedserver := server.GenericAPIServer.PrepareRun() // post api installation setup? We should have set up the api already?

	stop := make(chan struct{})
	glog.Infoln("run the server")
	preparedserver.Run(stop)
	return nil
}
