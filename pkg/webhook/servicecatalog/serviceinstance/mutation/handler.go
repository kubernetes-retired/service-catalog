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

	sc "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scfeatures "github.com/kubernetes-sigs/service-catalog/pkg/features"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhookutil"
	admissionTypes "k8s.io/api/admission/v1beta1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// CreateUpdateHandler handles ServiceInstance
type CreateUpdateHandler struct {
	decoder            *admission.Decoder
	UUID               webhookutil.UUIDGenerator
	defaultServicePlan *DefaultServicePlan
}

// NewCreateUpdateHandler return new CreateUpdateHandler
func NewCreateUpdateHandler() *CreateUpdateHandler {
	return &CreateUpdateHandler{
		defaultServicePlan: &DefaultServicePlan{},
	}
}

var _ admission.Handler = &CreateUpdateHandler{}
var _ admission.DecoderInjector = &CreateUpdateHandler{}

// Handle handles admission requests.
func (h *CreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start handling mutation operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)

	si := &sc.ServiceInstance{}
	if err := webhookutil.MatchKinds(si, req.Kind); err != nil {
		traced.Errorf("Error matching kinds: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := h.decoder.Decode(req, si); err != nil {
		traced.Errorf("Could not decode request object: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	mutated := si.DeepCopy()
	switch req.Operation {
	case admissionTypes.Create:
		h.mutateOnCreate(ctx, req, mutated)
	case admissionTypes.Update:
		oldObj := &sc.ServiceInstance{}
		if err := h.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			traced.Errorf("Could not decode request old object: %v", err)
			return admission.Errored(http.StatusBadRequest, err)
		}
		h.mutateOnUpdate(ctx, req, oldObj, mutated)
	default:
		traced.Infof("ServiceInstance mutation wehbook does not support action %q", req.Operation)
		return admission.Allowed("action not taken")
	}

	// Sets default plan for instance if it's not specified and only one plan exists
	if err := h.defaultServicePlan.SetDefaultPlan(ctx, mutated, traced); err != nil {
		switch err.Code() {
		case http.StatusForbidden:
			return admission.Denied(err.Error())
		default:
			return admission.Errored(err.Code(), err)
		}
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

func (h *CreateUpdateHandler) mutateOnCreate(ctx context.Context, req admission.Request, instance *sc.ServiceInstance) {
	// This feature was copied from Service Catalog registry: https://github.com/kubernetes-sigs/service-catalog/blob/master/pkg/registry/servicecatalog/instance/strategy.go
	// If you want to track previous changes please check there.

	if instance.Spec.ExternalID == "" {
		instance.Spec.ExternalID = string(h.UUID.New())
	}

	if utilfeature.DefaultFeatureGate.Enabled(scfeatures.OriginatingIdentity) {
		setServiceInstanceUserInfo(req, instance)
	}

	instance.Spec.ClusterServiceClassRef = nil
	instance.Spec.ClusterServicePlanRef = nil
	instance.Finalizers = []string{sc.FinalizerServiceCatalog}
}

func (h *CreateUpdateHandler) mutateOnUpdate(ctx context.Context, req admission.Request, oldServiceInstance, newServiceInstance *sc.ServiceInstance) {

	// Clear out the ClusterServicePlanRef so that it is resolved during reconciliation
	planUpdated := newServiceInstance.Spec.ClusterServicePlanExternalName != oldServiceInstance.Spec.ClusterServicePlanExternalName ||
		newServiceInstance.Spec.ClusterServicePlanExternalID != oldServiceInstance.Spec.ClusterServicePlanExternalID ||
		newServiceInstance.Spec.ClusterServicePlanName != oldServiceInstance.Spec.ClusterServicePlanName
	if planUpdated {
		newServiceInstance.Spec.ClusterServicePlanRef = nil
	}

	// Ignore the UpdateRequests field when it is the default value
	if newServiceInstance.Spec.UpdateRequests == 0 {
		newServiceInstance.Spec.UpdateRequests = oldServiceInstance.Spec.UpdateRequests
	}

	if !apiequality.Semantic.DeepEqual(oldServiceInstance.Spec, newServiceInstance.Spec) {
		if utilfeature.DefaultFeatureGate.Enabled(scfeatures.OriginatingIdentity) {
			setServiceInstanceUserInfo(req, newServiceInstance)
		}
	}
}

// setServiceInstanceUserInfo injects user.Info from the request context
func setServiceInstanceUserInfo(req admission.Request, instance *sc.ServiceInstance) {
	user := req.UserInfo

	instance.Spec.UserInfo = &sc.UserInfo{
		Username: user.Username,
		UID:      user.UID,
		Groups:   user.Groups,
	}
	if extra := user.Extra; len(extra) > 0 {
		instance.Spec.UserInfo.Extra = map[string]sc.ExtraValue{}
		for k, v := range extra {
			instance.Spec.UserInfo.Extra[k] = sc.ExtraValue(v)
		}
	}
}

// InjectClient injects the client
func (h *CreateUpdateHandler) InjectClient(c client.Client) error {
	_, err := inject.ClientInto(c, h.defaultServicePlan)
	if err != nil {
		return err
	}
	return nil
}
