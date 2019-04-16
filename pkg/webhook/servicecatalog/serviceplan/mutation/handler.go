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
	webhookutil "github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"

	admissionTypes "k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// CreateUpdateHandler handles ServicePlan
type CreateUpdateHandler struct {
	decoder *admission.Decoder
}

var _ admission.Handler = &CreateUpdateHandler{}

// Handle handles admission requests.
func (h *CreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start handling mutation operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)

	cb := &sc.ServicePlan{}
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
		oldObj := &sc.ServicePlan{}
		if err := h.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			traced.Errorf("Could not decode request old object: %v", err)
			return admission.Errored(http.StatusBadRequest, err)
		}
		h.mutateOnUpdate(ctx, oldObj, mutated)
	default:
		traced.Infof("ServicePlan mutation wehbook does not support action %q", req.Operation)
		return admission.Allowed("action not taken")
	}
	h.syncLabels(mutated)
	rawMutated, err := json.Marshal(mutated)
	if err != nil {
		traced.Errorf("Error marshaling mutated object: %v", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	traced.Infof("Completed successfully operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)
	return admission.PatchResponseFromRaw(req.Object.Raw, rawMutated)
}

var _ admission.DecoderInjector = &CreateUpdateHandler{}

// InjectDecoder injects the decoder
func (h *CreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *CreateUpdateHandler) mutateOnCreate(ctx context.Context, binding *sc.ServicePlan) {

}

func (h *CreateUpdateHandler) mutateOnUpdate(ctx context.Context, oldObj, newObj *sc.ServicePlan) {
	// This feature was copied from Service Catalog registry: https://github.com/kubernetes-incubator/service-catalog/blob/master/pkg/registry/servicecatalog/serviceplan/strategy.go
	// If you want to track previous changes please check there.

	newObj.Spec.ServiceClassRef = oldObj.Spec.ServiceClassRef
	newObj.Spec.ServiceBrokerName = oldObj.Spec.ServiceBrokerName
}

func (h *CreateUpdateHandler) syncLabels(obj *sc.ServicePlan) {
	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}

	obj.Labels[sc.GroupName+"/"+sc.FilterSpecExternalID] = obj.Spec.ExternalID
	obj.Labels[sc.GroupName+"/"+sc.FilterSpecExternalName] = obj.Spec.ExternalName
	obj.Labels[sc.GroupName+"/"+sc.FilterSpecServiceClassRefName] = obj.Spec.ServiceClassRef.Name
	obj.Labels[sc.GroupName+"/"+sc.FilterSpecServiceBrokerName] = obj.Spec.ServiceBrokerName
}
