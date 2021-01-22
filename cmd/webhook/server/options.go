/*
Copyright 2019 The Kubernetes Authors.

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

	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	genericserveroptions "k8s.io/apiserver/pkg/server/options"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
)

const (
	certDirectory                       = "/var/run/service-catalog-webhook"
	defaultWebhookServerPort            = 8444
	defaultHealthzServerPort            = 8080
	defaultControllerManagerMetricsPort = 8082
)

// WebhookServerOptions holds configuration for mutating/validating webhook server.
type WebhookServerOptions struct {
	SecureServingOptions         *genericserveroptions.SecureServingOptions
	ReleaseName                  string
	HealthzServerBindPort        int
	ControllerManagerMetricsPort int
}

// NewWebhookServerOptions creates a new WebhookServerOptions with a default settings.
func NewWebhookServerOptions() *WebhookServerOptions {
	opt := WebhookServerOptions{
		SecureServingOptions: genericserveroptions.NewSecureServingOptions(),
	}

	// set defaults, these can be overridden by user specified flags
	opt.SecureServingOptions.BindPort = defaultWebhookServerPort
	opt.SecureServingOptions.ServerCert.CertDirectory = certDirectory

	return &opt
}

// AddFlags adds flags for a WebhookServerOptions to the specified FlagSet.
func (s *WebhookServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&s.HealthzServerBindPort, "healthz-server-bind-port", defaultHealthzServerPort, "The port on which to serve HTTP  /healthz endpoint")
	fs.IntVar(&s.ControllerManagerMetricsPort, "controller-manager-metrics-bind-port", defaultControllerManagerMetricsPort, "The address the metric endpoint binds to")

	s.SecureServingOptions.AddFlags(fs)
	utilfeature.DefaultMutableFeatureGate.AddFlag(fs)
}

// Validate checks all subOptions flags have been set and that they
// have not been set in a conflictory manner.
func (s *WebhookServerOptions) Validate() error {
	var errors []error
	errors = append(errors, s.SecureServingOptions.Validate()...)

	if s.SecureServingOptions.BindPort == s.HealthzServerBindPort {
		errors = append(errors, fmt.Errorf("validation erorr: --secure-port and --healthz-server-bind-port MUST have different values"))
	}

	return utilerrors.NewAggregate(errors)
}
