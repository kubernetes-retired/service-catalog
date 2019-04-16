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
	"testing"

	"github.com/appscode/jsonpatch"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/clusterservicebroker/mutation"
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
  				"kind": "ClusterServiceBroker",
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
  				"kind": "ClusterServiceBroker",
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
						Kind:    "ClusterServiceBroker",
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

func TestCreateUpdateHandlerHandleUpdateSuccess(t *testing.T) {
	tests := map[string]struct {
		oldRawObject []byte
		newRawObj    []byte

		expPatches []jsonpatch.Operation
	}{
		"Should restore previous relist request, when not provided (set to 0)": {
			oldRawObject: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServiceBroker",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-broker",
                  "generation": 1
  				},
  				"spec": {
				  "relistRequests": 1,
				  "relistBehavior": "Duration",	
  				  "url": "http://localhost:8081/"
  				}
			}`),
			newRawObj: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServiceBroker",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-broker",
                  "generation": 1
  				},
  				"spec": {
				  "relistRequests": 0,
				  "relistBehavior": "Duration",
  				  "url": "http://localhost:8081/"
  				}
			}`),
			expPatches: []jsonpatch.Operation{
				{
					Operation: "replace",
					Path:      "/spec/relistRequests",
					Value:     float64(1),
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
					Operation: admissionv1beta1.Update,
					Name:      "test-broker",
					Namespace: "system",
					Kind: metav1.GroupVersionKind{
						Kind:    "ClusterServiceBroker",
						Version: "v1beta1",
						Group:   "servicecatalog.k8s.io",
					},
					OldObject: runtime.RawExtension{Raw: tc.oldRawObject},
					Object:    runtime.RawExtension{Raw: tc.newRawObj},
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

func TestCreateUpdateHandlerHandleDecoderErrors(t *testing.T) {
	tester.DiscardLoggedMsg()

	for _, fn := range []func(t *testing.T, handler tester.TestDecoderHandler, kind string){
		tester.AssertHandlerReturnErrorIfReqObjIsMalformed,
		tester.AssertHandlerReturnErrorIfGVKMismatch,
	} {
		handler := mutation.CreateUpdateHandler{}
		fn(t, &handler, "ClusterServiceBroker")
	}
}
