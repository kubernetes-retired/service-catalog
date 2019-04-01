// +build integration

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

package controller_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scfeatures "github.com/kubernetes-sigs/service-catalog/pkg/features"
	"github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"time"
)

// TestServiceBindingOrphanMitigation tests whether a binding has a proper status (OrphanMitigationSuccessful) after
// a bind request returns a status code that should trigger orphan mitigation.
func TestServiceBindingOrphanMitigation(t *testing.T) {
	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()
	// configure broker to respond with HTTP 500 for bind operation
	ct.SetOSBBindReactionWithHTTPError(http.StatusInternalServerError)
	require.NoError(t, ct.CreateSimpleClusterServiceBroker())
	require.NoError(t, ct.WaitForReadyBroker())
	ct.AssertClusterServiceClassAndPlan(t)
	require.NoError(t, ct.CreateServiceInstance())
	require.NoError(t, ct.WaitForReadyInstance())

	// WHEN
	assert.NoError(t, ct.CreateBinding())

	// THEN
	assert.NoError(t, ct.WaitForBindingOrphanMitigationSuccessful())
}

// TestServiceBindingFailure tests that a binding gets a failure condition when the
// broker returns a failure response for a bind operation.
func TestServiceBindingFailure(t *testing.T) {
	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()
	// configure broker to respond with HTTP 409 for bind operation
	ct.SetOSBBindReactionWithHTTPError(http.StatusConflict)
	require.NoError(t, ct.CreateSimpleClusterServiceBroker())
	require.NoError(t, ct.WaitForReadyBroker())
	ct.AssertClusterServiceClassAndPlan(t)
	require.NoError(t, ct.CreateServiceInstance())
	require.NoError(t, ct.WaitForReadyInstance())

	// WHEN
	assert.NoError(t, ct.CreateBinding())

	// THEN
	assert.NoError(t, ct.WaitForBindingFailed())
}

// TestServiceBindingRetryForNonExistingInstance try to bind to invalid service instance names.
// After the instance is created - the binding should became ready.
func TestServiceBindingRetryForNonExistingInstance(t *testing.T) {
	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()
	require.NoError(t, ct.CreateSimpleClusterServiceBroker())
	require.NoError(t, ct.WaitForReadyBroker())
	ct.AssertClusterServiceClassAndPlan(t)

	// WHEN
	// create a binding for non existing instance
	assert.NoError(t, ct.CreateBinding())
	assert.NoError(t, ct.WaitForNotReadyBinding())
	// create an instance referenced by the binding
	assert.NoError(t, ct.CreateServiceInstance())
	assert.NoError(t, ct.WaitForReadyInstance())

	// THEN
	assert.NoError(t, ct.WaitForReadyBinding())
}

// TestServiceBindingDeleteWithAsyncBindInProgress tests that you can delete a
// binding during an async bind operation.  Verify the binding is deleted when
// the bind operation completes regardless of success or failure.
func TestServiceBindingDeleteWithAsyncBindInProgress(t *testing.T) {

	for tn, state := range map[string]v2.LastOperationState{
		"binding succeeds": v2.StateSucceeded,
		"binding fails":    v2.StateFailed,
	} {
		t.Run(tn, func(t *testing.T) {
			// Enable the AsyncBindingOperations feature
			utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.AsyncBindingOperations))
			defer utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.AsyncBindingOperations))

			// GIVEN
			ct := newControllerTest(t)
			defer ct.TearDown()
			ct.EnableAsyncBind()
			ct.SetOSBPollBindingLastOperationReactionsState(v2.StateInProgress)
			require.NoError(t, ct.CreateSimpleClusterServiceBroker())
			require.NoError(t, ct.WaitForReadyBroker())
			ct.AssertClusterServiceClassAndPlan(t)
			assert.NoError(t, ct.CreateServiceInstance())
			assert.NoError(t, ct.WaitForReadyInstance())
			assert.NoError(t, ct.CreateBinding())
			assert.NoError(t, ct.WaitForBindingInProgress())

			// WHEN
			assert.NoError(t, ct.Unbind())
			// let's finish binding with a given state
			ct.SetOSBPollBindingLastOperationReactionsState(state)

			// THEN
			assert.NoError(t, ct.WaitForUnbindStatus(v1beta1.ServiceBindingUnbindStatusSucceeded))
			// at least one unbind call
			assert.NotZero(t, ct.NumberOfOSBUnbindingCalls())
		})
	}
}

// TestDeleteServiceBindingRetry tests whether deletion of a service binding
// retries after failing.
func TestDeleteServiceBindingFailureRetry(t *testing.T) {
	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()
	ct.SetFirstOSBUnbindReactionsHTTPError(1, http.StatusInternalServerError)
	require.NoError(t, ct.CreateSimpleClusterServiceBroker())
	require.NoError(t, ct.WaitForReadyBroker())
	ct.AssertClusterServiceClassAndPlan(t)
	assert.NoError(t, ct.CreateServiceInstance())
	assert.NoError(t, ct.WaitForReadyInstance())
	assert.NoError(t, ct.CreateBinding())
	assert.NoError(t, ct.WaitForReadyBinding())

	// WHEN
	assert.NoError(t, ct.Unbind())

	// THEN
	assert.NoError(t, ct.WaitForUnbindStatus(v1beta1.ServiceBindingUnbindStatusSucceeded))
	// at least two unbind calls
	assert.True(t, ct.NumberOfOSBUnbindingCalls() > 1)
}

// TestDeleteServiceBindingRetry tests whether deletion of a service binding
// retries after failing an asynchronous unbind.
func TestDeleteServiceBindingFailureRetryAsync(t *testing.T) {
	// GIVEN
	utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.AsyncBindingOperations))
	defer utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.AsyncBindingOperations))

	ct := newControllerTest(t)
	defer ct.TearDown()
	//async bind/unbind
	ct.EnableAsyncBind()
	ct.EnableAsyncUnbind()
	require.NoError(t, ct.CreateSimpleClusterServiceBroker())
	require.NoError(t, ct.WaitForReadyBroker())
	ct.AssertClusterServiceClassAndPlan(t)
	assert.NoError(t, ct.CreateServiceInstance())
	assert.NoError(t, ct.WaitForReadyInstance())
	assert.NoError(t, ct.CreateBinding())
	assert.NoError(t, ct.WaitForReadyBinding())

	// asynchronous unbinding, which is failing makes the ServiceBinding status condition constantly changing:
	//   {Type:Ready Status:Unknown  Reason:UnbindCallFailed}
	//   {Type:Ready Status:False    Reason:Unbinding
	// This is the reason we need to look at every change instead of polling the resource
	// Set up a Service Binding change listener which signals condition with reason "UnbindCallFailed"
	callFailedCh := make(chan struct{})
	ct.SetServiceBindingOnChangeListener(func(oldSb, newSb *v1beta1.ServiceBinding) {
		for _, c := range newSb.Status.Conditions {

			if c.Reason == "UnbindCallFailed" && c.Type == v1beta1.ServiceBindingConditionReady && c.Status == v1beta1.ConditionUnknown {
				callFailedCh <- struct{}{}
			}
		}
	})

	// WHEN
	// configure the broker last unbind operation to fail
	ct.SetOSBPollBindingLastOperationReactionsState(v2.StateFailed)
	assert.NoError(t, ct.Unbind())

	// THEN
	// wait for unbinding failed status
	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "waiting for ServiceBinding status condition UnbindCallFailed timed out")
	case <-callFailedCh:
	}

	// WHEN
	ct.SetOSBPollBindingLastOperationReactionsState(v2.StateSucceeded)

	// THEN
	// we expect to retry unbind with a success
	assert.NoError(t, ct.WaitForUnbindStatus(v1beta1.ServiceBindingUnbindStatusSucceeded))
	assert.True(t, ct.NumberOfOSBUnbindingCalls() > 1)

}
