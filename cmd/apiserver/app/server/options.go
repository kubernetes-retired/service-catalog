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
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	"github.com/spf13/pflag"
	genericserveroptions "k8s.io/kubernetes/pkg/genericapiserver/options"
)

// ServiceCatalogServerOptions contains the aggregation of configuration structs for
// the service-catalog server. It contains everything needed to configure a basic API server.
// It is public so that integration tests can access it.
type ServiceCatalogServerOptions struct {
	StorageTypeString string
	// the runtime configuration of our server
	GenericServerRunOptions *genericserveroptions.ServerRunOptions
	// the https configuration. certs, etc
	SecureServingOptions *genericserveroptions.SecureServingOptions
	// authn for the API
	AuthenticationOptions *genericserveroptions.DelegatingAuthenticationOptions
	// authz for the API
	AuthorizationOptions *genericserveroptions.DelegatingAuthorizationOptions
	// InsecureOptions are options for serving insecurely.
	InsecureServingOptions *genericserveroptions.ServingOptions
	// EtcdOptions are options for serving with etcd as the backing store
	EtcdOptions *EtcdOptions
	// TPROptions are options for serving with TPR as the backing store
	TPROptions *TPROptions
}

func (s *ServiceCatalogServerOptions) addFlags(flags *pflag.FlagSet) {
	flags.StringVar(
		&s.StorageTypeString,
		"storage-type",
		"service-catalog",
		"The type of backing storage this API server should use",
	)

	s.GenericServerRunOptions.AddUniversalFlags(flags)
	s.SecureServingOptions.AddFlags(flags)
	s.AuthenticationOptions.AddFlags(flags)
	s.AuthorizationOptions.AddFlags(flags)
	s.InsecureServingOptions.AddFlags(flags)
	s.EtcdOptions.addFlags(flags)
	s.TPROptions.addFlags(flags)
}

// StorageType returns the storage type configured on s, or a non-nil error if s holds an
// invalid storage type
func (s *ServiceCatalogServerOptions) StorageType() (server.StorageType, error) {
	return server.StorageTypeFromString(s.StorageTypeString)
}
