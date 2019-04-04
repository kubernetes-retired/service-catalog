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
	"net/http"

	scTypes "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	csbmutation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/clusterservicebroker/mutation"
	cscmutation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/clusterserviceclass/mutation"
	cspmutation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/clusterserviceplan/mutation"

	sbmutation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/servicebinding/mutation"
	brmutation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/servicebroker/mutation"
	scmutation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/serviceclass/mutation"
	simutation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/serviceinstance/mutation"
	spmutation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/serviceplan/mutation"

	csbrvalidation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/clusterservicebroker/validation"
	sbvalidation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/servicebinding/validation"
	sbrvalidation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/servicebroker/validation"
	sivalidation "github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/serviceinstance/validation"

	"github.com/pkg/errors"
	"k8s.io/apiserver/pkg/server/healthz"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// RunServer runs the webhook server with configuration according to opts
func RunServer(opts *WebhookServerOptions, stopCh <-chan struct{}) error {
	if stopCh == nil {
		/* the caller of RunServer should generate the stop channel
		if there is a need to stop the Webhook server */
		stopCh = make(chan struct{})
	}

	if err := opts.Validate(); nil != err {
		return err
	}

	return run(opts, stopCh)
}

func run(opts *WebhookServerOptions, stopCh <-chan struct{}) error {
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		return errors.Wrap(err, "while set up overall controller manager for webhook server")
	}

	scTypes.AddToScheme(mgr.GetScheme())

	// setup webhook server
	webhookSvr := &webhook.Server{
		Port:    opts.SecureServingOptions.BindPort,
		CertDir: opts.SecureServingOptions.ServerCert.CertDirectory,
	}

	webhooks := map[string]admission.Handler{
		"/mutating-clusterservicebrokers": &csbmutation.CreateUpdateHandler{},
		"/mutating-clusterserviceclasses": &cscmutation.CreateUpdateHandler{},
		"/mutating-clusterserviceplans":   &cspmutation.CreateUpdateHandler{},

		"/mutating-servicebindings":  &sbmutation.CreateUpdateHandler{},
		"/mutating-servicebrokers":   &brmutation.CreateUpdateHandler{},
		"/mutating-serviceclasses":   &scmutation.CreateUpdateHandler{},
		"/mutating-serviceplans":     &spmutation.CreateUpdateHandler{},
		"/mutating-serviceinstances": simutation.New(),

		"/validating-clusterservicebrokers": csbrvalidation.NewAdmissionHandler(),
		"/validating-servicebindings":       sbvalidation.NewAdmissionHandler(),
		"/validating-servicebrokers":        sbrvalidation.NewAdmissionHandler(),
		"/validating-serviceinstances":      sivalidation.NewAdmissionHandler(),
	}

	for path, handler := range webhooks {
		webhookSvr.Register(path, &webhook.Admission{Handler: handler})
	}

	// setup healthz server
	healthzSvr := manager.RunnableFunc(func(stopCh <-chan struct{}) error {
		mux := http.NewServeMux()
		// liveness registered at /healthz indicates if the container is responding
		healthz.InstallHandler(mux, healthz.PingHealthz)

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.HealthzServerBindPort),
			Handler: mux,
		}

		return server.ListenAndServe()
	})

	// register servers
	if err := mgr.Add(webhookSvr); err != nil {
		return errors.Wrap(err, "while registering webhook server with manager")
	}

	if err := mgr.Add(healthzSvr); err != nil {
		return errors.Wrap(err, "while registering healthz server with manager")
	}

	// starts the server blocks until the Stop channel is closed
	if err := mgr.Start(stopCh); err != nil {
		return errors.Wrap(err, "while running the webhook manager")

	}

	return nil
}
