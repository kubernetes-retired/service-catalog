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

	// glog.V(4).Info("Setting up authn (disabled)")
	// need to figure out what's throwing the `missing clientCA file` err
	/*
		if _, err := genericConfig.ApplyDelegatingAuthenticationOptions(serverOptions.AuthenticationOptions); err != nil {
			glog.Infoln(err)
			return err
		}
	*/

	// glog.V(4).Infoln("Setting up authz (disabled)")
	// having this enabled causes the server to crash for any call
	/*
		if _, err := genericConfig.ApplyDelegatingAuthorizationOptions(serverOptions.AuthorizationOptions); err != nil {
			glog.Infoln(err)
			return err
		}
	*/
	return genericConfig, nil
}
