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
	"fmt"
	"net/http"
	"testing"

	"github.com/appscode/jsonpatch"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scfeatures "github.com/kubernetes-incubator/service-catalog/pkg/features"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/serviceinstance/mutation"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestCreateUpdateHandlerHandleCreateSuccess(t *testing.T) {
	const fixUUID = "mocked-uuid-123-abc"
	tests := map[string]struct {
		givenRawObj []byte
		expPatches  []jsonpatch.Operation
	}{
		"Should set all default fields": {
			givenRawObj: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceInstance",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-instance"
  				},
  				"spec": {
				  "updateRequests": 1,
				  "clusterServiceClassExternalName": "some-class",
				  "clusterServicePlanExternalName": "some-plan" 
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
					Path:      "/spec/externalID",
					Value:     fixUUID,
				},
			},
		},
		"Should omit externalID and secretName if they are already set": {
			givenRawObj: []byte(`{
				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceInstance",
  				"metadata": {
  				  "creationTimestamp": null,
  				  "name": "test-instance"
  				},
  				"spec": {
				  "updateRequests": 1,
				  "clusterServiceClassExternalName": "some-class",
				  "clusterServicePlanExternalName": "some-plan",
				  "externalID": "my-external-id-123"
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

			err = utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.OriginatingIdentity))
			require.NoError(t, err, "cannot disable OriginatingIdentity feature")
			// restore default state
			defer utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.OriginatingIdentity))

			fixReq := admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Operation: admissionv1beta1.Create,
					Name:      "test-instance",
					Namespace: "system",
					Kind: metav1.GroupVersionKind{
						Kind:    "ServiceInstance",
						Version: "v1beta1",
						Group:   "servicecatalog.k8s.io",
					},
					Object: runtime.RawExtension{Raw: tc.givenRawObj},
				},
			}

			handler := mutation.CreateUpdateHandler{
				UUID: func() types.UID { return fixUUID },
			}
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

func TestCreateUpdateHandlerHandleSetUserInfoIfOriginatingIdentityIsEnabled(t *testing.T) {
	// given
	sc.AddToScheme(scheme.Scheme)
	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.NoError(t, err)

	// assumption that OriginatingIdentity feature is enabled by default

	reqUserInfo := authenticationv1.UserInfo{
		Username: "minikube",
		UID:      "123",
		Groups:   []string{"unauthorized"},
		Extra: map[string]authenticationv1.ExtraValue{
			"extra": {"val1", "val2"},
		},
	}

	fixReq := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			Operation: admissionv1beta1.Create,
			Name:      "test-instance",
			Namespace: "system",
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceInstance",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			UserInfo: reqUserInfo,
			Object: runtime.RawExtension{Raw: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceInstance",
  				"metadata": {
				  "finalizers": [ "kubernetes-incubator/service-catalog" ],
  				  "creationTimestamp": null,
  				  "name": "test-instance"
  				},
  				"spec": {
				  "updateRequests": 1,
				  "clusterServiceClassExternalName": "some-class",
				  "clusterServicePlanExternalName": "some-plan",
				  "externalID": "my-external-id-123"
  				}
			}`)},
		},
	}

	expPatches := []jsonpatch.Operation{
		{
			Operation: "add",
			Path:      "/spec/userInfo",
			Value: map[string]interface{}{
				"username": "minikube",
				"uid":      "123",
				"groups": []interface{}{
					"unauthorized",
				},
				"extra": map[string]interface{}{
					"extra": []interface{}{
						"val1", "val2",
					},
				},
			},
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

	require.Len(t, patches, len(expPatches))
	for _, expPatch := range expPatches {
		assert.Contains(t, patches, expPatch)
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
			Name:      "test-instance",
			Namespace: "system",
			Kind: metav1.GroupVersionKind{
				Kind:    "ClusterServiceClass",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
		},
	}

	expReqResult := &metav1.Status{
		Code:    http.StatusBadRequest,
		Message: "type mismatch: want: servicecatalog.k8s.io/v1beta1, Kind=ServiceInstance got: servicecatalog.k8s.io/v1beta1, Kind=ClusterServiceClass",
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
			Name:      "test-instance",
			Namespace: "system",
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceInstance",
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
