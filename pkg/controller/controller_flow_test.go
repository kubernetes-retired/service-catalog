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
	"testing"

	scfeatures "github.com/kubernetes-sigs/service-catalog/pkg/features"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
)

// TestBasicFlowWithBasicAuth tests whether the controller uses correct credentials when the secret changes
func TestBasicFlowWithBasicAuth(t *testing.T) {
	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()
	// create a secret with basic auth credentials stored
	require.NoError(t, ct.CreateSecretWithBasicAuth("user1", "p2sswd"))

	// WHEN
	assert.NoError(t, ct.CreateClusterServiceBrokerWithBasicAuth())
	assert.NoError(t, ct.WaitForReadyBroker())

	// THEN
	ct.AssertOSBBasicAuth(t, "user1", "p2sswd")
	ct.AssertClusterServiceClassAndPlan(t)

	// WHEN
	assert.NoError(t, ct.UpdateSecretWithBasicAuth("user1", "newp2sswd"))
	assert.NoError(t, ct.CreateServiceInstance())

	// THEN
	assert.NoError(t, ct.WaitForReadyInstance())
	// expected at least one provision call
	assert.NotZero(t, ct.NumberOfOSBProvisionCalls())

	// expected: new credentials must be used
	ct.AssertOSBBasicAuth(t, "user1", "newp2sswd")
}

// TestBasicFlow tests
//
// - add Broker
// - verify ClusterServiceClasses added
// - provision Instance
// - update Instance
// - make Binding
func TestBasicFlow(t *testing.T) {
	for tn, setupFunc := range map[string]func(ts *controllerTest){
		"sync": func(ts *controllerTest) {
		},
		"async instances with multiple polls": func(ct *controllerTest) {
			ct.AsyncForInstances()
			ct.SetFirstOSBPollLastOperationReactionsInProgress(2)
		},
		"async bindings": func(ct *controllerTest) {
			ct.AsyncForBindings()
		},
		"async instances and bindings": func(ct *controllerTest) {
			ct.AsyncForInstances()
			ct.AsyncForBindings()
		},
	} {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()
			// GIVEN
			utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.AsyncBindingOperations))
			defer utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.AsyncBindingOperations))
			ct := newControllerTest(t)
			defer ct.TearDown()
			setupFunc(ct)

			// WHEN
			require.NoError(t, ct.CreateSimpleClusterServiceBroker())
			// THEN
			assert.NoError(t, ct.WaitForReadyBroker())
			ct.AssertClusterServiceClassAndPlan(t)

			// WHEN
			require.NoError(t, ct.CreateServiceInstance())

			// THEN
			assert.NoError(t, ct.WaitForReadyInstance())
			// expected at least one provision call
			assert.NotZero(t, ct.NumberOfOSBProvisionCalls())

			// WHEN
			assert.NoError(t, ct.CreateBinding())

			// THEN
			assert.NoError(t, ct.WaitForReadyBinding())
			// expected at least one binding call
			assert.NotZero(t, ct.NumberOfOSBBindingCalls())
		})
	}
}

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
	require.NoError(t, ct.CreateServiceInstance())
	require.NoError(t, ct.WaitForReadyInstance())

	// WHEN
	ct.CreateBinding()

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
// After the instance is created - the binding shoul became ready.
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

// TestProvisionInstanceWithRetries tests creating a ServiceInstance
// with retry after temporary error without orphan mitigation.
func TestProvisionInstanceWithRetries(t *testing.T) {
	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()
	// configure first provision response with HTTP error
	ct.SetFirstOSBProvisionReactionsHTTPError(1, http.StatusConflict)
	require.NoError(t, ct.CreateSimpleClusterServiceBroker())
	require.NoError(t, ct.WaitForReadyBroker())

	// WHEN
	assert.NoError(t, ct.CreateServiceInstance())

	// THEN
	assert.NoError(t, ct.WaitForReadyInstance())
}
