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
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ReferenceDeletion handles ServiceBinding validation
type ReferenceDeletion struct {
	decoder *admission.Decoder
	client  client.Client
}

var _ admission.DecoderInjector = &ReferenceDeletion{}
var _ inject.Client = &ReferenceDeletion{}

// InjectDecoder injects the decoder
func (h *ReferenceDeletion) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// InjectClient injects the client
func (h *ReferenceDeletion) InjectClient(c client.Client) error {
	h.client = c
	return nil
}

// Validate checks if instance reference for ServiceBinding is not marked for deletion
// fail ServiceBinding operation if the ServiceInstance is marked for deletion
// This feature was copied from Service Catalog admission plugin https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.41/plugin/pkg/admission/servicebindings/lifecycle/admission.go
// If you want to track previous changes please check there.
func (h *ReferenceDeletion) Validate(ctx context.Context, req admission.Request, sb *sc.ServiceBinding, traced *webhookutil.TracedLogger) *webhookutil.WebhookError {
	instanceRef := sb.Spec.InstanceRef
	instance := &sc.ServiceInstance{}

	err := h.client.Get(ctx, types.NamespacedName{Namespace: sb.Namespace, Name: instanceRef.Name}, instance)
	if err != nil {
		traced.Infof("Could not get ServiceInstance by name %q: %v", instanceRef.Name, err)
		return nil
	}

	if instance.DeletionTimestamp != nil {
		warning := fmt.Sprintf(
			"ServiceBinding %s/%s references a ServiceInstance that is being deleted: %s/%s",
			sb.Namespace,
			sb.Name,
			sb.Namespace,
			instanceRef.Name)
		traced.Info(warning)
		return webhookutil.NewWebhookError(warning, http.StatusForbidden)
	}

	return nil
}
