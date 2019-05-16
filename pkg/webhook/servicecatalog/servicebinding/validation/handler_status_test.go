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

package validation_test

import (
	"context"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/servicebinding/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)

// TestHandlerStatusValidate tests basic cases of ServiceBindingStatus validation. All status validations tests
// are covered by pkg/apis/servicecatalog/v1beta1/validation package
func TestHandlerStatusValidate(t *testing.T) {
	tests := map[string]struct {
		givenOldRawObj  []byte
		givenNewRawObj  []byte
		expectedAllowed bool
	}{
		"Should not allow to set wrong unbind status": {
			givenOldRawObj: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBinding",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-binding",
                  "namespace": "default"
  				},
  				"spec": {
                  "externalID": "id-0123",
				  "instanceRef": {
					"name": "some-instance"
				  },
                  "secretName": "test-binding"
  				},
                "status": {
                  "conditions": [],
                  "unbindStatus": "NotRequired"
                }
			}`),
			givenNewRawObj: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBinding",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-binding",
                  "namespace": "default"
  				},
  				"spec": {
                  "externalID": "id-0123",
				  "instanceRef": {
					"name": "some-instance"
				  },
                  "secretName": "test-binding"
  				},
                "status": {
                  "conditions": [],
                  "unbindStatus": "not-allowed"
                }
			}`),
			expectedAllowed: false,
		},
		"Should not allow to set correct unbind status": {
			givenOldRawObj: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBinding",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-binding",
                  "namespace": "default"
  				},
  				"spec": {
                  "externalID": "id-0123",
				  "instanceRef": {
					"name": "some-instance"
				  },
                  "secretName": "test-binding"
  				},
                "status": {
                  "conditions": [],
                  "unbindStatus": "NotRequired"
                }
			}`),
			givenNewRawObj: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBinding",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-binding",
                  "namespace": "default"
  				},
  				"spec": {
                  "externalID": "id-0123",
				  "instanceRef": {
					"name": "some-instance"
				  },
                  "secretName": "test-binding"
  				},
                "status": {
                  "conditions": [],
                  "unbindStatus": "Succeeded"
                }
			}`),
			expectedAllowed: true,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			sc.AddToScheme(scheme.Scheme)
			decoder, err := admission.NewDecoder(scheme.Scheme)
			require.NoError(t, err)
			handler := &validation.StatusValidationHandler{}
			handler.InjectDecoder(decoder)

			req := admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Operation: admissionv1beta1.Update,
					Name:      "test-binding",
					Namespace: "system",
					Kind: metav1.GroupVersionKind{
						Kind:    "ServiceBinding",
						Version: "v1beta1",
						Group:   "servicecatalog.k8s.io",
					},
					OldObject:   runtime.RawExtension{Raw: tc.givenOldRawObj},
					Object:      runtime.RawExtension{Raw: tc.givenNewRawObj},
					SubResource: "status",
				},
			}

			// when
			resp := handler.Handle(context.Background(), req)

			// then
			assert.Equal(t, tc.expectedAllowed, resp.Allowed)
		})
	}

}
