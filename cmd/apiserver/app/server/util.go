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
	"github.com/golang/glog"

	genericapiserver "k8s.io/apiserver/pkg/server"
)

func setupBasicServer(s *ServiceCatalogServerOptions) (*genericapiserver.Config, error) {
	if _, err := s.SecureServingOptions.ServingOptions.DefaultExternalAddress(); err != nil {
		return nil, err
	}

	// server configuration options
	if err := s.SecureServingOptions.MaybeDefaultWithSelfSignedCerts(
		s.GenericServerRunOptions.AdvertiseAddress.String(),
	); err != nil {
		return nil, err
	}
	genericConfig := genericapiserver.NewConfig()
	if err := s.GenericServerRunOptions.ApplyTo(genericConfig); err != nil {
		return nil, err
	}
	if err := s.SecureServingOptions.ApplyTo(genericConfig); err != nil {
		return nil, err
	}

	if err := s.InsecureServingOptions.ApplyTo(genericConfig); err != nil {
		return nil, err
	}

	if !s.DisableAuth {
		if err := s.AuthenticationOptions.ApplyTo(genericConfig); err != nil {
			return nil, err
		}

		if err := s.AuthorizationOptions.ApplyTo(genericConfig); err != nil {
			return nil, err
		}
	} else {
		// always warn when auth is disabled, since this should only be used for testing
		glog.Infof("Authentication and authorization disabled for testing purposes")
	}

	return genericConfig, nil
}
