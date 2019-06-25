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
	"testing"

	scfeatures "github.com/kubernetes-sigs/service-catalog/pkg/features"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	utilfeature "k8s.io/apiserver/pkg/util/feature"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
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
// - unbind
// - deprovision
func TestBasicFlow(t *testing.T) {
	for tn, setupFunc := range map[string]func(ts *controllerTest){
		"sync": func(ts *controllerTest) {
		},
		"async instances with multiple polls": func(ct *controllerTest) {
			ct.EnableAsyncInstanceProvisioning()
			ct.EnableAsyncInstanceDeprovisioning()
			ct.SetFirstOSBPollLastOperationReactionsInProgress(2)
		},
		"async bindings": func(ct *controllerTest) {
			ct.EnableAsyncBind()
			ct.EnableAsyncUnbind()
		},
		"async instances and bindings": func(ct *controllerTest) {
			ct.EnableAsyncInstanceProvisioning()
			ct.EnableAsyncInstanceDeprovisioning()
			ct.EnableAsyncBind()
			ct.EnableAsyncUnbind()
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

			// Binding

			// WHEN
			assert.NoError(t, ct.CreateBinding())

			// THEN
			assert.NoError(t, ct.WaitForReadyBinding())
			// expected at least one binding call
			assert.NotZero(t, ct.NumberOfOSBBindingCalls())

			// Unbinding

			// WHEN
			require.NoError(t, ct.Unbind())

			// THEN
			assert.NoError(t, ct.WaitForUnbindStatus(v1beta1.ServiceBindingUnbindStatusSucceeded))
			assert.NotZero(t, ct.NumberOfOSBUnbindingCalls())

			// Deprovisioning

			// GIVEN
			// simulate k8s which removes the binding
			assert.NoError(t, ct.DeleteBinding())

			// WHEN
			assert.NoError(t, ct.Deprovision())

			// THEN
			assert.NoError(t, ct.WaitForDeprovisionStatus(v1beta1.ServiceInstanceDeprovisionStatusSucceeded))
			assert.NotZero(t, ct.NumberOfOSBDeprovisionCalls())
		})
	}
}

// TestClusterServiceClassRemovedFromCatalogAfterFiltering tests whether catalog restrictions filters service classes
func TestClusterServiceClassRemovedFromCatalogAfterFiltering(t *testing.T) {
	t.Parallel()
	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()
	assert.NoError(t, ct.CreateSimpleClusterServiceBroker())
	assert.NoError(t, ct.WaitForReadyBroker())
	ct.AssertClusterServiceClassAndPlan(t)

	// WHEN
	assert.NoError(t, ct.AddServiceClassRestrictionsToBroker())

	// THEN
	assert.NoError(t, ct.WaitForClusterServiceClassToNotExists())
}

// TestClusterServiceClassRemovedFromCatalogWithoutInstances tests whether a class marked as removed
// is removed by the controller.
func TestClusterServiceClassRemovedFromCatalogWithoutInstances(t *testing.T) {
	t.Parallel()
	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()
	assert.NoError(t, ct.CreateSimpleClusterServiceBroker())
	assert.NoError(t, ct.WaitForReadyBroker())
	ct.AssertClusterServiceClassAndPlan(t)

	// WHEN
	require.NoError(t, ct.MarkClusterServiceClassRemoved())

	// THEN
	assert.NoError(t, ct.WaitForClusterServiceClassToNotExists())
}

// TestClusterServiceClassRemovedFromCatalogWithoutInstances tests whether a plan marked as removed
// is removed by the controller.
func TestClusterServicePlanRemovedFromCatalogWithoutInstances(t *testing.T) {
	t.Parallel()
	// GIVEN
	ct := newControllerTest(t)
	defer ct.TearDown()
	assert.NoError(t, ct.CreateSimpleClusterServiceBroker())
	assert.NoError(t, ct.WaitForReadyBroker())
	ct.AssertClusterServiceClassAndPlan(t)

	// WHEN
	require.NoError(t, ct.MarkClusterServicePlanRemoved())

	// THEN
	assert.NoError(t, ct.WaitForClusterServicePlanToNotExists())
}
