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

	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scv "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/validation"
	"net/http"
)

// StaticCreate runs basic ServiceBroker validation for a Create operation.
type StaticCreate struct {
}

// StaticUpdate runs basic ServiceBroker validation for an Update operation.
type StaticUpdate struct {
	decoder *admission.Decoder
}

// Validate validate ServiceBinding instance
func (v *StaticCreate) Validate(ctx context.Context, req admission.Request, serviceBroker *sc.ServiceBroker, traced *webhookutil.TracedLogger) *webhookutil.WebhookError {
	err := scv.ValidateServiceBroker(serviceBroker).ToAggregate()
	if err != nil {
		return webhookutil.NewWebhookError(err.Error(), http.StatusForbidden)
	}
	return nil
}

// Validate validate ServiceBinding instance
func (v *StaticUpdate) Validate(ctx context.Context, req admission.Request, serviceBroker *sc.ServiceBroker, traced *webhookutil.TracedLogger) *webhookutil.WebhookError {
	originalObj := &sc.ServiceBroker{}
	if err := v.decoder.DecodeRaw(req.OldObject, originalObj); err != nil {
		return webhookutil.NewWebhookError(err.Error(), http.StatusBadRequest)
	}
	err := scv.ValidateServiceBrokerUpdate(serviceBroker, originalObj).ToAggregate()
	if err != nil {
		return webhookutil.NewWebhookError(err.Error(), http.StatusForbidden)
	}
	return nil
}

// InjectDecoder injects the decoder
func (v *StaticUpdate) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
