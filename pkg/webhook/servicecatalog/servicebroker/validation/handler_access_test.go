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
	"errors"
	sc "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/servicebroker/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)

const (
	AllowedSecretName = "csb-secret-name"
	DeniedSecretName  = "denied-csb-secret-name"
)

// Reactors are not implemented in 'sigs.k8s.io/controller-runtime/pkg/client/fake' package
// https://github.com/kubernetes-sigs/controller-runtime/issues/72
// instead it is used custom client with override Create method
type fakedClient struct {
	client.Client
}

// Create overrides real client Create method for the test
func (m *fakedClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOptionFunc) error {
	if _, ok := obj.(*v1.SubjectAccessReview); !ok {
		return errors.New("Input object is not SubjectAccessReview type")
	}

	if obj.(*v1.SubjectAccessReview).Spec.ResourceAttributes.Name == AllowedSecretName {
		obj.(*v1.SubjectAccessReview).Status.Allowed = true
	}

	return nil
}

func TestSpecValidationHandlerAccessToBrokerAllowed(t *testing.T) {
	// given
	err := sc.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UID:       "5555-eeee",
			Name:      "test-broker",
			Namespace: "test-handler",
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceBroker",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{},
		},
	}

	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.NoError(t, err)

	tests := map[string]struct {
		operation admissionv1beta1.Operation
		object    []byte
	}{
		"Request for Create ServiceBroker without AuthInfo should be allowed": {
			admissionv1beta1.Create,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local"
  				}
			}`),
		},
		"Request for Update ServiceBroker without AuthInfo should be allowed": {
			admissionv1beta1.Update,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local"
  				}
			}`),
		},
		"Request for Create ServiceBroker with AuthInfo should be allowed": {
			admissionv1beta1.Create,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local",
				  "authInfo": {
    			    "basic": {
      				  "secretRef": {
        			    "namespace": "test-handler",
						"name": "` + AllowedSecretName + `"
					  }
					}
				  }
  				}
			}`),
		},
		"Request for Update ServiceBroker with AuthInfo should be allowed": {
			admissionv1beta1.Update,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local",
				  "authInfo": {
    			    "bearer": {
      				  "secretRef": {
        			    "namespace": "test-handler",
						"name": "` + AllowedSecretName + `"
					  }
					}
				  }
				}
			}`),
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// given
			handler := validation.SpecValidationHandler{}
			handler.CreateValidators = []validation.Validator{&validation.AccessToBroker{}}
			handler.UpdateValidators = []validation.Validator{&validation.AccessToBroker{}}

			err := handler.InjectDecoder(decoder)
			require.NoError(t, err)
			err = handler.InjectClient(&fakedClient{})
			require.NoError(t, err)

			request.AdmissionRequest.Operation = test.operation
			request.AdmissionRequest.Object.Raw = test.object

			// when
			response := handler.Handle(context.Background(), request)

			// then
			assert.True(t, response.AdmissionResponse.Allowed)
		})
	}
}

func TestSpecValidationHandlerAccessToBrokerDenied(t *testing.T) {
	// given
	err := sc.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UID:       "6666-ffff",
			Name:      "test-broker",
			Namespace: "test-handler",
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceBroker",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{},
		},
	}

	sch := scheme.Scheme

	decoder, err := admission.NewDecoder(sch)
	require.NoError(t, err)

	tests := map[string]struct {
		operation admissionv1beta1.Operation
		object    []byte
	}{
		"Request for Create ServiceBroker should be denied": {
			admissionv1beta1.Create,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local",
				  "authInfo": {
    			    "bearer": {
      				  "secretRef": {
        			    "namespace": "test-handler",
						"name": "` + DeniedSecretName + `"
					  }
					}
				  }
  				}
			}`),
		},
		"Request for Update ServiceBroker should be denied": {
			admissionv1beta1.Update,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local",
				  "authInfo": {
    			    "basic": {
      				  "secretRef": {
        			    "namespace": "test-handler",
						"name": "` + DeniedSecretName + `"
					  }
					}
				  }
  				}
			}`),
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// given
			handler := validation.SpecValidationHandler{}
			handler.CreateValidators = []validation.Validator{&validation.AccessToBroker{}}
			handler.UpdateValidators = []validation.Validator{&validation.AccessToBroker{}}

			fakeClient := fake.NewFakeClientWithScheme(sch, &sc.ServiceBroker{})

			err := handler.InjectDecoder(decoder)
			require.NoError(t, err)
			err = handler.InjectClient(fakeClient)
			require.NoError(t, err)

			request.AdmissionRequest.Operation = test.operation
			request.AdmissionRequest.Object.Raw = test.object

			// when
			response := handler.Handle(context.Background(), request)

			// then
			assert.False(t, response.AdmissionResponse.Allowed)
		})
	}
}
