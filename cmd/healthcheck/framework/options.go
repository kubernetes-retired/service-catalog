/*
Copyright 2018 The Kubernetes Authors.

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

package framework

import (
	"os"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"

	genericoptions "k8s.io/apiserver/pkg/server/options"
)

// HealthCheckServer is the main context object for the health check server
type HealthCheckServer struct {
	KubeHost    string
	KubeConfig  string
	KubeContext string

	// HealthCheckInterval is how frequently the end to end health check should be run
	HealthCheckInterval  time.Duration
	SecureServingOptions *genericoptions.SecureServingOptions
	TestBrokerName       string
}

const (
	defaultHealthCheckInterval = 2 * time.Minute
	defaultSecurePort          = 443
	defaultCertDirectory       = "/var/run/service-catalog-healthcheck"
)

// NewHealthCheckServer creates a new HealthCheckServer with a default config.
func NewHealthCheckServer() *HealthCheckServer {
	s := HealthCheckServer{
		HealthCheckInterval:  defaultHealthCheckInterval,
		SecureServingOptions: genericoptions.NewSecureServingOptions(),
	}
	s.SecureServingOptions.BindPort = defaultSecurePort
	s.SecureServingOptions.ServerCert.CertDirectory = defaultCertDirectory
	return &s
}

// AddFlags adds flags for a ControllerManagerServer to the specified FlagSet.
func (s *HealthCheckServer) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.KubeHost, "kubernetes-host", "http://127.0.0.1:8080", "The kubernetes host, or apiserver, to connect to")
	fs.StringVar(&s.KubeConfig, "kubernetes-config", os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "Path to config containing embedded authinfo for kubernetes. Default value is from environment variable "+clientcmd.RecommendedConfigPathEnvVar)
	fs.StringVar(&s.KubeContext, "kubernetes-context", "", "config context to use for kuberentes. If unset, will use value from 'current-context'")
	fs.DurationVar(&s.HealthCheckInterval, "healthcheck-interval", s.HealthCheckInterval, "How frequently the end to end health check should be performed")
	fs.StringVar(&s.TestBrokerName, "broker-name", "ups-broker", "Broker Name to test against - can only be ups-broker or osb-stub.  You must ensure the specified broker is deployed.")
	s.SecureServingOptions.AddFlags(fs)
}
