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
	"net/http"
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pmorie/go-open-service-broker-client/v2"
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
