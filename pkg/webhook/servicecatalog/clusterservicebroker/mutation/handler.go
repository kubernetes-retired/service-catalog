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
	webhookutil "github.com/kubernetes-incubator/service-catalog/pkg/webhook/util"

	admissionTypes "k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// CreateUpdateHandler handles ClusterServiceBroker
type CreateUpdateHandler struct {
	// To use the client, you need to do the following:
	// - uncomment it
	// - "sigs.k8s.io/controller-runtime/pkg/client"
	// - "sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	// - uncomment the InjectClient method at the bottom of this file.
	//client client.Client

	// Decoder decodes objects
	decoder *admission.Decoder
}

var _ admission.Handler = &CreateUpdateHandler{}

// Handle handles admission requests.
func (h *CreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start handling operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)

	cb := &sc.ClusterServiceBroker{}
	if err := webhookutil.MatchKinds(cb, req.Kind); err != nil {
		traced.Errorf("Error matching kinds: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := h.decoder.Decode(req, cb); err != nil {
		traced.Errorf("Could not decode request object: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	mutated := cb.DeepCopy()
	switch req.Operation {
	case admissionTypes.Create:
		h.mutateOnCreate(ctx, mutated)
	case admissionTypes.Update:
		h.mutateOnUpdate(ctx, mutated)
	default:
		traced.Infof("ClusterServiceBroker mutation wehbook does not support action %q", req.Operation)
		return admission.Allowed("action not taken")
	}

	rawMutated, err := json.Marshal(mutated)
	if err != nil {
		traced.Errorf("Error marshaling mutated object: %v", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	traced.Infof("Completed successfully operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)
	return admission.PatchResponseFromRaw(req.Object.Raw, rawMutated)
}

//var _ inject.Client = &CreateUpdateHandler{}
//
//// InjectClient injects the client into the CreateUpdateHandler
//func (h *CreateUpdateHandler) InjectClient(c client.Client) error {
//	h.client = c
//	return nil
//}

var _ admission.DecoderInjector = &CreateUpdateHandler{}

// InjectDecoder injects the decoder into the CreateUpdateHandler
func (h *CreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *CreateUpdateHandler) mutateOnCreate(ctx context.Context, sb *sc.ClusterServiceBroker) {
	// TODO(mszostok): logic with finalizers was moved from aggregated api-server
	// question, should we reset whole finalizers entry or only append our own?
	sb.Finalizers = []string{sc.FinalizerServiceCatalog}

	if sb.Spec.RelistBehavior == "" {
		sb.Spec.RelistBehavior = sc.ServiceBrokerRelistBehaviorDuration
	}
}

func (h *CreateUpdateHandler) mutateOnUpdate(ctx context.Context, obj *sc.ClusterServiceBroker) {
	// TODO: implement logic from pkg/registry/servicecatalog/clusterservicebroker/strategy.go
}
