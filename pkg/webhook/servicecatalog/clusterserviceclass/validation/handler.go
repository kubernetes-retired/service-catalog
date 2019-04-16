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

package validation

import (
	"context"
	"net/http"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"

	admissionTypes "k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Validator is used to implement new validation logic
type Validator interface {
	Validate(context.Context, admission.Request, *sc.ClusterServiceClass, *webhookutil.TracedLogger) *webhookutil.WebhookError
}

// AdmissionHandler handles ServiceInstance validation
type AdmissionHandler struct {
	decoder *admission.Decoder
	client  client.Client

	CreateValidators []Validator
	UpdateValidators []Validator
}

var _ admission.Handler = &AdmissionHandler{}
var _ admission.DecoderInjector = &AdmissionHandler{}
var _ inject.Client = &AdmissionHandler{}

// NewAdmissionHandler creates new AdmissionHandler and initializes validators list
func NewAdmissionHandler() *AdmissionHandler {
	return &AdmissionHandler{
		CreateValidators: []Validator{&StaticCreate{}},
		UpdateValidators: []Validator{&StaticUpdate{}},
	}
}

// Handle handles admission requests.
func (h *AdmissionHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start handling validation operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)

	csc := &sc.ClusterServiceClass{}
	if err := webhookutil.MatchKinds(csc, req.Kind); err != nil {
		traced.Errorf("Error matching kinds: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := h.decoder.Decode(req, csc); err != nil {
		traced.Errorf("Could not decode request object: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	traced.Infof("start validation process for %s: %s/%s", csc.Kind, csc.Namespace, csc.Name)

	var err *webhookutil.WebhookError

	switch req.Operation {
	case admissionTypes.Create:
		for _, v := range h.CreateValidators {
			err = v.Validate(ctx, req, csc, traced)
			if err != nil {
				break
			}
		}
	case admissionTypes.Update:
		for _, v := range h.UpdateValidators {
			err = v.Validate(ctx, req, csc, traced)
			if err != nil {
				break
			}
		}
	default:
		traced.Infof("ClusterServiceBroker validation wehbook does not support action %q", req.Operation)
		return admission.Allowed("action not taken")
	}

	if err != nil {
		switch err.Code() {
		case http.StatusForbidden:
			return admission.Denied(err.Error())
		default:
			return admission.Errored(err.Code(), err)
		}
	}

	traced.Infof("Completed successfully validation operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)
	return admission.Allowed("ClusterServiceClass AdmissionHandler successful")
}

// InjectDecoder injects the decoder into the handlers
func (h *AdmissionHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d

	for _, v := range h.CreateValidators {
		admission.InjectDecoderInto(d, v)
	}
	for _, v := range h.UpdateValidators {
		admission.InjectDecoderInto(d, v)
	}

	return nil
}

// InjectClient injects the client into the handlers
func (h *AdmissionHandler) InjectClient(c client.Client) error {
	h.client = c

	for _, v := range h.CreateValidators {
		inject.ClientInto(c, v)
	}
	for _, v := range h.UpdateValidators {
		inject.ClientInto(c, v)
	}

	return nil
}
