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

// The controller is responsible for running control loops that reconcile
// the state of service catalog API resources with service brokers, service
// classes, service instances, and service bindings.

package options

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/componentconfig"
	"github.com/spf13/pflag"
	k8scomponentconfig "k8s.io/kubernetes/pkg/apis/componentconfig"
)

// ControllerServer is the main context object for the controller.
type ControllerServer struct {
	componentconfig.ControllerConfiguration
}

// NewControllerServer creates a new ControllerServer with a default config.
func NewControllerServer() *ControllerServer {
	return &ControllerServer{
		ControllerConfiguration: componentconfig.ControllerConfiguration{
			Address:        "0.0.0.0",
			Port:           10000,
			KubeconfigPath: "./kubeconfig",
		},
	}
}

// AddFlags adds flags for a specific CMServer to the specified FlagSet
func (s *ControllerServer) AddFlags(fs *pflag.FlagSet) {
	fs.Var(k8scomponentconfig.IPVar{Val: &s.Address}, "address", "The IP address to serve on (set to 0.0.0.0 for all interfaces)")
	fs.Int32Var(&s.Port, "port", s.Port, "The port that the controller-manager's http service runs on")
	fs.StringVar(&s.KubeconfigPath, "kubeconfig", "./kubeconfig", "Path to kubeconfig")
}
