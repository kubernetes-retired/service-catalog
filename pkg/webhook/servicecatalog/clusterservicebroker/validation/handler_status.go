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
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scv "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/validation"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"
	admissionTypes "k8s.io/api/admission/v1beta1"

	"context"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// StatusUpdateHandler provides status update resource validation handler
type StatusUpdateHandler struct {
	decoder *admission.Decoder
}

// Handle handles admission requests.
func (h *StatusUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start handling validation operation: %s for %s/%s: %q", req.Operation, req.Kind.Kind, req.SubResource, req.Name)

	if req.Operation != admissionTypes.Update {
		traced.Infof("Operation %s is not validated", req.Operation)
		return admission.Allowed("status operation allowed")
	}

	newSb := &sc.ClusterServiceBroker{}
	if err := h.decoder.Decode(req, newSb); err != nil {
		traced.Errorf("Could not decode request object: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}
	oldSb := &sc.ClusterServiceBroker{}
	if err := h.decoder.DecodeRaw(req.OldObject, oldSb); err != nil {
		traced.Errorf("Could not decode request object: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	eList := scv.ValidateClusterServiceBrokerStatusUpdate(newSb, oldSb)

	if err := eList.ToAggregate(); err != nil {
		traced.Infof("%s/%s update not allowed: %s", req.Kind.Kind, req.SubResource, err.Error())
		return admission.Denied(err.Error())
	}

	traced.Infof("Completed successfully validation operation: %s for %s/%s: %q", req.Operation, req.Kind.Kind, req.SubResource, req.Name)
	return admission.Allowed("status update allowed")
}

// InjectDecoder injects the decoder into the handlers
func (h *StatusUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}
