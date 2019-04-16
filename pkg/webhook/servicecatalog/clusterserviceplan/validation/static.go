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

// StaticCreate performs basic ClusterServicePlan validation for a Create operation.
type StaticCreate struct {
}

// StaticUpdate performs basic ClusterServicePlan validation for an Update operation.
type StaticUpdate struct {
	decoder *admission.Decoder
}

var _ Validator = &StaticCreate{}
var _ Validator = &StaticUpdate{}
var _ admission.DecoderInjector = &StaticUpdate{}

// Validate validates ClusterServicePlan instance
func (v *StaticCreate) Validate(ctx context.Context, req admission.Request, clusterServicePlan *sc.ClusterServicePlan, traced *webhookutil.TracedLogger) *webhookutil.WebhookError {
	err := scv.ValidateClusterServicePlan(clusterServicePlan).ToAggregate()
	if err != nil {
		return webhookutil.NewWebhookError(err.Error(), http.StatusForbidden)
	}
	return nil
}

// Validate validates ClusterServicePlan instance
func (v *StaticUpdate) Validate(ctx context.Context, req admission.Request, clusterServicePlan *sc.ClusterServicePlan, traced *webhookutil.TracedLogger) *webhookutil.WebhookError {
	originalObj := &sc.ClusterServicePlan{}
	if err := v.decoder.DecodeRaw(req.OldObject, originalObj); err != nil {
		return webhookutil.NewWebhookError(err.Error(), http.StatusBadRequest)
	}
	err := scv.ValidateClusterServicePlanUpdate(clusterServicePlan, originalObj).ToAggregate()
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
