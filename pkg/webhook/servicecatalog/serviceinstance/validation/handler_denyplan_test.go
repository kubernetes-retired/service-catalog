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
	sc "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/serviceinstance/validation"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhookutil/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)

func TestSpecValidationHandlerDenyPlanChangeIfNotUpdatableSimpleScenarios(t *testing.T) {
	tester.DiscardLoggedMsg()

	// given
	clusterServiceClassName := "csc-test"

	err := sc.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UID:       "uuid",
			Name:      "test-serviceinstance",
			Namespace: "ns-test",
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceInstance",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{Raw: []byte(`{
 				"metadata": {
 				  "name": "test-serviceinstance"
 				},
 				"spec": {
                  "clusterServiceClassRef": {
 					 "name": "` + clusterServiceClassName + `"
                  }
 				}
			}`)},
		},
	}
	sch, err := sc.SchemeBuilderRuntime.Build()
	require.NoError(t, err)

	decoder, err := admission.NewDecoder(sch)
	require.NoError(t, err)

	tests := map[string]struct {
		operation               admissionv1beta1.Operation
		serviceClassName        string
		serviceClassIsUpdatable bool
		responseAllowed         bool
		responseReason          string
	}{
		"UpdateablePlan set to false, no changes": {
			admissionv1beta1.Update,
			clusterServiceClassName,
			false,
			true,
			"ServiceInstance validation successful",
		},
		"UpdateablePlan set to true": {
			admissionv1beta1.Update,
			clusterServiceClassName,
			true,
			true,
			"ServiceInstance validation successful",
		},
		"Non-existing service class": {
			admissionv1beta1.Update,
			"NonExistingServiceClassName",
			true,
			false,
			"clusterserviceclasses.servicecatalog.k8s.io \"" + clusterServiceClassName + "\" not found",
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// given
			handler := validation.SpecValidationHandler{}
			handler.UpdateValidators = []validation.Validator{&validation.DenyPlanChangeIfNotUpdatable{}}
			fakeClient := fake.NewFakeClientWithScheme(sch, &sc.ClusterServiceClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      test.serviceClassName,
					Namespace: "",
				},
				Spec: sc.ClusterServiceClassSpec{
					CommonServiceClassSpec: sc.CommonServiceClassSpec{
						PlanUpdatable: test.serviceClassIsUpdatable,
					},
				},
			})
			err := handler.InjectDecoder(decoder)
			require.NoError(t, err)
			err = handler.InjectClient(fakeClient)
			require.NoError(t, err)
			request.AdmissionRequest.Operation = test.operation

			// when
			response := handler.Handle(context.Background(), request)

			// then
			assert.Equal(t, response.AdmissionResponse.Allowed, test.responseAllowed)
			assert.Contains(t, response.AdmissionResponse.Result.Reason, test.responseReason)
		})
	}
}

func TestSpecValidationHandlerDenyPlanChangeIfNotUpdatablePlanNameChanged(t *testing.T) {
	tester.DiscardLoggedMsg()

	// given
	clusterServiceClassName := "csc-test"

	err := sc.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UID:       "uuid",
			Name:      "test-serviceinstance",
			Namespace: "ns-test",
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceInstance",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{Raw: []byte(`{
 				"metadata": {
 				  "name": "test-serviceinstance"
 				},
 				"spec": {
                  "clusterServicePlanName": "micro",
                  "clusterServiceClassRef": {
 					 "name": "` + clusterServiceClassName + `"
                  }
 				}
			}`)},
			OldObject: runtime.RawExtension{Raw: []byte(`{
 				"metadata": {
 				  "name": "test-serviceinstance"
 				},
 				"spec": {
                  "clusterServicePlanName": "enterprise",
                  "clusterServiceClassRef": {
 					 "name": "` + clusterServiceClassName + `"
                  }
 				}
			}`)},
		},
	}
	sch, err := sc.SchemeBuilderRuntime.Build()
	require.NoError(t, err)

	decoder, err := admission.NewDecoder(sch)
	require.NoError(t, err)

	tests := map[string]struct {
		serviceClassName        string
		serviceClassIsUpdatable bool
		responseAllowed         bool
		responseReason          string
	}{
		"UpdateablePlan set to false, plan changed": {
			clusterServiceClassName,
			false,
			false,
			"The Service Class " + clusterServiceClassName + " does not allow plan changes.",
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// given
			handler := validation.SpecValidationHandler{}
			handler.UpdateValidators = []validation.Validator{&validation.DenyPlanChangeIfNotUpdatable{}}
			fakeClient := fake.NewFakeClientWithScheme(sch, &sc.ClusterServiceClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      test.serviceClassName,
					Namespace: "",
				},
				Spec: sc.ClusterServiceClassSpec{
					CommonServiceClassSpec: sc.CommonServiceClassSpec{
						PlanUpdatable: test.serviceClassIsUpdatable,
					},
				},
			})
			err := handler.InjectDecoder(decoder)
			require.NoError(t, err)
			err = handler.InjectClient(fakeClient)
			require.NoError(t, err)
			request.AdmissionRequest.Operation = admissionv1beta1.Update

			// when
			response := handler.Handle(context.Background(), request)

			// then
			assert.Equal(t, response.AdmissionResponse.Allowed, test.responseAllowed)
			assert.Contains(t, response.AdmissionResponse.Result.Reason, test.responseReason)
		})
	}
}
