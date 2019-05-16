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
	scfeatures "github.com/kubernetes-incubator/service-catalog/pkg/features"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"

	admissionTypes "k8s.io/api/admission/v1beta1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// CreateUpdateHandler handles ServiceBinding
type CreateUpdateHandler struct {
	decoder *admission.Decoder
	UUID    webhookutil.UUIDGenerator
}

var _ admission.Handler = &CreateUpdateHandler{}

// Handle handles admission requests.
func (h *CreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start handling mutation operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)

	sb := &sc.ServiceBinding{}
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
		h.mutateOnCreate(ctx, req, mutated)
	case admissionTypes.Update:
		oldObj := &sc.ServiceBinding{}
		if err := h.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			traced.Errorf("Could not decode request old object: %v", err)
			return admission.Errored(http.StatusBadRequest, err)
		}
		h.mutateOnUpdate(ctx, req, oldObj, mutated)
	default:
		traced.Infof("ServiceBinding mutation wehbook does not support action %q", req.Operation)
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

var _ admission.DecoderInjector = &CreateUpdateHandler{}

// InjectDecoder injects the decoder
func (h *CreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *CreateUpdateHandler) mutateOnCreate(ctx context.Context, req admission.Request, binding *sc.ServiceBinding) {
	// This feature was copied from Service Catalog registry: https://github.com/kubernetes-incubator/service-catalog/blob/master/pkg/registry/servicecatalog/binding/strategy.go
	// If you want to track previous changes please check there.

	if binding.Spec.ExternalID == "" {
		binding.Spec.ExternalID = string(h.UUID.New())
	}

	if binding.Spec.SecretName == "" {
		binding.Spec.SecretName = binding.Name
	}

	if utilfeature.DefaultFeatureGate.Enabled(scfeatures.OriginatingIdentity) {
		setServiceBindingUserInfo(req, binding)
	}

	binding.Finalizers = []string{sc.FinalizerServiceCatalog}
}

func (h *CreateUpdateHandler) mutateOnUpdate(ctx context.Context, req admission.Request, oldServiceBinding, newServiceBinding *sc.ServiceBinding) {
	// TODO: We currently don't handle any changes to the spec in the
	// reconciler. Once we do that, this check needs to be removed and
	// proper validation of allowed changes needs to be implemented in
	// ValidateUpdate. Also, the check for whether the generation needs
	// to be updated needs to be un-commented.
	// If the Spec change is allowed do not forget to update UserInfo (setServiceBindingUserInfo function)
	newServiceBinding.Spec = oldServiceBinding.Spec
}

// setServiceBindingUserInfo injects user.Info from the request context
func setServiceBindingUserInfo(req admission.Request, binding *sc.ServiceBinding) {
	user := req.UserInfo

	binding.Spec.UserInfo = &sc.UserInfo{
		Username: user.Username,
		UID:      user.UID,
		Groups:   user.Groups,
	}
	if extra := user.Extra; len(extra) > 0 {
		binding.Spec.UserInfo.Extra = map[string]sc.ExtraValue{}
		for k, v := range extra {
			binding.Spec.UserInfo.Extra[k] = sc.ExtraValue(v)
		}
	}
}
