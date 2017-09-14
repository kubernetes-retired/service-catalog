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
	"github.com/spf13/pflag"
	"k8s.io/client-go/pkg/api"
	restclient "k8s.io/client-go/rest"

	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
)

// CRDOptions contains the complete configuration for an API server that
// communicates with the core Kubernetes API server to use Custom Resources (CRs)
// as a database. It is exported so that integration tests can use it
type CRDOptions struct {
	RESTClient      restclient.Interface
	InstallCRDsFunc func() error
}

// NewCRDOptions creates a new, empty CRDOptions struct
func NewCRDOptions() *CRDOptions {
	return &CRDOptions{}
}

// NewStorageFactory returns a new StorageFactory from the config in opts
func (s *CRDOptions) storageFactory() serverstorage.StorageFactory {
	return serverstorage.NewDefaultStorageFactory(
		storagebackend.Config{},
		"application/json",
		api.Codecs,
		serverstorage.NewDefaultResourceEncodingConfig(api.Registry),
		serverstorage.NewResourceConfig(),
	)
}

func (s *CRDOptions) addFlags(fs *pflag.FlagSet) {
	// No CRD-specific flags
}
