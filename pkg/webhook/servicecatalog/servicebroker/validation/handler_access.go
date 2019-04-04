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
	"fmt"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"
	authenticationapi "k8s.io/api/authentication/v1"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// AccessToBroker handles ServiceBroker validation
type AccessToBroker struct {
	decoder *admission.Decoder
	client  client.Client
}

var _ admission.DecoderInjector = &AccessToBroker{}
var _ inject.Client = &AccessToBroker{}

// InjectDecoder injects the decoder
func (h *AccessToBroker) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// InjectClient injects the client
func (h *AccessToBroker) InjectClient(c client.Client) error {
	h.client = c
	return nil
}

// Validate checks if client has access to service broker if broker requires authentication
// This feature was copied from Service Catalog admission plugin https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.41/plugin/pkg/admission/broker/authsarcheck/admission.go
// If you want to track previous changes please check there.
func (h *AccessToBroker) Validate(ctx context.Context, req admission.Request, sb *sc.ServiceBroker, traced *webhookutil.TracedLogger) *webhookutil.WebhookError {
	if sb.Spec.AuthInfo == nil {
		traced.Infof("%s %q has no AuthInfo. Operation completed", sb.Kind, sb.Name)
		return nil
	}

	var secretRef *sc.LocalObjectReference
	if sb.Spec.AuthInfo.Basic != nil {
		secretRef = sb.Spec.AuthInfo.Basic.SecretRef
	} else if sb.Spec.AuthInfo.Bearer != nil {
		secretRef = sb.Spec.AuthInfo.Bearer.SecretRef
	}

	if secretRef == nil {
		traced.Infof("%s %q has no SecretRef neither in Basic nor Bearer auth. Operation completed", sb.Kind, sb.Name)
		return nil
	}

	user := req.UserInfo
	sar := &authorizationapi.SubjectAccessReview{
		Spec: authorizationapi.SubjectAccessReviewSpec{
			ResourceAttributes: &authorizationapi.ResourceAttributes{
				Namespace: sb.Namespace,
				Verb:      "get",
				Group:     corev1.SchemeGroupVersion.Group,
				Version:   corev1.SchemeGroupVersion.Version,
				Resource:  corev1.ResourceSecrets.String(),
				Name:      secretRef.Name,
			},
			User:   user.Username,
			Groups: user.Groups,
			Extra:  convertToSARExtra(user.Extra),
			UID:    user.UID,
		},
	}

	err := h.client.Create(ctx, sar)
	if err != nil {
		traced.Errorf("Could not create SubjectAccessReview for %s %q: %v", sb.Kind, sb.Name, err)
		return webhookutil.NewWebhookError(err.Error(), http.StatusForbidden)
	}

	if !sar.Status.Allowed {
		msg := fmt.Sprintf(
			"broker forbidden access to auth secret (%s): Reason: %s, EvaluationError: %s",
			secretRef.Name,
			sar.Status.Reason,
			sar.Status.EvaluationError)
		traced.Info(msg)
		return webhookutil.NewWebhookError(msg, http.StatusForbidden)
	}

	return nil
}

func convertToSARExtra(extra map[string]authenticationapi.ExtraValue) map[string]authorizationapi.ExtraValue {
	if extra == nil {
		return nil
	}

	ret := map[string]authorizationapi.ExtraValue{}
	for k, v := range extra {
		ret[k] = authorizationapi.ExtraValue(v)
	}

	return ret
}
