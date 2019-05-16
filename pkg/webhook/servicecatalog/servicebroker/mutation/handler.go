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

package mutation

import (
	"context"
	"encoding/json"
	"net/http"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"

	admissionTypes "k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// CreateUpdateHandler handles ServiceBroker
type CreateUpdateHandler struct {
	decoder *admission.Decoder
}

var _ admission.Handler = &CreateUpdateHandler{}
var _ admission.DecoderInjector = &CreateUpdateHandler{}

// Handle handles admission requests.
func (h *CreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start handling mutation operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)

	sb := &sc.ServiceBroker{}
	if err := webhookutil.MatchKinds(sb, req.Kind); err != nil {
		traced.Errorf("Error matching kinds: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := h.decoder.Decode(req, sb); err != nil {
		traced.Errorf("Could not decode request object: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	mutated := sb.DeepCopy()
	switch req.Operation {
	case admissionTypes.Create:
		h.mutateOnCreate(ctx, mutated)
	case admissionTypes.Update:
		oldObj := &sc.ServiceBroker{}
		if err := h.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			traced.Errorf("Could not decode request old object: %v", err)
			return admission.Errored(http.StatusBadRequest, err)
		}
		h.mutateOnUpdate(ctx, oldObj, mutated)
	default:
		traced.Infof("ServiceBroker mutation wehbook does not support action %q", req.Operation)
		return admission.Allowed("action not taken")
	}

	rawMutated, err := json.Marshal(mutated)
	if err != nil {
		traced.Errorf("Error marshaling mutated object: %v", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	traced.Infof("Completed successfully mutation operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)
	return admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, rawMutated)
}

// InjectDecoder injects the decoder
func (h *CreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *CreateUpdateHandler) mutateOnCreate(ctx context.Context, sb *sc.ServiceBroker) {
	sb.Finalizers = []string{sc.FinalizerServiceCatalog}

	if sb.Spec.RelistBehavior == "" {
		sb.Spec.RelistBehavior = sc.ServiceBrokerRelistBehaviorDuration
	}
}

func (h *CreateUpdateHandler) mutateOnUpdate(ctx context.Context, oldSb, newSb *sc.ServiceBroker) {
	// This feature was copied from Service Catalog registry: https://github.com/kubernetes-incubator/service-catalog/blob/master/pkg/registry/servicecatalog/servicebroker/strategy.go
	// If you want to track previous changes please check there.

	if newSb.Spec.RelistRequests == 0 {
		newSb.Spec.RelistRequests = oldSb.Spec.RelistRequests
	}
}
