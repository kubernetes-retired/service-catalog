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

package controller

import (
	"fmt"
	"testing"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	fakeosb "github.com/pmorie/go-open-service-broker-client/v2/fake"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	utilfeature "k8s.io/apiserver/pkg/util/feature"

	scfeatures "github.com/kubernetes-incubator/service-catalog/pkg/features"
)

// TestReconcileServiceInstanceNamespacedRefs tests synchronously provisioning a new service
func TestReconcileServiceInstanceNamespacedRefs(t *testing.T) {
	err := utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.NamespacedServiceBroker))
	if err != nil {
		t.Fatalf("Could not enable NamespacedServiceBroker feature flag.")
	}
	defer utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.NamespacedServiceBroker))

	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t, fakeosb.FakeClientConfiguration{
		ProvisionReaction: &fakeosb.ProvisionReaction{
			Response: &osb.ProvisionResponse{
				DashboardURL: &testDashboardURL,
			},
		},
	})

	addGetNamespaceReaction(fakeKubeClient)

	sharedInformers.ServiceBrokers().Informer().GetStore().Add(getTestServiceBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())
	sharedInformers.ServicePlans().Informer().GetStore().Add(getTestServicePlan())

	instance := getTestServiceInstanceWithNamespacedRefs()

	if err := reconcileServiceInstance(t, testController, instance); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	instance = assertServiceInstanceProvisionInProgressIsTheOnlyCatalogClientAction(t, fakeCatalogClient, instance)
	fakeCatalogClient.ClearActions()

	assertNumberOfBrokerActions(t, fakeBrokerClient.Actions(), 0)
	fakeKubeClient.ClearActions()

	if err := reconcileServiceInstance(t, testController, instance); err != nil {
		t.Fatalf("This should not fail : %v", err)
	}

	brokerActions := fakeBrokerClient.Actions()
	assertNumberOfBrokerActions(t, brokerActions, 1)
	assertProvision(t, brokerActions[0], &osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        testServiceInstanceGUID,
		ServiceID:         testServiceClassGUID,
		PlanID:            testServicePlanGUID,
		OrganizationGUID:  testNamespaceGUID,
		SpaceGUID:         testNamespaceGUID,
		Context:           testContext})

	instanceKey := testNamespace + "/" + testServiceInstanceName

	// Since synchronous operation, must not make it into the polling queue.
	if testController.instancePollingQueue.NumRequeues(instanceKey) != 0 {
		t.Fatalf("Expected polling queue to not have any record of test instance")
	}

	actions := fakeCatalogClient.Actions()
	assertNumberOfActions(t, actions, 1)

	// verify no kube resources created.
	// One single action comes from getting namespace uid
	kubeActions := fakeKubeClient.Actions()
	if err := checkKubeClientActions(kubeActions, []kubeClientAction{
		{verb: "get", resourceName: "namespaces", checkType: checkGetActionType},
	}); err != nil {
		t.Fatal(err)
	}

	updatedServiceInstance := assertUpdateStatus(t, actions[0], instance)
	assertServiceInstanceOperationSuccess(t, updatedServiceInstance, v1beta1.ServiceInstanceOperationProvision, testServicePlanName, testServicePlanGUID, instance)
	assertServiceInstanceDashboardURL(t, updatedServiceInstance, testDashboardURL)

	events := getRecordedEvents(testController)

	expectedEvent := normalEventBuilder(successProvisionReason).msg(successProvisionMessage)
	if err := checkEvents(events, expectedEvent.stringArr()); err != nil {
		t.Fatal(err)
	}
}

// TestReconcileServiceInstanceAsynchronousNamespacedRefs tests provisioning
// a new service from namespaced classes and plans, where the request results
// in a async response. Resulting status will indicate not ready and polling
// in progress.
func TestReconcileServiceInstanceAsynchronousNamespacedRefs(t *testing.T) {
	err := utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.NamespacedServiceBroker))
	if err != nil {
		t.Fatalf("Could not enable NamespacedServiceBroker feature flag.")
	}
	defer utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.NamespacedServiceBroker))

	key := osb.OperationKey(testOperation)

	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t, fakeosb.FakeClientConfiguration{
		ProvisionReaction: &fakeosb.ProvisionReaction{
			Response: &osb.ProvisionResponse{
				Async:        true,
				DashboardURL: &testDashboardURL,
				OperationKey: &key,
			},
		},
	})

	addGetNamespaceReaction(fakeKubeClient)

	sharedInformers.ServiceBrokers().Informer().GetStore().Add(getTestServiceBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())
	sharedInformers.ServicePlans().Informer().GetStore().Add(getTestServicePlan())

	instance := getTestServiceInstanceWithNamespacedRefs()

	if err := reconcileServiceInstance(t, testController, instance); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	instance = assertServiceInstanceProvisionInProgressIsTheOnlyCatalogClientAction(t, fakeCatalogClient, instance)
	fakeCatalogClient.ClearActions()
	fakeKubeClient.ClearActions()

	instanceKey := testNamespace + "/" + testServiceInstanceName

	if testController.instancePollingQueue.NumRequeues(instanceKey) != 0 {
		t.Fatalf("Expected polling queue to not have any record of test instance")
	}

	if err := reconcileServiceInstance(t, testController, instance); err != nil {
		t.Fatalf("This should not fail : %v", err)
	}

	brokerActions := fakeBrokerClient.Actions()
	assertNumberOfBrokerActions(t, brokerActions, 1)
	assertProvision(t, brokerActions[0], &osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        testServiceInstanceGUID,
		ServiceID:         testServiceClassGUID,
		PlanID:            testServicePlanGUID,
		OrganizationGUID:  testNamespaceGUID,
		SpaceGUID:         testNamespaceGUID,
		Context:           testContext,
	})

	actions := fakeCatalogClient.Actions()
	assertNumberOfActions(t, actions, 1)

	updatedServiceInstance := assertUpdateStatus(t, actions[0], instance)
	assertServiceInstanceAsyncStartInProgress(t, updatedServiceInstance, v1beta1.ServiceInstanceOperationProvision, testOperation, testServicePlanName, testServicePlanGUID, instance)
	assertServiceInstanceDashboardURL(t, updatedServiceInstance, testDashboardURL)

	// verify no kube resources created.
	// One single action comes from getting namespace uid
	kubeActions := fakeKubeClient.Actions()
	if e, a := 1, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	if testController.instancePollingQueue.NumRequeues(instanceKey) != 1 {
		t.Fatalf("Expected polling queue to have a record of seeing test instance once")
	}
}
