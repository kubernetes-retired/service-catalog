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

package mutation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/appscode/jsonpatch"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/servicebroker/mutation"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestCreateUpdateHandlerHandleCreateSuccess(t *testing.T) {
	tests := map[string]struct {
		givenRawObj []byte

		expPatches []jsonpatch.Operation
	}{
		"Should set all default fields": {
			givenRawObj: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBroker",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "relistRequests": 1,
  				  "url": "http://localhost:8081/"
  				}
			}`),
			expPatches: []jsonpatch.Operation{
				{
					Operation: "add",
					Path:      "/metadata/finalizers",
					Value: []interface{}{
						"kubernetes-incubator/service-catalog",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/relistBehavior",
					Value:     "Duration",
				},
			},
		},
		"Should omit relistBehavior if it's already set": {
			givenRawObj: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBroker",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "relistRequests": 1,
				  "relistBehavior": "Manual",
  				  "url": "http://localhost:8081/"
  				}
			}`),
			expPatches: []jsonpatch.Operation{
				{
					Operation: "add",
					Path:      "/metadata/finalizers",
					Value: []interface{}{
						"kubernetes-incubator/service-catalog",
					},
				},
			},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			sc.AddToScheme(scheme.Scheme)
			decoder, err := admission.NewDecoder(scheme.Scheme)
			require.NoError(t, err)

			fixReq := admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Operation: admissionv1beta1.Create,
					Name:      "test-broker",
					Namespace: "system",
					Kind: metav1.GroupVersionKind{
						Kind:    "ServiceBroker",
						Version: "v1beta1",
						Group:   "servicecatalog.k8s.io",
					},
					Object: runtime.RawExtension{Raw: tc.givenRawObj},
				},
			}

			handler := mutation.CreateUpdateHandler{}
			handler.InjectDecoder(decoder)

			// when
			resp := handler.Handle(context.Background(), fixReq)

			// then
			assert.True(t, resp.Allowed)
			require.NotNil(t, resp.PatchType)
			assert.Equal(t, admissionv1beta1.PatchTypeJSONPatch, *resp.PatchType)

			// filtering out status cause k8s api-server will discard this too
			patches := tester.FilterOutStatusPatch(resp.Patches)

			require.Len(t, patches, len(tc.expPatches))
			for _, expPatch := range tc.expPatches {
				assert.Contains(t, patches, expPatch)
			}
		})
	}
}

func TestCreateUpdateHandlerHandleReturnErrorIfGVKMismatch(t *testing.T) {
	// given
	sc.AddToScheme(scheme.Scheme)
	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.NoError(t, err)

	fixReq := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			Operation: admissionv1beta1.Create,
			Name:      "test-broker",
			Namespace: "system",
			Kind: metav1.GroupVersionKind{
				Kind:    "ClusterServiceBroker",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
		},
	}

	expReqResult := &metav1.Status{
		Code:    http.StatusBadRequest,
		Message: "type mismatch: want: servicecatalog.k8s.io/v1beta1, Kind=ServiceBroker got: servicecatalog.k8s.io/v1beta1, Kind=ClusterServiceBroker",
	}

	handler := mutation.CreateUpdateHandler{}
	handler.InjectDecoder(decoder)

	// when
	resp := handler.Handle(context.Background(), fixReq)

	// then
	assert.False(t, resp.Allowed)
	assert.Equal(t, expReqResult, resp.Result)
}

func TestCreateUpdateHandlerHandleReturnErrorIfReqObjIsMalformed(t *testing.T) {
	// given
	sc.AddToScheme(scheme.Scheme)
	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.NoError(t, err)

	fixReq := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			Operation: admissionv1beta1.Create,
			Name:      "test-broker",
			Namespace: "system",
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceBroker",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{Raw: []byte("{malformed: JSON,,")},
		},
	}

	expReqResult := &metav1.Status{
		Code:    http.StatusBadRequest,
		Message: "couldn't get version/kind; json parse error: invalid character 'm' looking for beginning of object key string",
	}

	handler := mutation.CreateUpdateHandler{}
	handler.InjectDecoder(decoder)

	// when
	resp := handler.Handle(context.Background(), fixReq)

	// then
	assert.False(t, resp.Allowed)
	assert.Equal(t, expReqResult, resp.Result)
}
