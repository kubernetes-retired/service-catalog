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
	sc "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/clusterserviceplan/mutation"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhookutil/tester"
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
		"Should copy spec fields to labels": {
			givenRawObj: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServicePlan",
  				"metadata": {
  				  "name": "test-service-plan"
  				},
  				"spec": {
                  "externalID": "id",
                  "externalName": "name",
                  "clusterServiceBrokerName": "broker",
                  "clusterServiceClassRef": {"name": "refbroker"}
  				}
			}`),
			expPatches: []jsonpatch.Operation{
				{
					Operation: "add",
					Path:      "/metadata/labels",
					Value: map[string]interface{}{
						sc.GroupName + "/" + sc.FilterSpecExternalID:                 "id",
						sc.GroupName + "/" + sc.FilterSpecExternalName:               "name",
						sc.GroupName + "/" + sc.FilterSpecClusterServiceBrokerName:   "broker",
						sc.GroupName + "/" + sc.FilterSpecClusterServiceClassRefName: "refbroker",
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
					Name:      "test-service-plan",
					Namespace: "system",
					Kind: metav1.GroupVersionKind{
						Kind:    "ClusterServicePlan",
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

			for _, expPatch := range tc.expPatches {
				assert.Contains(t, resp.Patches, expPatch)
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
		"Should reset broker name and class ref to old one and allow to change other fields": {
			oldRawObject: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServicePlan",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-plan",
                  "labels": {
                    "servicecatalog.k8s.io/spec.externalName":"external-name",
					"servicecatalog.k8s.io/spec.clusterServiceBrokerName":"test-broker",
                    "servicecatalog.k8s.io/spec.externalID": "external-id",
                    "servicecatalog.k8s.io/spec.clusterServiceClassRef.name":"external-class"
                  }
  				},
  				"spec": {
                  "clusterServiceBrokerName": "test-broker",
                  "clusterServiceClassRef": {"name":"external-class"},
                  "externalName": "external-name",
                  "externalID": "external-id",
				  "description":"a description",
				  "bindable": false,
                  "free": false
  				}
			}`),
			newRawObj: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServicePlan",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-plan",
				  "labels": {
                    "servicecatalog.k8s.io/spec.externalName":"external-name",
					"servicecatalog.k8s.io/spec.clusterServiceBrokerName":"test-broker",
                    "servicecatalog.k8s.io/spec.externalID": "external-id",
                    "servicecatalog.k8s.io/spec.clusterServiceClassRef.name":"external-class"
                  }
  				},
  				"spec": {
                  "clusterServiceBrokerName": "test-broker-changed",
                  "clusterServiceClassRef": {"name":"external-class-changed"},
                  "externalName": "external-name",
                  "externalID": "external-id",
				  "description":"a description",
				  "bindable": true,
                  "free": true
  				}
			}`),
			expPatches: []jsonpatch.Operation{
				{
					Operation: "replace",
					Path:      "/spec/clusterServiceBrokerName",
					Value:     "test-broker",
				},
				{
					Operation: "replace",
					Path:      "/spec/clusterServiceClassRef/name",
					Value:     "external-class",
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
					Name:      "test-plan",
					Namespace: "system",
					Kind: metav1.GroupVersionKind{
						Kind:    "ClusterServicePlan",
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
		fn(t, &handler, "ClusterServicePlan")
	}
}
