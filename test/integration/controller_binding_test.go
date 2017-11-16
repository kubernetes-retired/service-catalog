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
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
)

// TestCreateServiceInstanceNonExistentClusterServiceClassOrPlan tests that a ServiceInstance gets
// a Failed condition when the service class or service plan it references does not exist.
func TestCreateServiceBinding(t *testing.T) {
	cases := []struct {
		name                string
		instanceName        *string
		instanceClassName   *string
		instancePlanName    *string
		instanceNotReady    bool
		asyncForInstances   bool
		asyncForBindings    bool
		expectedErrorReason string
		expectedFailure     bool
		nonbindablePlan     bool
		duplicateParameters bool
	}{
		{
			name: "happy",
		},
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
		{
			name:                "bind to async-in-progress service instance",
			asyncForInstances:   true,
			expectedErrorReason: "ErrorAsyncOperationInProgress",
		},
		// TODO: Can't seem to get the test in this state.
		//{
		//	name: "unresolved ClusterServiceClass",
		//},
		// TODO: Can't seem to get the test in this state.
		//{
		//	name: "unresolved ClusterServicePlan",
		//},
		{
			name:                "non-bindable plan",
			nonbindablePlan:     true,
			expectedErrorReason: "ErrorNonbindableServiceClass",
		},
		{
			name:                "service instance not ready",
			instanceNotReady:    true,
			expectedErrorReason: "ErrorInstanceNotReady",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ct := &controllerTest{
				t: t,
				skipBindingCreateError: tc.expectedFailure,
				broker:                 getTestBroker(),
				instance: func() *v1beta1.ServiceInstance {
					i := getTestInstance()
					if tc.instanceClassName != nil {
						i.Spec.PlanReference.ClusterServiceClassExternalName = *tc.instanceClassName
					}
					if tc.instancePlanName != nil {
						i.Spec.PlanReference.ClusterServicePlanExternalName = *tc.instancePlanName
					}
					if tc.nonbindablePlan {
						i.Spec.PlanReference.ClusterServicePlanExternalName = testNonbindableClusterServicePlanName
					}
					return i
				}(),
				binding: func() *v1beta1.ServiceBinding {
					b := getTestBinding()
					if tc.instanceName != nil {
						b.Spec.ServiceInstanceRef.Name = *tc.instanceName
					}
					if tc.duplicateParameters {
						b.Spec.Parameters = &runtime.RawExtension{Raw: []byte(`{"a":"1","a":"2"}`)}
					}
					return b
				}(),
				setup: func(ct *controllerTest) {
					if tc.asyncForInstances {
						ct.osbClient.ProvisionReaction.(*fakeosb.ProvisionReaction).Response.Async = true
						ct.osbClient.UpdateInstanceReaction.(*fakeosb.UpdateInstanceReaction).Response.Async = true
						ct.osbClient.DeprovisionReaction.(*fakeosb.DeprovisionReaction).Response.Async = true

						ct.osbClient.PollLastOperationReactions = map[osb.OperationKey]*fakeosb.PollLastOperationReaction{
							testInstanceLastOperation: {
								Response: &osb.LastOperationResponse{
									State:       osb.StateInProgress,
									Description: strPtr("StateInProgress"),
								},
							},
						}

						ct.skipVerifyingInstanceSuccess = true
					}

					if tc.asyncForBindings {
						ct.osbClient.BindReaction.(*fakeosb.BindReaction).Response.Async = true
						ct.osbClient.UnbindReaction.(*fakeosb.UnbindReaction).Response.Async = true
					}

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
