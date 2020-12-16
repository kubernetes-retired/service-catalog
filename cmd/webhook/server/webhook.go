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

	scTypes "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/util"
	csbmutation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/clusterservicebroker/mutation"
	cscmutation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/clusterserviceclass/mutation"
	cspmutation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/clusterserviceplan/mutation"

	sbmutation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/servicebinding/mutation"
	brmutation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/servicebroker/mutation"
	scmutation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/serviceclass/mutation"
	simutation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/serviceinstance/mutation"
	spmutation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/serviceplan/mutation"

	csbrvalidation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/clusterservicebroker/validation"
	cscvalidation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/clusterserviceclass/validation"
	cspvalidation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/clusterserviceplan/validation"
	sbvalidation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/servicebinding/validation"
	sbrvalidation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/servicebroker/validation"
	scvalidation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/serviceclass/validation"
	sivalidation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/serviceinstance/validation"
	spvalidation "github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/serviceplan/validation"

	"github.com/kubernetes-sigs/service-catalog/pkg/probe"
	"github.com/pkg/errors"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
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
	cfg, err := config.GetConfig()
	if err != nil {
		return errors.Wrap(err, "while getting Kubernetes client config")
	}

	apiextensionsClient, err := apiextensionsclientset.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "while create apiextension clientset")
	}

	// It may take some time before Catalog CRDs registration shows up in main API Server.
	// We can start Service Catalog clients/informers only when CRDs are available.
	if err := util.WaitForServiceCatalogCRDs(cfg); err != nil {
		return fmt.Errorf("while waiting for ready Service Catalog CRDs: %v", err)
	}

	mgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress: fmt.Sprintf(":%d", opts.ControllerManagerMetricsPort)})
	if err != nil {
		return errors.Wrap(err, "while set up overall controller manager for webhook server")
	}

	err = scTypes.AddToScheme(mgr.GetScheme())
	if err != nil {
		return errors.Wrap(err, "while register Service Catalog scheme into manager")
	}

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
		"/mutating-serviceinstances": simutation.NewCreateUpdateHandler(),

		"/validating-clusterservicebrokers":        csbrvalidation.NewSpecValidationHandler(),
		"/validating-clusterservicebrokers/status": &csbrvalidation.StatusValidationHandler{},
		"/validating-clusterserviceclasses":        cscvalidation.NewSpecValidationHandler(),
		"/validating-clusterserviceplans":          cspvalidation.NewSpecValidationHandler(),

		"/validating-servicebindings":        sbvalidation.NewSpecValidationHandler(),
		"/validating-servicebindings/status": &sbvalidation.StatusValidationHandler{},
		"/validating-servicebrokers":         sbrvalidation.NewSpecValidationHandler(),
		"/validating-servicebrokers/status":  &sbrvalidation.StatusValidationHandler{},
		"/validating-serviceclasses":         scvalidation.NewSpecValidationHandler(),
		"/validating-serviceplans":           spvalidation.NewSpecValidationHandler(),
		"/validating-serviceinstances":       sivalidation.NewSpecValidationHandler(),
	}

	for path, handler := range webhooks {
		webhookSvr.Register(path, &webhook.Admission{Handler: handler})
	}

	// setup healthz server
	healthzSvr := manager.RunnableFunc(func(stopCh <-chan struct{}) error {
		mux := http.NewServeMux()

		// readiness registered at /healthz/ready indicates if traffic should be routed to this container
		healthz.InstallPathHandler(mux, "/healthz/ready", probe.NewCRDProbe(apiextensionsClient, probe.CRDProbeIterationGap))

		// liveness registered at /healthz indicates if the container is responding
		healthz.InstallHandler(mux, healthz.PingHealthz, probe.NewCRDProbe(apiextensionsClient, probe.CRDProbeIterationGap))

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
