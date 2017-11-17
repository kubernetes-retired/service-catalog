/*
Copyright 2017 The Kubernetes Authors.

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

package integration

import (
	"reflect"
	"testing"

	// avoid error `servicecatalog/v1beta1 is not enabled`
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	fakeosb "github.com/pmorie/go-open-service-broker-client/v2/fake"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/test/util"
	"net/http"
)

// TestCreateServiceBindingSuccess successful paths binding
func TestCreateServiceBindingSuccess(t *testing.T) {
	cases := []struct {
		name string
	}{
		{
			name: "defaults",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ct := &controllerTest{
				t:        t,
				broker:   getTestBroker(),
				instance: getTestInstance(),
				binding:  getTestBinding(),
			}
			ct.run(func(ct *controllerTest) {
				{
					condition := v1beta1.ServiceBindingCondition{
						Type:   v1beta1.ServiceBindingConditionReady,
						Status: v1beta1.ConditionTrue,
					}
					if cond, err := util.WaitForBindingConditionLastSeenOfType(ct.client, testNamespace, testBindingName, condition); err != nil {
						t.Fatalf("error waiting for binding condition: %v\n"+"expecting: %+v\n"+"last seen: %+v", err, condition, cond)
					}
				}
			})
		})
	}
}

// TestCreateServiceBindingInvalidInstance try to bind to invalid service instance names
func TestCreateServiceBindingInvalidInstance(t *testing.T) {
	cases := []struct {
		name                string
		instanceName        *string
		expectedErrorReason string
		expectedFailure     bool
	}{
		{
			name:                "non-existent service instance name",
			instanceName:        strPtr("nothereinstance"),
			expectedErrorReason: "ReferencesNonexistentInstance",
		},
		{
			name:                "invalid service instance name",
			instanceName:        strPtr(""),
			expectedErrorReason: "ReferencesNonexistentInstance",
			expectedFailure:     true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ct := &controllerTest{
				t: t,
				skipBindingCreateError: tc.expectedFailure,
				broker:                 getTestBroker(),
				instance:               getTestInstance(),
				binding: func() *v1beta1.ServiceBinding {
					b := getTestBinding()
					if tc.instanceName != nil {
						b.Spec.ServiceInstanceRef.Name = *tc.instanceName
					}
					return b
				}(),
				skipVerifyingBindingSuccess: tc.expectedErrorReason != "",
			}
			ct.run(func(ct *controllerTest) {
				{
					status := v1beta1.ConditionTrue
					if tc.expectedErrorReason != "" {
						status = v1beta1.ConditionFalse
					}
					condition := v1beta1.ServiceBindingCondition{
						Type:   v1beta1.ServiceBindingConditionReady,
						Status: status,
						Reason: tc.expectedErrorReason,
					}
					if ct.skipBindingCreateError {
						if err := util.WaitForBindingToNotExist(ct.client, testNamespace, testBindingName); err != nil {
							t.Fatalf("error waiting for binding to not exist: %v", err)
						}
					} else {
						if cond, err := util.WaitForBindingConditionLastSeenOfType(ct.client, testNamespace, testBindingName, condition); err != nil {
							t.Fatalf("error waiting for binding condition: %v\n"+"expecting: %+v\n"+"last seen: %+v", err, condition, cond)
						}
					}

				}
			})
		})
	}
}

// TODO: this one is still broken:
// E1116 16:10:21.295296    8302 controller_instance.go:1844] ServiceInstance "test-namespace/test-instance": Failed to update status: ServiceInstance.servicecatalog.k8s.io "test-instance" is invalid: status.currentOperation: Forbidden: currentOperation must not be present when reconciledGeneration and generation are equal
// TestCreateServiceBindingAsyncServiceInstance try to bind to a in progress service instance.
//func TestCreateServiceBindingAsync(t *testing.T) {
//	cases := []struct {
//		name                string
//		asyncForInstances   bool
//		expectedErrorReason string
//	}{
//		{
//			name:                "bind to async-in-progress service instance",
//			asyncForInstances:   true,
//			expectedErrorReason: "ErrorAsyncOperationInProgress",
//		},
//	}
//	for _, tc := range cases {
//		t.Run(tc.name, func(t *testing.T) {
//			ct := &controllerTest{
//				t:        t,
//				broker:   getTestBroker(),
//				instance: getTestInstance(),
//				binding:  getTestBinding(),
//				setup: func(ct *controllerTest) {
//					if tc.asyncForInstances {
//						ct.osbClient.ProvisionReaction.(*fakeosb.ProvisionReaction).Response.Async = true
//						ct.osbClient.UpdateInstanceReaction.(*fakeosb.UpdateInstanceReaction).Response.Async = true
//						ct.osbClient.DeprovisionReaction.(*fakeosb.DeprovisionReaction).Response.Async = true
//
//						key := osb.OperationKey(testInstanceLastOperation)
//
//						ct.osbClient.PollLastOperationReactions = map[osb.OperationKey]*fakeosb.PollLastOperationReaction{
//							key: {
//								Response: &osb.LastOperationResponse{
//									State:       osb.StateInProgress,
//									Description: strPtr("StateInProgress"),
//								},
//							},
//						}
//
//						ct.osbClient.ProvisionReaction.(*fakeosb.ProvisionReaction).Response.OperationKey = &key
//
//						ct.skipVerifyingInstanceSuccess = true
//					}
//
//				},
//				preDeleteInstance: func(ct *controllerTest) {
//					// Let the instance finish.
//					ct.osbClient.PollLastOperationReactions = map[osb.OperationKey]*fakeosb.PollLastOperationReaction{}
//				},
//				skipVerifyingBindingSuccess: tc.expectedErrorReason != "",
//			}
//			ct.run(func(ct *controllerTest) {
//				{
//					status := v1beta1.ConditionTrue
//					if tc.expectedErrorReason != "" {
//						status = v1beta1.ConditionFalse
//					}
//					condition := v1beta1.ServiceBindingCondition{
//						Type:   v1beta1.ServiceBindingConditionReady,
//						Status: status,
//						Reason: tc.expectedErrorReason,
//					}
//					if ct.skipBindingCreateError {
//						if err := util.WaitForBindingToNotExist(ct.client, testNamespace, testBindingName); err != nil {
//							t.Fatalf("error waiting for binding to not exist: %v", err)
//						}
//					} else {
//						if cond, err := util.WaitForBindingConditionLastSeenOfType(ct.client, testNamespace, testBindingName, condition); err != nil {
//							t.Fatalf("error waiting for binding condition: %v\n"+"expecting: %+v\n"+"last seen: %+v", err, condition, cond)
//						}
//					}
//
//				}
//			})
//		})
//	}
//}

// TestCreateServiceBindingNonBindable bind to a non-bindable service class / plan.
func TestCreateServiceBindingNonBindable(t *testing.T) {
	cases := []struct {
		name                string
		expectedErrorReason string
		nonbindablePlan     bool
	}{
		{
			name:                "non-bindable plan",
			nonbindablePlan:     true,
			expectedErrorReason: "ErrorNonbindableServiceClass",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ct := &controllerTest{
				t:      t,
				broker: getTestBroker(),
				instance: func() *v1beta1.ServiceInstance {
					i := getTestInstance()
					if tc.nonbindablePlan {
						i.Spec.PlanReference.ClusterServicePlanExternalName = testNonbindableClusterServicePlanName
					}
					return i
				}(),
				binding:                     getTestBinding(),
				skipVerifyingBindingSuccess: tc.expectedErrorReason != "",
			}
			ct.run(func(ct *controllerTest) {
				{
					status := v1beta1.ConditionTrue
					if tc.expectedErrorReason != "" {
						status = v1beta1.ConditionFalse
					}
					condition := v1beta1.ServiceBindingCondition{
						Type:   v1beta1.ServiceBindingConditionReady,
						Status: status,
						Reason: tc.expectedErrorReason,
					}
					if ct.skipBindingCreateError {
						if err := util.WaitForBindingToNotExist(ct.client, testNamespace, testBindingName); err != nil {
							t.Fatalf("error waiting for binding to not exist: %v", err)
						}
					} else {
						if cond, err := util.WaitForBindingConditionLastSeenOfType(ct.client, testNamespace, testBindingName, condition); err != nil {
							t.Fatalf("error waiting for binding condition: %v\n"+"expecting: %+v\n"+"last seen: %+v", err, condition, cond)
						}
					}

				}
			})
		})
	}
}

// TestCreateServiceBindingInstanceNotReady bind to a service instance in the ready false state.
func TestCreateServiceBindingInstanceNotReady(t *testing.T) {
	cases := []struct {
		name                string
		instanceNotReady    bool
		expectedErrorReason string
	}{
		{
			name:                "service instance not ready",
			instanceNotReady:    true,
			expectedErrorReason: "ErrorInstanceNotReady",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ct := &controllerTest{
				t:        t,
				broker:   getTestBroker(),
				instance: getTestInstance(),
				binding:  getTestBinding(),
				setup: func(ct *controllerTest) {
					if tc.instanceNotReady {
						reactionError := osb.HTTPStatusCodeError{
							StatusCode:   http.StatusBadGateway,
							ErrorMessage: strPtr("error message"),
							Description:  strPtr("response description"),
						}
						ct.osbClient.ProvisionReaction = &fakeosb.ProvisionReaction{
							Error: reactionError,
						}
						ct.skipVerifyingInstanceSuccess = true
					}
				},
				skipVerifyingBindingSuccess: tc.expectedErrorReason != "",
			}
			ct.run(func(ct *controllerTest) {
				{
					status := v1beta1.ConditionTrue
					if tc.expectedErrorReason != "" {
						status = v1beta1.ConditionFalse
					}
					condition := v1beta1.ServiceBindingCondition{
						Type:   v1beta1.ServiceBindingConditionReady,
						Status: status,
						Reason: tc.expectedErrorReason,
					}
					if ct.skipBindingCreateError {
						if err := util.WaitForBindingToNotExist(ct.client, testNamespace, testBindingName); err != nil {
							t.Fatalf("error waiting for binding to not exist: %v", err)
						}
					} else {
						if cond, err := util.WaitForBindingConditionLastSeenOfType(ct.client, testNamespace, testBindingName, condition); err != nil {
							t.Fatalf("error waiting for binding condition: %v\n"+"expecting: %+v\n"+"last seen: %+v", err, condition, cond)
						}
					}

				}
			})
		})
	}
}

// TestCreateServiceBindingWithParameters tests creating a ServiceBinding
// with parameters.
func TestCreateServiceBindingWithParameters(t *testing.T) {
	type secretDef struct {
		name string
		data map[string][]byte
	}
	cases := []struct {
		name           string
		params         map[string]interface{}
		paramsFrom     []v1beta1.ParametersFromSource
		secrets        []secretDef
		expectedParams map[string]interface{}
		expectedError  bool
	}{
		{
			name:           "no params",
			expectedParams: nil,
		},
		{
			name: "plain params",
			params: map[string]interface{}{
				"Name": "test-param",
				"Args": map[string]interface{}{
					"first":  "first-arg",
					"second": "second-arg",
				},
			},
			expectedParams: map[string]interface{}{
				"Name": "test-param",
				"Args": map[string]interface{}{
					"first":  "first-arg",
					"second": "second-arg",
				},
			},
		},
		{
			name: "secret params",
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
			secrets: []secretDef{
				{
					name: "secret-name",
					data: map[string][]byte{
						"secret-key": []byte(`{"A":"B","C":{"D":"E","F":"G"}}`),
					},
				},
			},
			expectedParams: map[string]interface{}{
				"A": "B",
				"C": map[string]interface{}{
					"D": "E",
					"F": "G",
				},
			},
		},
		{
			name: "plain and secret params",
			params: map[string]interface{}{
				"Name": "test-param",
				"Args": map[string]interface{}{
					"first":  "first-arg",
					"second": "second-arg",
				},
			},
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
			secrets: []secretDef{
				{
					name: "secret-name",
					data: map[string][]byte{
						"secret-key": []byte(`{"A":"B","C":{"D":"E","F":"G"}}`),
					},
				},
			},
			expectedParams: map[string]interface{}{
				"Name": "test-param",
				"Args": map[string]interface{}{
					"first":  "first-arg",
					"second": "second-arg",
				},
				"A": "B",
				"C": map[string]interface{}{
					"D": "E",
					"F": "G",
				},
			},
		},
		{
			name: "missing secret",
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "missing secret key",
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "other-secret-key",
					},
				},
			},
			secrets: []secretDef{
				{
					name: "secret-name",
					data: map[string][]byte{
						"secret-key": []byte(`bad`),
					},
				},
			},
			expectedError: true,
		},
		{
			name: "empty secret data",
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
			secrets: []secretDef{
				{
					name: "secret-name",
					data: map[string][]byte{},
				},
			},
			expectedError: true,
		},
		{
			name: "bad secret data",
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
			secrets: []secretDef{
				{
					name: "secret-name",
					data: map[string][]byte{
						"secret-key": []byte(`bad`),
					},
				},
			},
			expectedError: true,
		},
		{
			name: "no params in secret data",
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
			secrets: []secretDef{
				{
					name: "secret-name",
					data: map[string][]byte{
						"secret-key": []byte(`{}`),
					},
				},
			},
			expectedParams: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ct := &controllerTest{
				t:        t,
				broker:   getTestBroker(),
				instance: getTestInstance(),
				binding: func() *v1beta1.ServiceBinding {
					b := getTestBinding()
					if tc.params != nil {
						b.Spec.Parameters = convertParametersIntoRawExtension(t, tc.params)
					}
					b.Spec.ParametersFrom = tc.paramsFrom
					return b
				}(),
				skipVerifyingBindingSuccess: tc.expectedError,
				setup: func(ct *controllerTest) {
					for _, secret := range tc.secrets {
						prependGetSecretReaction(ct.kubeClient, secret.name, secret.data)
					}
				},
			}
			ct.run(func(ct *controllerTest) {
				if tc.expectedError {
					if err := util.WaitForBindingCondition(ct.client, testNamespace, testBindingName, v1beta1.ServiceBindingCondition{
						Type:   v1beta1.ServiceBindingConditionReady,
						Status: v1beta1.ConditionFalse,
						Reason: "ErrorWithParameters",
					}); err != nil {
						t.Fatalf("error waiting for binding reconciliation to fail: %v", err)
					}
				} else {
					brokerAction := getLastBrokerAction(t, ct.osbClient, fakeosb.Bind)
					if e, a := tc.expectedParams, brokerAction.Request.(*osb.BindRequest).Parameters; !reflect.DeepEqual(e, a) {
						t.Fatalf("unexpected diff in provision parameters: expected %v, got %v", e, a)
					}
				}
			})
		})
	}
}
