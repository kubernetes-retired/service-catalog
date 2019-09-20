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
	"github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/servicebinding/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
	"time"
)

const (
	UpToDateInstance  = "up-to-date-instance"
	OutOfDateInstance = "out-of-date-instance"
)

func TestSpecValidationHandlerServiceInstanceReferenceUpToDate(t *testing.T) {
	// given
	namespace := "test-handler"
	err := sc.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UID:       "1111-aaaa",
			Name:      "test-binding",
			Namespace: namespace,
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceBinding",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{Raw: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBinding",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-binding"
  				},
  				"spec": {
				  "instanceRef": {
					"name": "` + UpToDateInstance + `"
				  },
				  "externalID": "123-abc",
				  "secretName": "test-binding"
  				}
			}`)},
		},
	}

	sch, err := sc.SchemeBuilderRuntime.Build()
	require.NoError(t, err)

	decoder, err := admission.NewDecoder(sch)
	require.NoError(t, err)

	tests := map[string]struct {
		operation admissionv1beta1.Operation
	}{
		"Request for Create ServiceBinding should be allowed": {
			admissionv1beta1.Create,
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// given
			handler := validation.SpecValidationHandler{}
			handler.CreateValidators = []validation.Validator{&validation.ReferenceDeletion{}}

			fakeClient := fake.NewFakeClientWithScheme(sch, &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      UpToDateInstance,
					Namespace: namespace,
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
			assert.True(t, response.AdmissionResponse.Allowed)
		})
	}
}

func TestSpecValidationHandlerServiceInstanceReferenceOutOfDate(t *testing.T) {
	// given
	namespace := "test-handler"
	err := sc.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UID:       "2222-bbbb",
			Name:      "test-binding",
			Namespace: namespace,
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceBinding",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{Raw: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBinding",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-binding"
  				},
  				"spec": {
				  "instanceRef": {
					"name": "` + OutOfDateInstance + `"
				  },
				  "externalID": "123-abc",
				  "secretName": "test-binding"
  				}
			}`)},
		},
	}

	sch, err := sc.SchemeBuilderRuntime.Build()
	require.NoError(t, err)

	decoder, err := admission.NewDecoder(sch)
	require.NoError(t, err)

	tests := map[string]struct {
		operation admissionv1beta1.Operation
	}{
		"Request for Create ServiceBinding should be denied": {
			admissionv1beta1.Create,
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// given
			handler := validation.SpecValidationHandler{}
			handler.CreateValidators = []validation.Validator{&validation.ReferenceDeletion{}}

			fakeClient := fake.NewFakeClientWithScheme(sch, &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:              OutOfDateInstance,
					Namespace:         namespace,
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
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
			assert.False(t, response.AdmissionResponse.Allowed)
		})
	}
}
