// +build integration

/*
Copyright 2018 The Kubernetes Authors.

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

package controller_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/kubernetes-sigs/go-open-service-broker-client/v2"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProvisionInstanceWithRetries tests creating a ServiceInstance
// with retry after temporary error with/without orphan mitigation.
func TestProvisionInstanceWithRetries(t *testing.T) {
	for name, tc := range map[string]struct {
		isOrphanMitigation bool
		statusCode         int
	}{
		"With orphan mitigation": {
			isOrphanMitigation: true,
			statusCode:         http.StatusInternalServerError,
		},
		"Without orphan mitigation": {
			isOrphanMitigation: false,
			statusCode:         http.StatusUnauthorized,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ct := newControllerTest(t)
			defer ct.TearDown()
			// configure first provision response with HTTP error
			ct.SetFirstOSBProvisionReactionsHTTPError(1, tc.statusCode)
			require.NoError(t, ct.CreateSimpleClusterServiceBroker())
			require.NoError(t, ct.WaitForReadyBroker())

			// WHEN
			assert.NoError(t, ct.CreateServiceInstance())

			// THEN
			assert.NoError(t, ct.WaitForReadyInstance())
			if tc.isOrphanMitigation {
				assert.NotZero(t, ct.NumberOfOSBDeprovisionCalls())
			} else {
				assert.Zero(t, ct.NumberOfOSBDeprovisionCalls())
			}
		})
	}
}

// TestRetryAsyncDeprovision tests whether asynchronous deprovisioning retries
// by attempting a new deprovision after failing.
func TestRetryAsyncDeprovision(t *testing.T) {
	t.Parallel()

	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()
	ct.EnableAsyncInstanceDeprovisioning()
	ct.SetFirstOSBPollLastOperationReactionsFailed(1)
	require.NoError(t, ct.CreateSimpleClusterServiceBroker())
	require.NoError(t, ct.WaitForReadyBroker())
	ct.AssertClusterServiceClassAndPlan(t)

	assert.NoError(t, ct.CreateServiceInstance())
	assert.NoError(t, ct.WaitForReadyInstance())

	// WHEN
	assert.NoError(t, ct.Deprovision())

	// THEN
	assert.NoError(t, ct.WaitForDeprovisionStatus(v1beta1.ServiceInstanceDeprovisionStatusSucceeded))
	// first deprovisioning fails, expected second one
	assert.True(t, ct.NumberOfOSBDeprovisionCalls() > 1)
}

// TestServiceInstanceDeleteWithAsyncProvisionInProgress tests that you can
// delete an instance during an async provision.  Verify the instance is deleted
// when the provisioning completes regardless of success or failure.
func TestServiceInstanceDeleteWithAsyncProvisionInProgress(t *testing.T) {
	for tn, state := range map[string]v2.LastOperationState{
		"provision succeeds": v2.StateSucceeded,
		"provision fails":    v2.StateFailed,
	} {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			ct := newControllerTest(t)
			defer ct.TearDown()
			ct.EnableAsyncInstanceProvisioning()
			ct.SetOSBPollLastOperationReactionsState(v2.StateInProgress)
			require.NoError(t, ct.CreateSimpleClusterServiceBroker())
			require.NoError(t, ct.WaitForReadyBroker())
			ct.AssertClusterServiceClassAndPlan(t)
			assert.NoError(t, ct.CreateServiceInstance())
			assert.NoError(t, ct.WaitForAsyncProvisioningInProgress())

			// WHEN
			assert.NoError(t, ct.Deprovision())
			// let's finish provisioning with a given state
			ct.SetOSBPollLastOperationReactionsState(state)

			// THEN
			assert.NoError(t, ct.WaitForDeprovisionStatus(v1beta1.ServiceInstanceDeprovisionStatusSucceeded))
			// at least one deprovisioning call
			assert.NotZero(t, ct.NumberOfOSBDeprovisionCalls())
		})
	}
}

// TestServiceInstanceDeleteWithAsyncUpdateInProgress tests that you can delete
// an instance during an async update.  That is, if you request a delete during
// an instance update, the instance will be deleted when the update completes
// regardless of success or failure.
func TestServiceInstanceDeleteWithAsyncUpdateInProgress(t *testing.T) {
	for tn, state := range map[string]v2.LastOperationState{
		"update succeeds": v2.StateSucceeded,
		"update fails":    v2.StateFailed,
	} {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			ct := newControllerTest(t)
			defer ct.TearDown()
			ct.EnableAsyncInstanceUpdate()
			ct.SetOSBPollLastOperationReactionsState(v2.StateInProgress)
			require.NoError(t, ct.CreateSimpleClusterServiceBroker())
			require.NoError(t, ct.WaitForReadyBroker())
			ct.AssertClusterServiceClassAndPlan(t)
			assert.NoError(t, ct.CreateServiceInstance())
			assert.NoError(t, ct.WaitForReadyInstance())
			assert.NoError(t, ct.UpdateServiceInstanceParameters())
			assert.NoError(t, ct.WaitForInstanceUpdating())

			// WHEN
			assert.NoError(t, ct.Deprovision())
			// let's finish updating with a given state
			ct.SetOSBPollLastOperationReactionsState(state)

			// THEN
			assert.NoError(t, ct.WaitForDeprovisionStatus(v1beta1.ServiceInstanceDeprovisionStatusSucceeded))
			// at least one deprovisioning call
			assert.NotZero(t, ct.NumberOfOSBDeprovisionCalls())
		})
	}
}

// TestCreateServiceInstanceFailsWithNonexistentPlan tests creating a ServiceInstance whose ServicePlan
// does not exist
func TestCreateServiceInstanceFailsWithNonexistentPlan(t *testing.T) {
	t.Parallel()

	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()

	ct.SetupEmptyPlanListForOSBClient()
	require.NoError(t, ct.CreateSimpleClusterServiceBroker())
	require.NoError(t, ct.WaitForReadyBroker())

	// WHEN
	require.NoError(t, ct.CreateServiceInstance())

	// THEN
	condition := v1beta1.ServiceInstanceCondition{
		Type:   v1beta1.ServiceInstanceConditionReady,
		Status: v1beta1.ConditionFalse,
		Reason: "ReferencesNonexistentServicePlan",
	}
	require.NoError(t, ct.WaitForInstanceCondition(condition))
}

// TestCreateServiceInstanceNonExistentClusterServiceBroker tests creating a
// ServiceInstance whose broker does not exist.
func TestCreateServiceInstanceNonExistentClusterServiceBroker(t *testing.T) {
	t.Parallel()

	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()

	require.NoError(t, ct.CreateClusterServiceClass())
	require.NoError(t, ct.WaitForClusterServiceClass())

	require.NoError(t, ct.CreateClusterServicePlan())
	require.NoError(t, ct.WaitForClusterServicePlan())

	// WHEN
	require.NoError(t, ct.CreateServiceInstance())

	// THEN
	condition := v1beta1.ServiceInstanceCondition{
		Type:   v1beta1.ServiceInstanceConditionReady,
		Status: v1beta1.ConditionFalse,
		Reason: "ReferencesNonexistentBroker",
	}
	require.NoError(t, ct.WaitForInstanceCondition(condition))
}

// TestCreateServiceInstanceNonExistentClusterServiceClassOrPlan tests that a ServiceInstance gets
// a Failed condition when the service class or service plan it references does not exist.
func TestCreateServiceInstanceNonExistentClusterServiceClassOrPlan(t *testing.T) {
	// TODO: test should be added after merging the CRDs, for now cannot rewrite tests because fakeClient does not support `FieldSelector` during filtering ServiceInstance list
	t.Skip("Test skipped because fakeClient does not support `FieldSelector` during filtering ServiceInstance list")
}

// TestCreateServiceInstanceWithInvalidParameters tests creating a ServiceInstance
// with invalid parameters.
func TestCreateServiceInstanceWithInvalidParameters(t *testing.T) {
	t.Parallel()

	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()

	require.NoError(t, ct.CreateSimpleClusterServiceBroker())
	require.NoError(t, ct.WaitForReadyBroker())

	// WHEN
	require.NoError(t, ct.CreateServiceInstanceWithInvalidParameters())

	// THEN
	condition := v1beta1.ServiceInstanceCondition{
		Type:   v1beta1.ServiceInstanceConditionReady,
		Status: v1beta1.ConditionFalse,
		Reason: "ErrorWithParameters",
	}
	require.NoError(t, ct.WaitForInstanceCondition(condition))
}

// TestCreateServiceInstanceWithParameters tests creating a ServiceInstance
// with parameters.
func TestCreateServiceInstanceWithParameters(t *testing.T) {
	type secretDef struct {
		name string
		data map[string][]byte
	}

	for tn, state := range map[string]struct {
		params                  map[string]interface{}
		expectedParams          map[string]interface{}
		paramsFrom              []v1beta1.ParametersFromSource
		secret                  secretDef
		expectedConditionStatus v1beta1.ConditionStatus
		expectedConditionReason string
	}{
		"no params": {
			params:                  nil,
			expectedParams:          nil,
			expectedConditionStatus: v1beta1.ConditionTrue,
			expectedConditionReason: "ProvisionedSuccessfully",
		},
		"plain params": {
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
			expectedConditionStatus: v1beta1.ConditionTrue,
			expectedConditionReason: "ProvisionedSuccessfully",
		},
		"secret params": {
			expectedParams: map[string]interface{}{
				"A": "B",
				"C": map[string]interface{}{
					"D": "E",
					"F": "G",
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
			secret: secretDef{
				name: "secret-name",
				data: map[string][]byte{
					"secret-key": []byte(`{"A":"B","C":{"D":"E","F":"G"}}`),
				},
			},
			expectedConditionStatus: v1beta1.ConditionTrue,
			expectedConditionReason: "ProvisionedSuccessfully",
		},
		"plain and secret params": {
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
			secret: secretDef{
				name: "secret-name",
				data: map[string][]byte{
					"secret-key": []byte(`{"A":"B","C":{"D":"E","F":"G"}}`),
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
			expectedConditionStatus: v1beta1.ConditionTrue,
			expectedConditionReason: "ProvisionedSuccessfully",
		},
		"missing secret": {
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
			expectedConditionStatus: v1beta1.ConditionFalse,
			expectedConditionReason: "ErrorWithParameters",
		},
		"missing secret key": {
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "other-secret-key",
					},
				},
			},
			secret: secretDef{
				name: "secret-name",
				data: map[string][]byte{
					"secret-key": []byte(`bad`),
				},
			},
			expectedConditionStatus: v1beta1.ConditionFalse,
			expectedConditionReason: "ErrorWithParameters",
		},
		"empty secret data": {
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
			secret: secretDef{
				name: "secret-name",
				data: map[string][]byte{},
			},
			expectedConditionStatus: v1beta1.ConditionFalse,
			expectedConditionReason: "ErrorWithParameters",
		},
		"bad secret data": {
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
			secret: secretDef{
				name: "secret-name",
				data: map[string][]byte{
					"secret-key": []byte(`bad`),
				},
			},
			expectedConditionStatus: v1beta1.ConditionFalse,
			expectedConditionReason: "ErrorWithParameters",
		},
		"no params in secret data": {
			paramsFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
			secret: secretDef{
				name: "secret-name",
				data: map[string][]byte{
					"secret-key": []byte(`{}`),
				},
			},
			expectedParams:          nil,
			expectedConditionStatus: v1beta1.ConditionTrue,
			expectedConditionReason: "ProvisionedSuccessfully",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			ct := newControllerTest(t)
			defer ct.TearDown()

			require.NoError(t, ct.CreateSimpleClusterServiceBroker())
			require.NoError(t, ct.WaitForReadyBroker())

			// WHEN
			_, err := ct.CreateServiceInstanceWithParameters(state.params, state.paramsFrom)
			require.NoError(t, err)
			require.NoError(t, ct.CreateSecret(state.secret.name, state.secret.data))

			// THEN
			condition := v1beta1.ServiceInstanceCondition{
				Type:   v1beta1.ServiceInstanceConditionReady,
				Status: state.expectedConditionStatus,
				Reason: state.expectedConditionReason,
			}
			require.NoError(t, ct.WaitForInstanceCondition(condition))

			if state.expectedConditionStatus == v1beta1.ConditionTrue {
				ct.AssertLastBindRequest(t, state.expectedParams)
			}
		})
	}
}

// TestCreateServiceInstanceWithProvisionFailure tests creating a ServiceInstance
// with various failure results in response to the provision request.
func TestCreateServiceInstanceWithProvisionFailure(t *testing.T) {
	for tn, state := range map[string]struct {
		orphanMitigation            bool
		provisionResponseStatusCode int
		firstFailedReason           string
		orphanMitigationReason      string
		failReason                  string
		nonHTTPResponseError        error
	}{
		"Status OK": {
			orphanMitigation:            false,
			provisionResponseStatusCode: http.StatusOK,
			firstFailedReason:           "ProvisionCallFailed",
		},
		"Status Created": {
			orphanMitigation:            true,
			provisionResponseStatusCode: http.StatusCreated,
			firstFailedReason:           "StartingInstanceOrphanMitigation",
		},
		"Other 2xx": {
			orphanMitigation:            true,
			provisionResponseStatusCode: http.StatusNoContent,
			firstFailedReason:           "StartingInstanceOrphanMitigation",
		},
		"3XX": {
			orphanMitigation:            false,
			provisionResponseStatusCode: http.StatusMultipleChoices,
			firstFailedReason:           "ProvisionCallFailed",
		},
		"Status Request Timeout": {
			orphanMitigation:            false,
			provisionResponseStatusCode: http.StatusRequestTimeout,
			firstFailedReason:           "ProvisionCallFailed",
		},
		"400": {
			orphanMitigation:            false,
			provisionResponseStatusCode: http.StatusBadRequest,
			firstFailedReason:           "ProvisionCallFailed",
			failReason:                  "ClusterServiceBrokerReturnedFailure",
		},
		"Other 4XX": {
			orphanMitigation:            false,
			provisionResponseStatusCode: http.StatusForbidden,
			firstFailedReason:           "ProvisionCallFailed",
		},
		"5XX": {
			orphanMitigation:            true,
			provisionResponseStatusCode: http.StatusInternalServerError,
			firstFailedReason:           "StartingInstanceOrphanMitigation",
		},
		"Non url transport error": {
			orphanMitigation:     false,
			nonHTTPResponseError: fmt.Errorf("non-url error"),
			firstFailedReason:    "ErrorCallingProvision",
		},
		"Non timeout url error": {
			orphanMitigation: false,
			nonHTTPResponseError: &url.Error{
				Op:  "Put",
				URL: "https://fakebroker.com/v2/service_instances/instance_id",
				Err: fmt.Errorf("non-timeout error"),
			},
			firstFailedReason: "ErrorCallingProvision",
		},
		"Network timeout": {
			orphanMitigation: true,
			nonHTTPResponseError: &url.Error{
				Op:  "Put",
				URL: "https://fakebroker.com/v2/service_instances/instance_id",
				Err: TimeoutError("timeout error"),
			},
			firstFailedReason:      "StartingInstanceOrphanMitigation",
			orphanMitigationReason: "ErrorCallingProvision",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			ct := newControllerTest(t)
			defer ct.TearDown()

			require.NoError(t, ct.CreateSimpleClusterServiceBroker())
			require.NoError(t, ct.WaitForReadyBroker())

			blockDeprovisioning := make(chan bool)

			// WHEN
			if state.nonHTTPResponseError != nil {
				ct.SetCustomErrorReactionForProvisioningToOSBClient(state.nonHTTPResponseError)
			} else {
				ct.SetErrorReactionForProvisioningToOSBClient(state.provisionResponseStatusCode)
			}
			ct.SetErrorReactionForDeprovisioningToOSBClient(http.StatusInternalServerError, blockDeprovisioning)
			require.NoError(t, ct.CreateServiceInstance())

			// THEN
			// Wait for the provision to fail
			condition := v1beta1.ServiceInstanceCondition{
				Type:   v1beta1.ServiceInstanceConditionReady,
				Status: v1beta1.ConditionFalse,
				Reason: state.firstFailedReason,
			}
			require.NoError(t, ct.WaitForInstanceCondition(condition))

			// In original test next step is making sure that the latest generation has been observed
			// it means `ObservedGeneration` parameters should be equal 1, here `ObservedGeneration` is equal 0
			// because in original test apiserver is used and its functionality set `instance.Generation` to 1
			// inside file `pkg/registry/servicecatalog/instance/strategy.go` method `PrepareForCreate()`.
			// In this test the `instance.Generation` is not update so in `pkg/controller/controller_instance.go`
			// in method `reconcileServiceInstanceAdd` will not increase `ObservedGeneration` value
			// because `instance.Status.ObservedGeneration != instance.Generation` condition is not met

			// If the provision failed with a terminating failure
			if state.failReason != "" {
				condition = v1beta1.ServiceInstanceCondition{
					Type:   v1beta1.ServiceInstanceConditionFailed,
					Status: v1beta1.ConditionTrue,
					Reason: state.failReason,
				}
				require.NoError(t, ct.WaitForInstanceCondition(condition))
				assert.Zero(t, ct.NumberOfOSBDeprovisionCalls())

				return
			}

			// Assert that the orphan mitigation reason was set correctly
			if state.orphanMitigation {
				ct.AssertServiceInstanceOrphanMitigationStatus(t, true)
				blockDeprovisioning <- false

				condition = v1beta1.ServiceInstanceCondition{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionUnknown,
					Reason: "DeprovisionCallFailed",
				}
				require.NoError(t, ct.WaitForInstanceCondition(condition))
			} else {
				condition = v1beta1.ServiceInstanceCondition{
					Type:   v1beta1.ServiceInstanceConditionOrphanMitigation,
					Status: v1beta1.ConditionFalse,
				}
				ct.AssertServiceInstanceHasNoCondition(t, condition)
				ct.AssertServiceInstanceOrphanMitigationStatus(t, false)
			}

			ct.SetSuccessfullyReactionForProvisioningToOSBClient()
			ct.SetSuccessfullyReactionForDeprovisioningToOSBClient()

			// Wait for the instance to be provisioned successfully
			condition = v1beta1.ServiceInstanceCondition{
				Type:   v1beta1.ServiceInstanceConditionReady,
				Status: v1beta1.ConditionTrue,
				Reason: "ProvisionedSuccessfully",
			}
			require.NoError(t, ct.WaitForInstanceCondition(condition))

			// Assert that the observed generation is up-to-date, that orphan mitigation is not in progress,
			// and that the instance is not in a failed state.
			ct.AssertObservedGenerationIsCorrect(t)
			ct.AssertServiceInstanceOrphanMitigationStatus(t, false)
			condition = v1beta1.ServiceInstanceCondition{
				Type:   v1beta1.ServiceInstanceConditionFailed,
				Status: v1beta1.ConditionFalse,
			}
			ct.AssertServiceInstanceHasNoCondition(t, condition)
			assert.NotZero(t, ct.NumberOfOSBProvisionCalls())
		})
	}
}

// TestUpdateServiceInstanceChangePlans tests changing plans for an existing
// ServiceInstance.
func TestUpdateServiceInstanceChangePlans(t *testing.T) {
	const (
		simpleErrorUpdateReaction = "returnSimpleErrorUpdateReaction"
		errorUpdateReaction       = "returnErrorUpdateReaction"
		asyncUpdateReaction       = "returnAsyncUpdateReaction"
	)

	for tn, state := range map[string]struct {
		useExternalNames              bool
		dynamicUpdateInstanceReaction string
	}{
		"External": {
			useExternalNames: true,
		},
		"K8s": {
			useExternalNames: false,
		},
		"External name with two update call failures": {
			useExternalNames:              true,
			dynamicUpdateInstanceReaction: simpleErrorUpdateReaction,
		},
		"External name with two update failures": {
			useExternalNames:              true,
			dynamicUpdateInstanceReaction: errorUpdateReaction,
		},
		"External name update response async": {
			useExternalNames:              true,
			dynamicUpdateInstanceReaction: asyncUpdateReaction,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			ct := newControllerTest(t)
			defer ct.TearDown()

			require.NoError(t, ct.CreateSimpleClusterServiceBroker())
			require.NoError(t, ct.WaitForReadyBroker())
			require.NoError(t, ct.WaitForClusterServiceClass())
			require.NoError(t, ct.WaitForClusterServicePlan())

			require.NoError(t, ct.CreateServiceInstance())
			require.NoError(t, ct.WaitForReadyInstance())

			switch state.dynamicUpdateInstanceReaction {
			case simpleErrorUpdateReaction:
				ct.SetSimpleErrorUpdateInstanceReaction()
			case errorUpdateReaction:
				ct.SetErrorUpdateInstanceReaction()
			case asyncUpdateReaction:
				ct.AsyncForInstanceUpdate()
			}

			var (
				generation int64
				err        error
			)
			// WHEN
			if state.useExternalNames {
				generation, err = ct.UpdateServiceInstanceExternalPlanName(testOtherPlanExternalID)
			} else {
				generation, err = ct.UpdateServiceInstanceInternalPlanName(testOtherClusterServicePlanName)
			}
			require.NoError(t, err)
			require.NoError(t, ct.WaitForServiceInstanceProcessedGeneration(generation))

			// THEN
			assert.NotZero(t, ct.NumberOfOSBUpdateCalls())
			ct.AssertLastOSBUpdatePlanID(t)
		})
	}
}

// TestUpdateServiceInstanceChangePlansToNonexistentPlan tests changing plans
// to a non-existent plan.
func TestUpdateServiceInstanceChangePlansToNonexistentPlan(t *testing.T) {
	t.Parallel()

	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()

	require.NoError(t, ct.CreateSimpleClusterServiceBroker())
	require.NoError(t, ct.WaitForReadyBroker())
	require.NoError(t, ct.WaitForClusterServiceClass())
	require.NoError(t, ct.WaitForClusterServicePlan())

	require.NoError(t, ct.CreateServiceInstance())
	require.NoError(t, ct.WaitForReadyInstance())

	// WHEN
	_, err := ct.UpdateServiceInstanceExternalPlanName("non-existing-plan-id")
	require.NoError(t, err)

	// THEN
	condition := v1beta1.ServiceInstanceCondition{
		Type:   v1beta1.ServiceInstanceConditionReady,
		Status: v1beta1.ConditionFalse,
		Reason: "ReferencesNonexistentServicePlan",
	}
	require.NoError(t, ct.WaitForInstanceCondition(condition))
}

// TestUpdateServiceInstanceNewDashboardResponse tests setting Dashboard URL when
// and update Instance request returns a new DashboardURL.
// CAUTION: the test cannot run parallel because it changes global flag which can include on working other tests
func TestUpdateServiceInstanceNewDashboardResponse(t *testing.T) {
	for tn, state := range map[string]struct {
		enableFeatureGate bool
	}{
		"Alpha features enabled": {
			enableFeatureGate: true,
		},
		"Alpha feature disabled": {
			enableFeatureGate: false,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			ct := newControllerTest(t)
			defer ct.TearDown()

			require.NoError(t, ct.SetFeatureGateDashboardURL(state.enableFeatureGate))
			// default value for defaultFuture is false
			// see https://github.com/kubernetes/apiserver/blob/release-1.14/pkg/util/feature/feature_gate.go
			defer require.NoError(t, ct.SetFeatureGateDashboardURL(false))

			require.NoError(t, ct.CreateSimpleClusterServiceBroker())
			require.NoError(t, ct.CreateServiceInstance())
			require.NoError(t, ct.WaitForReadyInstance())
			ct.SetUpdateServiceInstanceResponseWithDashboardURL()

			// WHEN
			require.NoError(t, ct.UpdateServiceInstanceParameters())
			require.NoError(t, ct.WaitForReadyUpdateInstance())

			// THEN
			if state.enableFeatureGate {
				ct.AssertServiceInstanceDashboardURL(t)
			} else {
				ct.AssertServiceInstanceEmptyDashboardURL(t)
			}
		})
	}
}

// TestUpdateServiceInstanceUpdateParameters tests updating the parameters
// of an existing ServiceInstance.
func TestUpdateServiceInstanceUpdateParameters(t *testing.T) {
	for tn, state := range map[string]struct {
		instanceWithParams           bool
		instanceWithParamsFromSecret bool
		updateParams                 bool
		updateParamsFromSecret       bool
		deleteParams                 bool
		deleteParamsFromSecret       bool
		expectedParams               map[string]interface{}
	}{
		"Add param": {
			updateParams:   true,
			expectedParams: map[string]interface{}{"param-key": "new-param-value"},
		},
		"Update param": {
			instanceWithParams: true,
			updateParams:       true,
			expectedParams:     map[string]interface{}{"param-key": "new-param-value"},
		},
		"Delete param": {
			instanceWithParams: true,
			deleteParams:       true,
			expectedParams:     map[string]interface{}{},
		},
		"Add param with secret": {
			instanceWithParamsFromSecret: true,
			updateParams:                 true,
			expectedParams: map[string]interface{}{
				"secret-param-key": "secret-param-value",
				"param-key":        "new-param-value"},
		},
		"Update param with secret": {
			instanceWithParams:           true,
			instanceWithParamsFromSecret: true,
			updateParams:                 true,
			expectedParams: map[string]interface{}{
				"param-key":        "new-param-value",
				"secret-param-key": "secret-param-value"},
		},
		"Delete param with secret": {
			instanceWithParams:           true,
			instanceWithParamsFromSecret: true,
			deleteParams:                 true,
			expectedParams:               map[string]interface{}{"secret-param-key": "secret-param-value"},
		},
		"Add secret param": {
			instanceWithParamsFromSecret: true,
			updateParamsFromSecret:       true,
			expectedParams:               map[string]interface{}{"other-secret-param-key": "other-secret-param-value"},
		},
		"Update secret param": {
			instanceWithParamsFromSecret: true,
			updateParamsFromSecret:       true,
			expectedParams:               map[string]interface{}{"other-secret-param-key": "other-secret-param-value"},
		},
		"Delete secret param": {
			instanceWithParamsFromSecret: true,
			deleteParamsFromSecret:       true,
			expectedParams:               map[string]interface{}{},
		},
		"Add secret param with plain param": {
			instanceWithParams:     true,
			updateParamsFromSecret: true,
			expectedParams: map[string]interface{}{
				"param-key":              "param-value",
				"other-secret-param-key": "other-secret-param-value"},
		},
		"Update secret param with plain param": {
			instanceWithParams:           true,
			instanceWithParamsFromSecret: true,
			updateParamsFromSecret:       true,
			expectedParams: map[string]interface{}{
				"param-key":              "param-value",
				"other-secret-param-key": "other-secret-param-value"},
		},
		"Delete secret param with plain param": {
			instanceWithParams:           true,
			instanceWithParamsFromSecret: true,
			deleteParamsFromSecret:       true,
			expectedParams:               map[string]interface{}{"param-key": "param-value"},
		},
		"Update secret": {
			instanceWithParamsFromSecret: true,
			updateParamsFromSecret:       true,
			expectedParams:               map[string]interface{}{"other-secret-param-key": "other-secret-param-value"},
		},
		"Update secret with plain param": {
			instanceWithParams:           true,
			instanceWithParamsFromSecret: true,
			updateParamsFromSecret:       true,
			expectedParams: map[string]interface{}{
				"param-key":              "param-value",
				"other-secret-param-key": "other-secret-param-value"},
		},
		"Add plain and secret param": {
			updateParams:           true,
			updateParamsFromSecret: true,
			expectedParams: map[string]interface{}{
				"param-key":              "new-param-value",
				"other-secret-param-key": "other-secret-param-value"},
		},
		"Update plain and secret param": {
			instanceWithParams:           true,
			instanceWithParamsFromSecret: true,
			updateParams:                 true,
			updateParamsFromSecret:       true,
			expectedParams: map[string]interface{}{
				"param-key":              "new-param-value",
				"other-secret-param-key": "other-secret-param-value"},
		},
		"Delete plain and secret param": {
			instanceWithParams:           true,
			instanceWithParamsFromSecret: true,
			deleteParams:                 true,
			deleteParamsFromSecret:       true,
			expectedParams:               map[string]interface{}{},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			ct := newControllerTest(t)
			defer ct.TearDown()

			require.NoError(t, ct.CreateSimpleClusterServiceBroker())
			require.NoError(t, ct.WaitForReadyBroker())

			require.NoError(t, ct.CreateSecretsForServiceInstanceWithSecretParams())

			require.NoError(t, ct.CreateServiceInstanceWithCustomParameters(
				state.instanceWithParams,
				state.instanceWithParamsFromSecret))
			require.NoError(t, ct.WaitForReadyInstance())

			// WHEN
			generation, err := ct.UpdateCustomServiceInstanceParameters(
				state.updateParams,
				state.updateParamsFromSecret,
				state.deleteParams,
				state.deleteParamsFromSecret)
			require.NoError(t, err)
			require.NoError(t, ct.WaitForServiceInstanceProcessedGeneration(generation))

			// THEN
			assert.NotZero(t, ct.NumberOfOSBUpdateCalls())
			ct.AssertBrokerUpdateActionWithParametersExist(t, state.expectedParams)
		})
	}
}
