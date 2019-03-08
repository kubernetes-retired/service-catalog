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
	"github.com/kubernetes-incubator/service-catalog/cmd/webhook/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/hyperkube"
)

// NewWebhookServer creates a new hyperkube Server object that includes the
// description and flags of the Webhook server.
func NewWebhookServer() *hyperkube.Server {
	opts := server.NewWebhookServerOptions()

	hks := hyperkube.Server{
		PrimaryName:     "webhook",
		AlternativeName: "service-catalog-webhook",
		SimpleUsage:     "webhook",
		Long:            "The Service Catalog Webhook server which manages fields defaulting and validating of the Service Catalog CRDs.",
		Run: func(_ *hyperkube.Server, args []string, stopCh <-chan struct{}) error {
			return server.RunServer(opts, stopCh)
		},
		RespectsStopCh: true,
	}
	opts.AddFlags(hks.Flags())

	return &hks
}
