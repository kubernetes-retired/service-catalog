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
	"testing"

	// avoid error `servicecatalog/v1beta1 is not enabled`
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	fakeosb "github.com/pmorie/go-open-service-broker-client/v2/fake"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/test/util"
)

// TestCreateServiceInstanceNonExistentClusterServiceClassOrPlan tests that a ServiceInstance gets
// a Failed condition when the service class or service plan it references does not exist.
func TestCreateServiceBinding(t *testing.T) {
	cases := []struct {
		name                           string
		serviceInstanceName            *string
		serviceInstanceAsyncInProgress bool
		asyncForInstances              bool
		asyncForBindings               bool
		expectedErrorReason            string
		expectedFailure                bool
	}{
		{
			name: "happy",
		},
		{
			name:                "non-existent service instance name",
			serviceInstanceName: strPtr("nothereinstance"),
			expectedErrorReason: "ReferencesNonexistentInstance",
		},
		{
			name:                "invalid service instance name",
			serviceInstanceName: strPtr(""),
			expectedErrorReason: "ReferencesNonexistentInstance",
			expectedFailure:     true,
		},
		{
			name: "bind to async-in-progress service instance",
			serviceInstanceAsyncInProgress: true,
			asyncForInstances:              true,
			expectedErrorReason:            "ErrorAsyncOperationInProgress",
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
					i.Status.AsyncOpInProgress = tc.serviceInstanceAsyncInProgress
					return i
				}(),
				binding: func() *v1beta1.ServiceBinding {
					b := getTestBinding()
					if tc.serviceInstanceName != nil {
						b.Spec.ServiceInstanceRef.Name = *tc.serviceInstanceName
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
