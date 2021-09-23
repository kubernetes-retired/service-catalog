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

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	sc "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scv "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/validation"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhookutil"
)

// StaticCreate performs basic ClusterServiceClass validation for a Create operation.
type StaticCreate struct {
}

// StaticUpdate performs basic ClusterServiceClass validation for a Create operation.
type StaticUpdate struct {
	decoder *admission.Decoder
}

var _ Validator = &StaticCreate{}
var _ Validator = &StaticUpdate{}
var _ admission.DecoderInjector = &StaticUpdate{}

// Validate validate ClusterServicePlan instance
func (v *StaticCreate) Validate(ctx context.Context, req admission.Request, clusterServiceClass *sc.ClusterServiceClass, traced *webhookutil.TracedLogger) *webhookutil.WebhookError {
	err := scv.ValidateClusterServiceClass(clusterServiceClass).ToAggregate()
	if err != nil {
		return webhookutil.NewWebhookError(err.Error(), http.StatusForbidden)
	}
	return nil
}

// Validate validate ClusterServicePlan instance
func (v *StaticUpdate) Validate(ctx context.Context, req admission.Request, clusterServiceClass *sc.ClusterServiceClass, traced *webhookutil.TracedLogger) *webhookutil.WebhookError {
	originalObj := &sc.ClusterServiceClass{}
	if err := v.decoder.DecodeRaw(req.OldObject, originalObj); err != nil {
		return webhookutil.NewWebhookError(err.Error(), http.StatusBadRequest)
	}
	err := scv.ValidateClusterServiceClassUpdate(clusterServiceClass, originalObj).ToAggregate()
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
