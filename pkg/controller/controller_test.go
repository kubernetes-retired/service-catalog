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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	fakebrokerapi "github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake"
	servicecataloginformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
	v1alpha1informers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions/servicecatalog/v1alpha1"

	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	clientgofake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	clientgotesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
)

const (
	serviceClassGUID = "SCGUID"
	planGUID         = "PGUID"
	instanceGUID     = "IGUID"
	bindingGUID      = "BGUID"

	testBrokerName        = "test-broker"
	testServiceClassName  = "test-serviceclass"
	testPlanName          = "test-plan"
	testInstanceName      = "test-instance"
	testBindingName       = "test-binding"
	testNamespace         = "test-ns"
	testBindingSecretName = "test-secret"
)

const testCatalog = `{
  "services": [{
    "name": "fake-service",
    "id": "acb56d7c-XXXX-XXXX-XXXX-feb140a59a66",
    "description": "fake service",
    "tags": ["no-sql", "relational"],
    "requires": ["route_forwarding"],
    "max_db_per_node": 5,
    "bindable": true,
    "metadata": {
      "provider": {
        "name": "The name"
      },
      "listing": {
        "imageUrl": "http://example.com/cat.gif",
        "blurb": "Add a blurb here",
        "longDescription": "A long time ago, in a galaxy far far away..."
      },
      "displayName": "The Fake Broker"
    },
    "dashboard_client": {
      "id": "398e2f8e-XXXX-XXXX-XXXX-19a71ecbcf64",
      "secret": "277cabb0-XXXX-XXXX-XXXX-7822c0a90e5d",
      "redirect_uri": "http://localhost:1234"
    },
    "plan_updateable": true,
    "plans": [{
      "name": "fake-plan-1",
      "id": "d3031751-XXXX-XXXX-XXXX-a42377d3320e",
      "description": "Shared fake Server, 5tb persistent disk, 40 max concurrent connections",
      "max_storage_tb": 5,
      "metadata": {
        "costs":[
            {
               "amount":{
                  "usd":99.0
               },
               "unit":"MONTHLY"
            },
            {
               "amount":{
                  "usd":0.99
               },
               "unit":"1GB of messages over 20GB"
            }
         ],
        "bullets": [
            "Shared fake server",
            "5 TB storage",
            "40 concurrent connections"
        ]
      }
    }, {
      "name": "fake-plan-2",
      "id": "0f4008b5-XXXX-XXXX-XXXX-dace631cd648",
      "description": "Shared fake Server, 5tb persistent disk, 40 max concurrent connections. 100 async",
      "max_storage_tb": 5,
      "metadata": {
        "costs":[
            {
               "amount":{
                  "usd":199.0
               },
               "unit":"MONTHLY"
            },
            {
               "amount":{
                  "usd":0.99
               },
               "unit":"1GB of messages over 20GB"
            }
         ],
        "bullets": [
          "40 concurrent connections"
        ]
      }
    }]
  }]
}`

const testCatalogWithMultipleServices = `{
  "services": [
    {
      "name": "service1",
      "description": "service 1 description",
      "metadata": {
        "field1": "value1"
      },
      "plans": [{
        "name": "s1plan1",
        "id": "s1_plan1_id",
        "description": "s1 plan1 description"
      },
      {
        "name": "s1plan2",
        "id": "s1_plan2_id",
        "description": "s1 plan2 description",
        "metadata": {
          "planmeta": "planvalue"
        }
      }]
    },
    {
      "name": "service2",
      "description": "service 2 description",
      "metadata": ["first", "second", "third"],
      "plans": [{
        "name": "s2plan1",
        "id": "s2_plan1_id",
        "description": "s2 plan1 description"
      },
      {
        "name": "s2plan2",
        "id": "s2_plan2_id",
        "description": "s2 plan2 description",
        "metadata": {
          "planmeta": "planvalue"
      }
      }]
    }
]}`

// broker used in most of the tests that need a broker
func getTestBroker() *v1alpha1.Broker {
	return &v1alpha1.Broker{
		ObjectMeta: metav1.ObjectMeta{Name: testBrokerName},
		Spec: v1alpha1.BrokerSpec{
			URL: "https://example.com",
		},
	}
}

func getTestBrokerWithStatus(status v1alpha1.ConditionStatus) *v1alpha1.Broker {
	broker := getTestBroker()
	broker.Status = v1alpha1.BrokerStatus{
		Conditions: []v1alpha1.BrokerCondition{{
			Type:               v1alpha1.BrokerConditionReady,
			Status:             status,
			LastTransitionTime: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
		}},
	}

	return broker
}

// service class wired to the result of getTestBroker()
func getTestServiceClass() *v1alpha1.ServiceClass {
	return &v1alpha1.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{Name: testServiceClassName},
		BrokerName: testBrokerName,
		OSBGUID:    serviceClassGUID,
		Plans: []v1alpha1.ServicePlan{{
			Name:    testPlanName,
			OSBFree: true,
			OSBGUID: planGUID,
		}},
	}
}

// broker catalog that provides the service class named in of
// getTestServiceClass()
func getTestCatalog() *brokerapi.Catalog {
	return &brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				Name:        testServiceClassName,
				ID:          serviceClassGUID,
				Description: "a test service",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        testPlanName,
						Free:        true,
						ID:          planGUID,
						Description: "a test plan",
					},
				},
			},
		},
	}
}

// instance referencing the result of getTestServiceClass()
func getTestInstance() *v1alpha1.Instance {
	return &v1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{Name: testInstanceName, Namespace: testNamespace},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}
}

// binding referencing the result of getTestBinding()
func getTestBinding() *v1alpha1.Binding {
	return &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{Name: testBindingName, Namespace: testNamespace},
		Spec: v1alpha1.BindingSpec{
			InstanceRef: v1.LocalObjectReference{Name: testInstanceName},
			OSBGUID:     bindingGUID,
		},
	}

}

type instanceParameters struct {
	Name string            `json:"name"`
	Args map[string]string `json:"args"`
}

type bindingParameters struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

func TestShouldReconcileBroker(t *testing.T) {
	cases := []struct {
		name      string
		broker    *v1alpha1.Broker
		now       time.Time
		interval  time.Duration
		reconcile bool
	}{
		{
			name:      "no status",
			broker:    getTestBroker(),
			now:       time.Now(),
			interval:  3 * time.Minute,
			reconcile: true,
		},
		{
			name: "no ready condition",
			broker: func() *v1alpha1.Broker {
				b := getTestBroker()
				b.Status = v1alpha1.BrokerStatus{
					Conditions: []v1alpha1.BrokerCondition{
						{
							Type:   v1alpha1.BrokerConditionType("NotARealCondition"),
							Status: v1alpha1.ConditionTrue,
						},
					},
				}
				return b
			}(),
			now:       time.Now(),
			interval:  3 * time.Minute,
			reconcile: true,
		},
		{
			name:      "not ready",
			broker:    getTestBrokerWithStatus(v1alpha1.ConditionFalse),
			now:       time.Now(),
			interval:  3 * time.Minute,
			reconcile: true,
		},
		{
			name: "ready, interval elapsed",
			broker: func() *v1alpha1.Broker {
				broker := getTestBrokerWithStatus(v1alpha1.ConditionTrue)
				return broker
			}(),
			now:       time.Now(),
			interval:  3 * time.Minute,
			reconcile: true,
		},
		{
			name: "ready, interval not elapsed",
			broker: func() *v1alpha1.Broker {
				broker := getTestBrokerWithStatus(v1alpha1.ConditionTrue)
				return broker
			}(),
			now:       time.Now(),
			interval:  3 * time.Hour,
			reconcile: false,
		},
	}

	for _, tc := range cases {
		var ltt *time.Time
		if len(tc.broker.Status.Conditions) != 0 {
			ltt = &tc.broker.Status.Conditions[0].LastTransitionTime.Time
		}

		t.Logf("%v: now: %v, interval: %v, last transition time: %v", tc.name, tc.now, tc.interval, ltt)
		actual := shouldReconcileBroker(tc.broker, tc.now, tc.interval)

		if e, a := tc.reconcile, actual; e != a {
			t.Errorf("%v: unexpected result: expected %v, got %v", tc.name, e, a)
		}
	}
}

func TestReconcileBroker(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, _ := newTestController(t)
	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	testController.reconcileBroker(getTestBroker())

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 2); err != nil {
		t.Fatal(err)
	}

	// first action should be a create action for a service class
	if _, err := assertCreate(actions[0], getTestServiceClass()); err != nil {
		t.Fatal(err)
	}

	// second action should be an update action for broker status subresource
	updatedBroker, err := assertUpdateStatus(actions[1], getTestBroker())
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBrokerReadyTrue(updatedBroker); err != nil {
		t.Fatal(err)
	}

	// verify no kube resources created
	if err := assertNumberOfActions(t, fakeKubeClient.Actions(), 0); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeNormal + " " + successFetchedCatalogReason + " " + successFetchedCatalogMessage
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileBrokerExistingServiceClass(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	testServiceClass := getTestServiceClass()
	sharedInformers.ServiceClasses().Informer().GetStore().Add(testServiceClass)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	testController.reconcileBroker(getTestBroker())

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 2); err != nil {
		t.Fatal(err)
	}

	// first action should be an update action for a service class
	if _, err := assertUpdate(actions[0], testServiceClass); err != nil {
		t.Fatal(err)
	}

	// second action should be an update action for broker status subresource
	updatedBroker, err := assertUpdateStatus(actions[1], getTestBroker())
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBrokerReadyTrue(updatedBroker); err != nil {
		t.Fatal(err)
	}

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 0); err != nil {
		t.Fatal(err)
	}
}

func TestReconcileBrokerExistingServiceClassDifferentOSBGUID(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	testServiceClass := getTestServiceClass()
	testServiceClass.OSBGUID = "notTheSame"
	sharedInformers.ServiceClasses().Informer().GetStore().Add(testServiceClass)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	testController.reconcileBroker(getTestBroker())

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	updatedBroker, err := assertUpdateStatus(actions[0], getTestBroker())
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBrokerReadyFalse(updatedBroker); err != nil {
		t.Fatal(err)
	}

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 0); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorSyncingCatalogReason + ` Error reconciling serviceClass "test-serviceclass" (broker "test-broker"): ServiceClass "test-serviceclass" already exists with OSB guid "notTheSame", received different guid "SCGUID"`
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event; expected\n%v, got\n%v", e, a)
	}
}

func TestReconcileBrokerExistingServiceClassDifferentBroker(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	testServiceClass := getTestServiceClass()
	testServiceClass.BrokerName = "notTheSame"
	sharedInformers.ServiceClasses().Informer().GetStore().Add(testServiceClass)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	testController.reconcileBroker(getTestBroker())

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	updatedBroker, err := assertUpdateStatus(actions[0], getTestBroker())
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBrokerReadyFalse(updatedBroker); err != nil {
		t.Fatal(err)
	}

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 0); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorSyncingCatalogReason + ` Error reconciling serviceClass "test-serviceclass" (broker "test-broker"): ServiceClass "test-serviceclass" for Broker "test-broker" already exists for Broker "notTheSame"`
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event; expected\n%v, got\n%v", e, a)
	}
}

func TestReconcileBrokerDelete(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, _, testController, sharedInformers := newTestController(t)

	testServiceClass := getTestServiceClass()
	sharedInformers.ServiceClasses().Informer().GetStore().Add(testServiceClass)

	broker := getTestBroker()
	broker.DeletionTimestamp = &metav1.Time{}
	broker.Finalizers = []string{"kubernetes"}

	testController.reconcileBroker(broker)

	// Verify no core kube actions occurred
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 0); err != nil {
		t.Fatal(err)
	}

	actions := fakeCatalogClient.Actions()
	// The three actions should be:
	// 0. Deleting the associated ServiceClass
	// 1. Updating the ready condition
	// 2. Removing the finalizer
	if err := assertNumberOfActions(t, actions, 3); err != nil {
		t.Fatal(err)
	}

	if err := assertDelete(actions[0], testServiceClass); err != nil {
		t.Fatal(err)
	}

	updatedBroker, err := assertUpdateStatus(actions[1], broker)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBrokerReadyFalse(updatedBroker); err != nil {
		t.Fatal(err)
	}

	updatedBroker, err = assertUpdateStatus(actions[2], broker)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertEmptyFinalizers(updatedBroker); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeNormal + " " + successBrokerDeletedReason + " " + "The broker test-broker was deleted successfully."
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileBrokerErrorFetchingCatalog(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, _ := newTestController(t)

	fakeBrokerClient.CatalogClient.RetErr = fakebrokerapi.ErrInstanceNotFound
	broker := getTestBroker()

	testController.reconcileBroker(broker)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	updatedBroker, err := assertUpdateStatus(actions[0], broker)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBrokerReadyFalse(updatedBroker); err != nil {
		t.Fatal(err)
	}

	if err := assertNumberOfActions(t, fakeKubeClient.Actions(), 0); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorFetchingCatalogReason + " " + "Error getting broker catalog for broker \"test-broker\": instance not found"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileBrokerWithAuthError(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, _, testController, _ := newTestController(t)

	broker := getTestBroker()
	broker.Spec.AuthSecret = &v1.ObjectReference{
		Namespace: "does_not_exist",
		Name:      "auth-name",
	}

	fakeKubeClient.AddReactor("get", "secrets", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("no secret defined")
	})

	testController.reconcileBroker(broker)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	updatedBroker, err := assertUpdateStatus(actions[0], broker)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBrokerReadyFalse(updatedBroker); err != nil {
		t.Fatal(err)
	}

	// verify one kube action occurred
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 1); err != nil {
		t.Fatal(err)
	}

	getAction := kubeActions[0].(clientgotesting.GetAction)
	if e, a := "get", getAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on action; expected %v, got %v", e, a)
	}
	if e, a := "secrets", getAction.GetResource().Resource; e != a {
		t.Fatalf("Unexpected resource on action; expected %v, got %v", e, a)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorAuthCredentialsReason + " " + "Error getting broker auth credentials for broker \"test-broker\": no secret defined"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileBrokerWithReconcileError(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, _, testController, _ := newTestController(t)

	broker := getTestBroker()
	broker.Spec.AuthSecret = &v1.ObjectReference{
		Namespace: "does_not_exist",
		Name:      "auth-name",
	}

	fakeCatalogClient.AddReactor("create", "serviceclasses", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("error creating serviceclass")
	})

	testController.reconcileBroker(broker)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	updatedBroker, err := assertUpdateStatus(actions[0], broker)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBrokerReadyFalse(updatedBroker); err != nil {
		t.Fatal(err)
	}

	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 1); err != nil {
		t.Fatal(err)
	}

	getAction := kubeActions[0].(clientgotesting.GetAction)
	if e, a := "get", getAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on action; expected %v, got %v", e, a)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorAuthCredentialsReason + " " + "Error getting broker auth credentials for broker \"test-broker\": auth secret didn't contain username"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestUpdateBrokerCondition(t *testing.T) {
	cases := []struct {
		name                  string
		input                 *v1alpha1.Broker
		status                v1alpha1.ConditionStatus
		transitionTimeChanged bool
	}{

		{
			name:                  "initially unset",
			input:                 getTestBroker(),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: true,
		},
		{
			name:                  "not ready -> not ready",
			input:                 getTestBrokerWithStatus(v1alpha1.ConditionFalse),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: false,
		},
		{
			name:                  "not ready -> ready",
			input:                 getTestBrokerWithStatus(v1alpha1.ConditionFalse),
			status:                v1alpha1.ConditionTrue,
			transitionTimeChanged: true,
		},
		{
			name:                  "ready -> ready",
			input:                 getTestBrokerWithStatus(v1alpha1.ConditionTrue),
			status:                v1alpha1.ConditionTrue,
			transitionTimeChanged: false,
		},
		{
			name:                  "ready -> not ready",
			input:                 getTestBrokerWithStatus(v1alpha1.ConditionTrue),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: true,
		},
	}

	for _, tc := range cases {
		_, fakeCatalogClient, _, testController, _ := newTestController(t)

		clone, err := api.Scheme.DeepCopy(tc.input)
		if err != nil {
			t.Errorf("%v: deep copy failed", tc.name)
			continue
		}

		inputClone := clone.(*v1alpha1.Broker)

		err = testController.updateBrokerCondition(tc.input, v1alpha1.BrokerConditionReady, tc.status, "reason", "message")
		if err != nil {
			t.Errorf("%v: error updating broker condition: %v", tc.name, err)
			continue
		}

		if !reflect.DeepEqual(tc.input, inputClone) {
			t.Errorf("%v: updating broker condition mutated input: expected %v, got %v", tc.name, inputClone, tc.input)
			continue
		}

		actions := fakeCatalogClient.Actions()
		if err := assertNumberOfActions(t, actions, 1); err != nil {
			t.Errorf("%v: %v", tc.name, err)
			continue
		}

		updatedBroker, err := assertUpdateStatus(actions[0], inputClone)
		if err != nil {
			t.Errorf("%v: %v", tc.name, err)
			continue
		}
		updateActionObject, ok := updatedBroker.(*v1alpha1.Broker)
		if !ok {
			t.Errorf("%v: couldn't convert to a broker", tc.name)
			continue
		}

		var initialTs metav1.Time
		if len(inputClone.Status.Conditions) != 0 {
			initialTs = inputClone.Status.Conditions[0].LastTransitionTime
		}

		newTs := updateActionObject.Status.Conditions[0].LastTransitionTime

		if tc.transitionTimeChanged && initialTs == newTs {
			t.Errorf("%v: transition time didn't change when it should have", tc.name)
			continue
		} else if !tc.transitionTimeChanged && initialTs != newTs {
			t.Errorf("%v: transition time changed when it shouldn't have", tc.name)
			continue
		}
	}
}

func TestReconcileInstanceNonExistentServiceClass(t *testing.T) {
	_, fakeCatalogClient, _, testController, _ := newTestController(t)

	instance := &v1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{Name: testInstanceName},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: "nothere",
			PlanName:         "nothere",
			OSBGUID:          instanceGUID,
		},
	}

	testController.reconcileInstance(instance)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	// There should only be one action that says it failed because no such class exists.
	updatedInstance, err := assertUpdateStatus(actions[0], instance)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertInstanceReadyFalse(updatedInstance, errorNonexistentServiceClassReason); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorNonexistentServiceClassReason + " " + "Instance \"/test-instance\" references a non-existent ServiceClass \"nothere\""
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileInstanceNonExistentBroker(t *testing.T) {
	_, fakeCatalogClient, _, testController, sharedInformers := newTestController(t)

	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := getTestInstance()

	testController.reconcileInstance(instance)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	// There should only be one action that says it failed because no such broker exists.
	updatedInstance, err := assertUpdateStatus(actions[0], instance)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertInstanceReadyFalse(updatedInstance, errorNonexistentBrokerReason); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorNonexistentBrokerReason + " " + "Instance \"test-ns/test-instance\" references a non-existent broker \"test-broker\""
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileInstanceWithAuthError(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, _, testController, sharedInformers := newTestController(t)

	broker := getTestBroker()
	broker.Spec.AuthSecret = &v1.ObjectReference{
		Namespace: "does_not_exist",
		Name:      "auth-name",
	}
	sharedInformers.Brokers().Informer().GetStore().Add(broker)
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := getTestInstance()

	fakeKubeClient.AddReactor("get", "secrets", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("no secret defined")
	})

	testController.reconcileInstance(instance)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	updatedInstance, err := assertUpdateStatus(actions[0], instance)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertInstanceReadyFalse(updatedInstance, errorAuthCredentialsReason); err != nil {
		t.Fatal(err)
	}

	// verify one kube action occurred
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 1); err != nil {
		t.Fatal(err)
	}

	getAction := kubeActions[0].(clientgotesting.GetAction)
	if e, a := "get", getAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on action; expected %v, got %v", e, a)
	}
	if e, a := "secrets", getAction.GetResource().Resource; e != a {
		t.Fatalf("Unexpected resource on action; expected %v, got %v", e, a)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorAuthCredentialsReason + " " + "Error getting broker auth credentials for broker \"test-broker\": no secret defined"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileInstanceNonExistentServicePlan(t *testing.T) {
	_, fakeCatalogClient, _, testController, sharedInformers := newTestController(t)

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := &v1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{Name: testInstanceName},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         "nothere",
			OSBGUID:          instanceGUID,
		},
	}

	testController.reconcileInstance(instance)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	// There should only be one action that says it failed because no such class exists.
	updatedInstance, err := assertUpdateStatus(actions[0], instance)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertInstanceReadyFalse(updatedInstance, errorNonexistentServicePlanReason); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorNonexistentServicePlanReason + " " + "Instance \"/test-instance\" references a non-existent ServicePlan \"nothere\" on ServiceClass \"test-serviceclass\""
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileInstanceWithParameters(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := getTestInstance()

	parameters := instanceParameters{Name: "test-param", Args: make(map[string]string)}
	parameters.Args["first"] = "first-arg"
	parameters.Args["second"] = "second-arg"

	b, err := json.Marshal(parameters)
	if err != nil {
		t.Fatalf("Failed to marshal parameters %v : %v", parameters, err)
	}
	instance.Spec.Parameters = &runtime.RawExtension{Raw: b}

	testController.reconcileInstance(instance)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	// verify no kube resources created
	// One single action comes from getting namespace uid
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 1); err != nil {
		t.Fatal(err)
	}

	updatedInstance, err := assertUpdateStatus(actions[0], instance)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertInstanceReadyTrue(updatedInstance); err != nil {
		t.Fatal(err)
	}

	updateObject, ok := updatedInstance.(*v1alpha1.Instance)
	if !ok {
		t.Fatalf("couldn't convert to *v1alpha1.Instance")
	}

	// Verify parameters are what we'd expect them to be, basically name, map with two values in it.
	if len(updateObject.Spec.Parameters.Raw) == 0 {
		t.Fatalf("Parameters was unexpectedly empty")
	}
	if si, ok := fakeBrokerClient.InstanceClient.Instances[instanceGUID]; !ok {
		t.Fatalf("Did not find the created Instance in fakeInstanceClient after creation")
	} else {
		if len(si.Parameters) == 0 {
			t.Fatalf("Expected parameters but got none")
		}
		if e, a := "test-param", si.Parameters["name"].(string); e != a {
			t.Fatalf("Unexpected name for parameters: expected %v, got %v", e, a)
		}
		argsMap := si.Parameters["args"].(map[string]interface{})
		if e, a := "first-arg", argsMap["first"].(string); e != a {
			t.Fatalf("Unexpected value in parameter map: expected %v, got %v", e, a)
		}
		if e, a := "second-arg", argsMap["second"].(string); e != a {
			t.Fatalf("Unexpected value in parameter map: expected %v, got %v", e, a)
		}
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeNormal + " " + successProvisionReason + " " + "The instance was provisioned successfully"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileInstanceWithInvalidParameters(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := getTestInstance()
	parameters := instanceParameters{Name: "test-param", Args: make(map[string]string)}
	parameters.Args["first"] = "first-arg"
	parameters.Args["second"] = "second-arg"

	b, err := json.Marshal(parameters)
	if err != nil {
		t.Fatalf("Failed to marshal parameters %v : %v", parameters, err)
	}
	// corrupt the byte slice to begin with a '!' instead of an opening JSON bracket '{'
	b[0] = 0x21
	instance.Spec.Parameters = &runtime.RawExtension{Raw: b}

	testController.reconcileInstance(instance)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 0); err != nil {
		t.Fatal(err)
	}

	updatedInstance, err := assertUpdateStatus(actions[0], instance)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertInstanceReadyFalse(updatedInstance); err != nil {
		t.Fatal(err)
	}

	if si, notOK := fakeBrokerClient.InstanceClient.Instances[instanceGUID]; notOK {
		t.Fatalf("Unexpectedly found created Instance: %+v in fakeInstanceClient after creation", si)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorWithParameters + " " + "Failed to unmarshal Instance parameters"
	if e, a := expectedEvent, events[0]; !strings.Contains(a, e) { // event contains RawExtension, so just compare error message
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileInstanceWithProvisionFailure(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := getTestInstance()
	parameters := instanceParameters{Name: "test-param", Args: make(map[string]string)}
	parameters.Args["first"] = "first-arg"
	parameters.Args["second"] = "second-arg"

	b, err := json.Marshal(parameters)
	if err != nil {
		t.Fatalf("Failed to marshal parameters %v : %v", parameters, err)
	}
	instance.Spec.Parameters = &runtime.RawExtension{Raw: b}

	fakeBrokerClient.InstanceClient.CreateErr = errors.New("fake creation failure")

	testController.reconcileInstance(instance)

	// verify no kube resources created
	// One single action comes from getting namespace uid
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 1); err != nil {
		t.Fatal(err)
	}

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	updatedInstance, err := assertUpdateStatus(actions[0], instance)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertInstanceReadyFalse(updatedInstance); err != nil {
		t.Fatal(err)
	}

	if si, notOK := fakeBrokerClient.InstanceClient.Instances[instanceGUID]; notOK {
		t.Fatalf("Unexpectedly found created Instance: %+v in fakeInstanceClient after creation", si)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorProvisionCalledReason + " " + "Error provisioning Instance \"test-ns/test-instance\" of ServiceClass \"test-serviceclass\" at Broker \"test-broker\": fake creation failure"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileInstance(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	testNsUID := "test_uid_foo"

	fakeKubeClient.AddReactor("get", "namespaces", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				UID: types.UID(testNsUID),
			},
		}, nil
	})

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := getTestInstance()

	testController.reconcileInstance(instance)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	// verify no kube resources created.
	// One single action comes from getting namespace uid
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 1); err != nil {
		t.Fatal(err)
	}

	updatedInstance, err := assertUpdateStatus(actions[0], instance)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertInstanceReadyTrue(updatedInstance); err != nil {
		t.Fatal(err)
	}

	si, ok := fakeBrokerClient.InstanceClient.Instances[instanceGUID]
	if !ok {
		t.Fatalf("Did not find the created Instance in fakeInstanceClient after creation")
	}
	if len(si.Parameters) > 0 {
		t.Fatalf("Unexpected parameters, expected none, got %+v", si.Parameters)
	}
	if testNsUID != si.OrganizationGUID {
		t.Fatalf("Unexpected OrganizationGUID: expected %q, got %q", testNsUID, si.OrganizationGUID)
	}
	if testNsUID != si.SpaceGUID {
		t.Fatalf("Unexpected SpaceGUID: expected %q, got %q", testNsUID, si.SpaceGUID)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeNormal + " " + successProvisionReason + " " + successProvisionMessage
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileInstanceNamespaceError(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	fakeKubeClient.AddReactor("get", "namespaces", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, &v1.Namespace{}, errors.New("No namespace")
	})

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := getTestInstance()

	testController.reconcileInstance(instance)

	// verify no kube resources created.
	// One single action comes from getting namespace uid
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 1); err != nil {
		t.Fatal(err)
	}

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	updateAction := actions[0].(clientgotesting.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}
	updatedInstance := updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := instance.Name, updatedInstance.Name; e != a {
		t.Fatalf("Unexpected name of instance: expected %v, got %v", e, a)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorFindingNamespaceInstanceReason + " " + "Failed to get namespace \"test-ns\" during instance create: No namespace"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileInstanceDelete(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.InstanceClient.Instances = map[string]*brokerapi.ServiceInstance{
		instanceGUID: {},
	}

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := getTestInstance()
	instance.ObjectMeta.DeletionTimestamp = &metav1.Time{}
	instance.ObjectMeta.Finalizers = []string{"kubernetes"}

	fakeCatalogClient.AddReactor("get", "instances", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, instance, nil
	})

	testController.reconcileInstance(instance)

	// Verify no core kube actions occurred
	kubeActions := fakeKubeClient.Actions()
	if err := assertNumberOfActions(t, kubeActions, 0); err != nil {
		t.Fatal(err)
	}

	actions := fakeCatalogClient.Actions()
	// The three actions should be:
	// 0. Updating the ready condition
	// 1. Get against the instance
	// 2. Removing the finalizer
	if err := assertNumberOfActions(t, actions, 3); err != nil {
		t.Fatal(err)
	}

	updatedInstance, err := assertUpdateStatus(actions[0], instance)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertInstanceReadyFalse(updatedInstance); err != nil {
		t.Fatal(err)
	}

	if err := assertGet(actions[1], instance); err != nil {
		t.Fatal(err)
	}
	updatedInstance, err = assertUpdateStatus(actions[2], instance)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertEmptyFinalizers(updatedInstance); err != nil {
		t.Fatal(err)
	}

	if _, ok := fakeBrokerClient.InstanceClient.Instances[instanceGUID]; ok {
		t.Fatalf("Found the deleted Instance in fakeInstanceClient after deletion")
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeNormal + " " + successDeprovisionReason + " " + "The instance was deprovisioned successfully"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestUpdateInstanceCondition(t *testing.T) {
	getTestInstanceWithStatus := func(status v1alpha1.ConditionStatus) *v1alpha1.Instance {
		instance := getTestInstance()
		instance.Status = v1alpha1.InstanceStatus{
			Conditions: []v1alpha1.InstanceCondition{{
				Type:               v1alpha1.InstanceConditionReady,
				Status:             status,
				LastTransitionTime: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
			}},
		}

		return instance
	}

	cases := []struct {
		name                  string
		input                 *v1alpha1.Instance
		status                v1alpha1.ConditionStatus
		transitionTimeChanged bool
	}{

		{
			name:                  "initially unset",
			input:                 getTestInstance(),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: true,
		},
		{
			name:                  "not ready -> not ready",
			input:                 getTestInstanceWithStatus(v1alpha1.ConditionFalse),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: false,
		},
		{
			name:                  "not ready -> ready",
			input:                 getTestInstanceWithStatus(v1alpha1.ConditionFalse),
			status:                v1alpha1.ConditionTrue,
			transitionTimeChanged: true,
		},
		{
			name:                  "ready -> ready",
			input:                 getTestInstanceWithStatus(v1alpha1.ConditionTrue),
			status:                v1alpha1.ConditionTrue,
			transitionTimeChanged: false,
		},
		{
			name:                  "ready -> not ready",
			input:                 getTestInstanceWithStatus(v1alpha1.ConditionTrue),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: true,
		},
	}

	for _, tc := range cases {
		_, fakeCatalogClient, _, testController, _ := newTestController(t)

		clone, err := api.Scheme.DeepCopy(tc.input)
		if err != nil {
			t.Errorf("%v: deep copy failed", tc.name)
			continue
		}
		inputClone := clone.(*v1alpha1.Instance)

		err = testController.updateInstanceCondition(tc.input, v1alpha1.InstanceConditionReady, tc.status, "reason", "message")
		if err != nil {
			t.Errorf("%v: error updating broker condition: %v", tc.name, err)
			continue
		}

		if !reflect.DeepEqual(tc.input, inputClone) {
			t.Errorf("%v: updating broker condition mutated input: expected %v, got %v", tc.name, inputClone, tc.input)
			continue
		}

		actions := fakeCatalogClient.Actions()
		if err := assertNumberOfActions(t, actions, 1); err != nil {
			t.Errorf("%v: %v", tc.name, err)
			continue
		}

		updatedInstance, err := assertUpdateStatus(actions[0], inputClone)
		if err != nil {
			t.Errorf("%v: %v", tc.name, err)
			continue
		}
		updateActionObject, ok := updatedInstance.(*v1alpha1.Instance)
		if !ok {
			t.Errorf("%v: couldn't convert to an instance", tc.name)
			continue
		}

		var initialTs metav1.Time
		if len(inputClone.Status.Conditions) != 0 {
			initialTs = inputClone.Status.Conditions[0].LastTransitionTime
		}

		newTs := updateActionObject.Status.Conditions[0].LastTransitionTime

		if tc.transitionTimeChanged && initialTs == newTs {
			t.Errorf("%v: transition time didn't change when it should have", tc.name)
			continue
		} else if !tc.transitionTimeChanged && initialTs != newTs {
			t.Errorf("%v: transition time changed when it shouldn't have", tc.name)
			continue
		}
	}
}

func TestReconcileBindingNonExistingInstance(t *testing.T) {
	_, fakeCatalogClient, _, testController, _ := newTestController(t)

	binding := &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{Name: testBindingName},
		Spec: v1alpha1.BindingSpec{
			InstanceRef: v1.LocalObjectReference{Name: "nothere"},
			OSBGUID:     bindingGUID,
		},
	}

	testController.reconcileBinding(binding)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	// There should only be one action that says it failed because no such instance exists.
	updatedBinding, err := assertUpdateStatus(actions[0], binding)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBindingReadyFalse(updatedBinding, errorNonexistentInstanceReason); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorNonexistentInstanceReason + " " + "Binding \"/test-binding\" references a non-existent Instance \"/nothere\""
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileBindingNonExistingServiceClass(t *testing.T) {
	_, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())
	instance := &v1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{Name: testInstanceName, Namespace: testNamespace},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: "nothere",
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}
	sharedInformers.Instances().Informer().GetStore().Add(instance)

	binding := &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{Name: testBindingName, Namespace: testNamespace},
		Spec: v1alpha1.BindingSpec{
			InstanceRef: v1.LocalObjectReference{Name: testInstanceName},
			OSBGUID:     bindingGUID,
		},
	}

	testController.reconcileBinding(binding)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	// There should only be one action that says it failed because no such service class.
	updatedBinding, err := assertUpdateStatus(actions[0], binding)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBindingReadyFalse(updatedBinding, errorNonexistentServiceClassMessage); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorNonexistentServiceClassMessage + " " + "Binding \"test-ns/test-binding\" references a non-existent ServiceClass \"nothere\""
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileBindingWithParameters(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	testNsUID := "test_ns_uid"

	fakeKubeClient.AddReactor("get", "namespaces", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				UID: types.UID(testNsUID),
			},
		}, nil
	})

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())
	sharedInformers.Instances().Informer().GetStore().Add(getTestInstance())

	binding := &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{Name: testBindingName, Namespace: testNamespace},
		Spec: v1alpha1.BindingSpec{
			InstanceRef: v1.LocalObjectReference{Name: testInstanceName},
			OSBGUID:     bindingGUID,
		},
	}

	parameters := bindingParameters{Name: "test-param"}
	parameters.Args = append(parameters.Args, "first-arg")
	parameters.Args = append(parameters.Args, "second-arg")
	b, err := json.Marshal(parameters)
	if err != nil {
		t.Fatalf("Failed to marshal parameters %v : %v", parameters, err)
	}
	binding.Spec.Parameters = &runtime.RawExtension{Raw: b}

	testController.reconcileBinding(binding)

	if testNsUID != fakeBrokerClient.Bindings[fakebrokerapi.BindingsMapKey(instanceGUID, bindingGUID)].AppID {
		t.Fatalf("Unexpected broker AppID: expected %q, got %q", testNsUID, fakeBrokerClient.Bindings[instanceGUID+":"+bindingGUID].AppID)
	}

	bindResource := fakeBrokerClient.BindingRequests[fakebrokerapi.BindingsMapKey(instanceGUID, bindingGUID)].BindResource
	if appGUID := bindResource["app_guid"]; testNsUID != fmt.Sprintf("%v", appGUID) {
		t.Fatalf("Unexpected broker AppID: expected %q, got %q", testNsUID, appGUID)
	}

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	// There should only be one action that says binding was created
	updatedBinding, err := assertUpdateStatus(actions[0], binding)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBindingReadyTrue(updatedBinding); err != nil {
		t.Fatal(err)
	}

	updateObject, ok := updatedBinding.(*v1alpha1.Binding)
	if !ok {
		t.Fatalf("couldn't convert to *v1alpha1.Binding")
	}

	// Verify parameters are what we'd expect them to be, basically name, array with two values in it.
	if len(updateObject.Spec.Parameters.Raw) == 0 {
		t.Fatalf("Parameters was unexpectedly empty")
	}
	if b, ok := fakeBrokerClient.BindingClient.Bindings[fakebrokerapi.BindingsMapKey(instanceGUID, bindingGUID)]; !ok {
		t.Fatalf("Did not find the created Binding in fakeInstanceBinding after creation")
	} else {
		if len(b.Parameters) == 0 {
			t.Fatalf("Expected parameters, but got none")
		}
		if e, a := "test-param", b.Parameters["name"].(string); e != a {
			t.Fatalf("Unexpected name for parameters: expected %v, got %v", e, a)
		}
		argsArray := b.Parameters["args"].([]interface{})
		if len(argsArray) != 2 {
			t.Fatalf("Expected 2 elements in args array, but got %d", len(argsArray))
		}
		foundFirst := false
		foundSecond := false
		for _, el := range argsArray {
			if el.(string) == "first-arg" {
				foundFirst = true
			}
			if el.(string) == "second-arg" {
				foundSecond = true
			}
		}
		if !foundFirst {
			t.Fatalf("Failed to find 'first-arg' in array, was %v", argsArray)
		}
		if !foundSecond {
			t.Fatalf("Failed to find 'second-arg' in array, was %v", argsArray)
		}
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeNormal + " " + successInjectedBindResultReason + " " + successInjectedBindResultMessage
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileBindingNamespaceError(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	fakeKubeClient.AddReactor("get", "namespaces", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, &v1.Namespace{}, errors.New("No namespace")
	})

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())
	sharedInformers.Instances().Informer().GetStore().Add(getTestInstance())

	binding := &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{Name: testBindingName, Namespace: testNamespace},
		Spec: v1alpha1.BindingSpec{
			InstanceRef: v1.LocalObjectReference{Name: testInstanceName},
			OSBGUID:     bindingGUID,
		},
	}

	testController.reconcileBinding(binding)

	actions := fakeCatalogClient.Actions()
	if err := assertNumberOfActions(t, actions, 1); err != nil {
		t.Fatal(err)
	}

	updatedBinding, err := assertUpdateStatus(actions[0], binding)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBindingReadyFalse(updatedBinding); err != nil {
		t.Fatal(err)
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeWarning + " " + errorFindingNamespaceInstanceReason + " " + "Failed to get namespace \"test-ns\" during binding: No namespace"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestReconcileBindingDelete(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	bindingsMapKey := fakebrokerapi.BindingsMapKey(instanceGUID, bindingGUID)

	fakeBrokerClient.BindingClient.Bindings = map[string]*brokerapi.ServiceBinding{bindingsMapKey: {}}

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())
	sharedInformers.Instances().Informer().GetStore().Add(getTestInstance())

	binding := &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:              testBindingName,
			Namespace:         testNamespace,
			DeletionTimestamp: &metav1.Time{},
			Finalizers:        []string{"kubernetes"},
		},
		Spec: v1alpha1.BindingSpec{
			InstanceRef: v1.LocalObjectReference{Name: testInstanceName},
			OSBGUID:     bindingGUID,
			SecretName:  testBindingSecretName,
		},
	}

	fakeCatalogClient.AddReactor("get", "bindings", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, binding, nil
	})

	testController.reconcileBinding(binding)

	kubeActions := fakeKubeClient.Actions()
	// The two actions should be:
	// 0. Getting the secret
	// 1. Deleting the secret
	if err := assertNumberOfActions(t, kubeActions, 2); err != nil {
		t.Fatal(err)
	}

	getAction := kubeActions[0].(clientgotesting.GetActionImpl)
	if e, a := "get", getAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on kubeActions[0]; expected %v, got %v", e, a)
	}

	if e, a := binding.Spec.SecretName, getAction.Name; e != a {
		t.Fatalf("Unexpected name of secret: expected %v, got %v", e, a)
	}

	deleteAction := kubeActions[1].(clientgotesting.DeleteActionImpl)
	if e, a := "delete", deleteAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on kubeActions[1]; expected %v, got %v", e, a)
	}

	if e, a := binding.Spec.SecretName, deleteAction.Name; e != a {
		t.Fatalf("Unexpected name of secret: expected %v, got %v", e, a)
	}

	actions := fakeCatalogClient.Actions()
	// The three actions should be:
	// 0. Updating the ready condition
	// 1. Get against the binding in question
	// 2. Removing the finalizer
	if err := assertNumberOfActions(t, actions, 3); err != nil {
		t.Fatal(err)
	}

	updatedBinding, err := assertUpdateStatus(actions[0], binding)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBindingReadyFalse(updatedBinding); err != nil {
		t.Fatal(err)
	}

	if err := assertGet(actions[1], binding); err != nil {
		t.Fatal(err)
	}

	updatedBinding, err = assertUpdateStatus(actions[2], binding)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertEmptyFinalizers(updatedBinding); err != nil {
		t.Fatal(err)
	}

	if _, ok := fakeBrokerClient.BindingClient.Bindings[bindingsMapKey]; ok {
		t.Fatalf("Found the deleted Binding in fakeBindingClient after deletion")
	}

	events := getRecordedEvents(testController)
	if err := assertNumEvents(events, 1); err != nil {
		t.Fatal(err)
	}

	expectedEvent := api.EventTypeNormal + " " + successUnboundReason + " " + "This binding was deleted successfully"
	if e, a := expectedEvent, events[0]; e != a {
		t.Fatalf("Received unexpected event: %v", a)
	}
}

func TestUpdateBindingCondition(t *testing.T) {
	getTestBindingWithStatus := func(status v1alpha1.ConditionStatus) *v1alpha1.Binding {
		instance := getTestBinding()
		instance.Status = v1alpha1.BindingStatus{
			Conditions: []v1alpha1.BindingCondition{{
				Type:               v1alpha1.BindingConditionReady,
				Status:             status,
				LastTransitionTime: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
			}},
		}

		return instance
	}

	cases := []struct {
		name                  string
		input                 *v1alpha1.Binding
		status                v1alpha1.ConditionStatus
		transitionTimeChanged bool
	}{

		{
			name:                  "initially unset",
			input:                 getTestBinding(),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: true,
		},
		{
			name:                  "not ready -> not ready",
			input:                 getTestBindingWithStatus(v1alpha1.ConditionFalse),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: false,
		},
		{
			name:                  "not ready -> ready",
			input:                 getTestBindingWithStatus(v1alpha1.ConditionFalse),
			status:                v1alpha1.ConditionTrue,
			transitionTimeChanged: true,
		},
		{
			name:                  "ready -> ready",
			input:                 getTestBindingWithStatus(v1alpha1.ConditionTrue),
			status:                v1alpha1.ConditionTrue,
			transitionTimeChanged: false,
		},
		{
			name:                  "ready -> not ready",
			input:                 getTestBindingWithStatus(v1alpha1.ConditionTrue),
			status:                v1alpha1.ConditionFalse,
			transitionTimeChanged: true,
		},
	}

	for _, tc := range cases {
		_, fakeCatalogClient, _, testController, _ := newTestController(t)

		clone, err := api.Scheme.DeepCopy(tc.input)
		if err != nil {
			t.Errorf("%v: deep copy failed", tc.name)
			continue
		}
		inputClone := clone.(*v1alpha1.Binding)

		err = testController.updateBindingCondition(tc.input, v1alpha1.BindingConditionReady, tc.status, "reason", "message")
		if err != nil {
			t.Errorf("%v: error updating broker condition: %v", tc.name, err)
			continue
		}

		if !reflect.DeepEqual(tc.input, inputClone) {
			t.Errorf("%v: updating broker condition mutated input: expected %v, got %v", tc.name, inputClone, tc.input)
			continue
		}

		actions := fakeCatalogClient.Actions()
		if err := assertNumberOfActions(t, actions, 1); err != nil {
			t.Errorf("%v: %v", tc.name, err)
			continue
		}

		updatedBinding, err := assertUpdateStatus(actions[0], inputClone)
		if err != nil {
			t.Errorf("%v: %v", tc.name, err)
			continue
		}
		updateActionObject, ok := updatedBinding.(*v1alpha1.Binding)
		if !ok {
			t.Errorf("%v: couldn't convert to a binding", tc.name)
			continue
		}

		var initialTs metav1.Time
		if len(inputClone.Status.Conditions) != 0 {
			initialTs = inputClone.Status.Conditions[0].LastTransitionTime
		}

		newTs := updateActionObject.Status.Conditions[0].LastTransitionTime

		if tc.transitionTimeChanged && initialTs == newTs {
			t.Errorf("%v: transition time didn't change when it should have", tc.name)
			continue
		} else if !tc.transitionTimeChanged && initialTs != newTs {
			t.Errorf("%v: transition time changed when it shouldn't have", tc.name)
			continue
		}
	}
}

func TestEmptyCatalogConversion(t *testing.T) {
	serviceClasses, err := convertCatalog(&brokerapi.Catalog{})
	if err != nil {
		t.Fatalf("Failed to convertCatalog: %v", err)
	}
	if len(serviceClasses) != 0 {
		t.Fatalf("Expected 0 serviceclasses for empty catalog, but got: %d", len(serviceClasses))
	}
}

func TestCatalogConversion(t *testing.T) {
	catalog := &brokerapi.Catalog{}
	err := json.Unmarshal([]byte(testCatalog), &catalog)
	if err != nil {
		t.Fatalf("Failed to unmarshal test catalog: %v", err)
	}
	serviceClasses, err := convertCatalog(catalog)
	if err != nil {
		t.Fatalf("Failed to convertCatalog: %v", err)
	}
	if len(serviceClasses) != 1 {
		t.Fatalf("Expected 1 serviceclasses for testCatalog, but got: %d", len(serviceClasses))
	}
	serviceClass := serviceClasses[0]
	if len(serviceClass.Plans) != 2 {
		t.Fatalf("Expected 2 plans for testCatalog, but got: %d", len(serviceClass.Plans))
	}

	checkPlan(serviceClass, 0, "fake-plan-1", "Shared fake Server, 5tb persistent disk, 40 max concurrent connections", t)
	checkPlan(serviceClass, 1, "fake-plan-2", "Shared fake Server, 5tb persistent disk, 40 max concurrent connections. 100 async", t)
}

func checkPlan(serviceClass *v1alpha1.ServiceClass, index int, planName, planDescription string, t *testing.T) {
	plan := serviceClass.Plans[index]
	if plan.Name != planName {
		t.Fatalf("Expected plan %d's name to be \"%s\", but was: %s", index, planName, plan.Name)
	}
	if *plan.Description != planDescription {
		t.Fatalf("Expected plan %d's description to be \"%s\", but was: %s", index, planDescription, *plan.Description)
	}
}

func TestCatalogConversionMultipleServiceClasses(t *testing.T) {
	catalog := &brokerapi.Catalog{}
	err := json.Unmarshal([]byte(testCatalogWithMultipleServices), &catalog)
	if err != nil {
		t.Fatalf("Failed to unmarshal test catalog: %v", err)
	}

	serviceClasses, err := convertCatalog(catalog)
	if err != nil {
		t.Fatalf("Failed to convertCatalog: %v", err)
	}
	if len(serviceClasses) != 2 {
		t.Fatalf("Expected 2 serviceclasses for empty catalog, but got: %d", len(serviceClasses))
	}
	foundSvcMeta1 := false
	foundSvcMeta2 := false
	foundPlanMeta := false
	for _, sc := range serviceClasses {
		// For service1 make sure we have service level metadata with field1 = value1 as the blob
		// and for service1 plan s1plan2 we have planmeta = planvalue as the blob.
		if sc.Name == "service1" {
			if *sc.Description != "service 1 description" {
				t.Fatalf("Expected service1's description to be \"service 1 description\", but was: %s", sc.Description)
			}
			if sc.OSBMetadata != nil && len(sc.OSBMetadata.Raw) > 0 {
				m := make(map[string]string)
				if err := json.Unmarshal(sc.OSBMetadata.Raw, &m); err == nil {
					if m["field1"] == "value1" {
						foundSvcMeta1 = true
					}
				}

			}
			if len(sc.Plans) != 2 {
				t.Fatalf("Expected 2 plans for service1 but got: %d", len(sc.Plans))
			}
			for _, sp := range sc.Plans {
				if sp.Name == "s1plan2" {
					if sp.OSBMetadata != nil && len(sp.OSBMetadata.Raw) > 0 {
						m := make(map[string]string)
						if err := json.Unmarshal(sp.OSBMetadata.Raw, &m); err != nil {
							t.Fatalf("Failed to unmarshal plan metadata: %s: %v", string(sp.OSBMetadata.Raw), err)
						}
						if m["planmeta"] == "planvalue" {
							foundPlanMeta = true
						}
					}
				}
			}
		}
		// For service2 make sure we have service level metadata with three element array with elements
		// "first", "second", and "third"
		if sc.Name == "service2" {
			if *sc.Description != "service 2 description" {
				t.Fatalf("Expected service2's description to be \"service 2 description\", but was: %s", sc.Description)
			}
			if sc.OSBMetadata != nil && len(sc.OSBMetadata.Raw) > 0 {
				m := make([]string, 0)
				if err := json.Unmarshal(sc.OSBMetadata.Raw, &m); err != nil {
					t.Fatalf("Failed to unmarshal service metadata: %s: %v", string(sc.OSBMetadata.Raw), err)
				}
				if len(m) != 3 {
					t.Fatalf("Expected 3 fields in metadata, but got %d", len(m))
				}
				foundFirst := false
				foundSecond := false
				foundThird := false
				for _, e := range m {
					if e == "first" {
						foundFirst = true
					}
					if e == "second" {
						foundSecond = true
					}
					if e == "third" {
						foundThird = true
					}
				}
				if !foundFirst {
					t.Fatalf("Didn't find 'first' in plan metadata")
				}
				if !foundSecond {
					t.Fatalf("Didn't find 'second' in plan metadata")
				}
				if !foundThird {
					t.Fatalf("Didn't find 'third' in plan metadata")
				}
				foundSvcMeta2 = true
			}
		}
	}
	if !foundSvcMeta1 {
		t.Fatalf("Didn't find metadata in service1")
	}
	if !foundSvcMeta2 {
		t.Fatalf("Didn't find metadata in service2")
	}
	if !foundPlanMeta {
		t.Fatalf("Didn't find metadata '' in service1 plan2")
	}

}

// newTestController creates a new test controller injected with fake clients
// and returns:
//
// - a fake kubernetes core api client
// - a fake service catalog api client
// - a fake broker catalog client
// - a fake broker instance client
// - a fake broker binding client
// - a test controller
// - the shared informers for the service catalog v1alpha1 api
//
// If there is an error, newTestController calls 'Fatal' on the injected
// testing.T.
func newTestController(t *testing.T) (
	*clientgofake.Clientset,
	*servicecatalogclientset.Clientset,
	*fakebrokerapi.Client,
	*controller,
	v1alpha1informers.Interface) {
	// create a fake kube client
	fakeKubeClient := &clientgofake.Clientset{}
	// create a fake sc client
	fakeCatalogClient := &servicecatalogclientset.Clientset{}

	catalogCl := &fakebrokerapi.CatalogClient{}
	instanceCl := fakebrokerapi.NewInstanceClient()
	bindingCl := fakebrokerapi.NewBindingClient()
	fakeBrokerClient := &fakebrokerapi.Client{
		CatalogClient:  catalogCl,
		InstanceClient: instanceCl,
		BindingClient:  bindingCl,
	}

	brokerClFunc := fakebrokerapi.NewClientFunc(catalogCl, instanceCl, bindingCl)

	// create informers
	informerFactory := servicecataloginformers.NewSharedInformerFactory(fakeCatalogClient, 0)
	serviceCatalogSharedInformers := informerFactory.Servicecatalog().V1alpha1()

	fakeRecorder := record.NewFakeRecorder(5)

	// create a test controller
	testController, err := NewController(
		fakeKubeClient,
		fakeCatalogClient.ServicecatalogV1alpha1(),
		serviceCatalogSharedInformers.Brokers(),
		serviceCatalogSharedInformers.ServiceClasses(),
		serviceCatalogSharedInformers.Instances(),
		serviceCatalogSharedInformers.Bindings(),
		brokerClFunc,
		24*time.Hour,
		true, /* enable OSB context profile */
		fakeRecorder,
	)
	if err != nil {
		t.Fatal(err)
	}

	return fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController.(*controller), serviceCatalogSharedInformers
}

func getRecordedEvents(testController *controller) []string {
	source := testController.recorder.(*record.FakeRecorder).Events
	done := false
	events := []string{}
	for !done {
		select {
		case event := <-source:
			events = append(events, event)
		default:
			done = true
		}
	}
	return events
}

func assertNumEvents(strings []string, number int) error {
	if e, a := number, len(strings); e != a {
		return fmt.Errorf("Unexpected number of events: expected %v, got %v", e, a)
	}

	return nil
}

func assertNumberOfActions(t *testing.T, actions []clientgotesting.Action, number int) error {
	if e, a := number, len(actions); e != a {
		t.Logf("actions: %+v\n", actions)
		return fmt.Errorf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	return nil
}

func assertGet(action clientgotesting.Action, obj interface{}) error {
	_, err := assertActionFor(action, "get", "" /* subresource */, obj)
	return err
}

func assertCreate(action clientgotesting.Action, obj interface{}) (runtime.Object, error) {
	return assertActionFor(action, "create", "" /* subresource */, obj)
}

func assertUpdate(action clientgotesting.Action, obj interface{}) (runtime.Object, error) {
	return assertActionFor(action, "update", "" /* subresource */, obj)
}

func assertUpdateStatus(action clientgotesting.Action, obj interface{}) (runtime.Object, error) {
	return assertActionFor(action, "update", "status", obj)
}

func assertDelete(action clientgotesting.Action, obj interface{}) error {
	_, err := assertActionFor(action, "delete", "" /* subresource */, obj)
	return err
}

// assertActionFor makes an assertion that the given action is for the given
// verb and subresource (if provided) for the resource of the object's type,
// with the name of the given object. It returns an error if one of these
// assertions fails, and returns the runtime.Object involved in the action (if
// possible).
func assertActionFor(action clientgotesting.Action, verb, subresource string, obj interface{}) (runtime.Object, error) {
	if e, a := verb, action.GetVerb(); e != a {
		return nil, fmt.Errorf("Unexpected verb: expected %v, got %v", e, a)
	}

	var resource string

	switch obj.(type) {
	case *v1alpha1.Broker:
		resource = "brokers"
	case *v1alpha1.ServiceClass:
		resource = "serviceclasses"
	case *v1alpha1.Instance:
		resource = "instances"
	case *v1alpha1.Binding:
		resource = "bindings"
	}

	if e, a := resource, action.GetResource().Resource; e != a {
		return nil, fmt.Errorf("Unexpected resource; expected %v, got %v", e, a)
	}

	if e, a := subresource, action.GetSubresource(); e != a {
		return nil, fmt.Errorf("Unexpected subresource; expected %v, got %v", e, a)
	}

	rtObject, ok := obj.(runtime.Object)
	if !ok {
		return nil, fmt.Errorf("Object %+v was not a runtime.Object", obj)
	}

	paramAccessor, err := metav1.ObjectMetaFor(rtObject)
	if err != nil {
		return nil, fmt.Errorf("Error creating ObjectMetaAccessor for param object %+v: %v", rtObject, err)
	}

	var (
		objectMeta   metav1.Object
		fakeRtObject runtime.Object
	)

	switch verb {
	case "get":
		getAction, ok := action.(clientgotesting.GetAction)
		if !ok {
			return nil, fmt.Errorf("Unexpected type; failed to convert action %+v to DeleteAction", action)
		}

		if e, a := paramAccessor.GetName(), getAction.GetName(); e != a {
			return nil, fmt.Errorf("unexpected name: expected %v, got %v", e, a)
		}

		return nil, nil
	case "delete":
		deleteAction, ok := action.(clientgotesting.DeleteAction)
		if !ok {
			return nil, fmt.Errorf("Unexpected type; failed to convert action %+v to DeleteAction", action)
		}

		if e, a := paramAccessor.GetName(), deleteAction.GetName(); e != a {
			return nil, fmt.Errorf("unexpected name: expected %v, got %v", e, a)
		}

		return nil, nil
	case "create":
		createAction, ok := action.(clientgotesting.CreateAction)
		if !ok {
			return nil, fmt.Errorf("Unexpected type; failed to convert action %+v to CreateAction", action)
		}

		fakeRtObject = createAction.GetObject()
		objectMeta, err = metav1.ObjectMetaFor(fakeRtObject)
		if err != nil {
			return nil, fmt.Errorf("Error creating ObjectMetaAccessor for %+v", fakeRtObject)
		}
	case "update":
		updateAction, ok := action.(clientgotesting.UpdateAction)
		if !ok {
			return nil, fmt.Errorf("Unexpected type; failed to convert action %+v to UpdateAction", action)
		}

		fakeRtObject = updateAction.GetObject()
		objectMeta, err = metav1.ObjectMetaFor(fakeRtObject)
		if err != nil {
			return nil, fmt.Errorf("Error creating ObjectMetaAccessor for %+v", fakeRtObject)
		}
	}

	if e, a := paramAccessor.GetName(), objectMeta.GetName(); e != a {
		return nil, fmt.Errorf("unexpected name: expected %v, got %v", e, a)
	}

	fakeValue := reflect.ValueOf(fakeRtObject)
	paramValue := reflect.ValueOf(obj)

	if e, a := paramValue.Type(), fakeValue.Type(); e != a {
		return nil, fmt.Errorf("Unexpected type of object passed to fake client; expected %v, got %v", e, a)
	}

	return fakeRtObject, nil
}

func assertBrokerReadyTrue(obj runtime.Object) error {
	return assertBrokerReadyCondition(obj, v1alpha1.ConditionTrue)
}

func assertBrokerReadyFalse(obj runtime.Object) error {
	return assertBrokerReadyCondition(obj, v1alpha1.ConditionFalse)
}

func assertBrokerReadyCondition(obj runtime.Object, status v1alpha1.ConditionStatus) error {
	broker, ok := obj.(*v1alpha1.Broker)
	if !ok {
		return fmt.Errorf("Couldn't convert object %+v into a *v1alpha1.Broker", obj)
	}

	for _, condition := range broker.Status.Conditions {
		if condition.Type == v1alpha1.BrokerConditionReady && condition.Status != status {
			return fmt.Errorf("ready condition had unexpected status; expected %v, got %v", status, condition.Status)
		}
	}

	return nil
}

func assertInstanceReadyTrue(obj runtime.Object) error {
	return assertInstanceReadyCondition(obj, v1alpha1.ConditionTrue)
}

func assertInstanceReadyFalse(obj runtime.Object, reason ...string) error {
	return assertInstanceReadyCondition(obj, v1alpha1.ConditionFalse, reason...)
}

func assertInstanceReadyCondition(obj runtime.Object, status v1alpha1.ConditionStatus, reason ...string) error {
	instance, ok := obj.(*v1alpha1.Instance)
	if !ok {
		return fmt.Errorf("Couldn't convert object %+v into a *v1alpha1.Instance", obj)
	}

	for _, condition := range instance.Status.Conditions {
		if condition.Type == v1alpha1.InstanceConditionReady && condition.Status != status {
			return fmt.Errorf("ready condition had unexpected status; expected %v, got %v", status, condition.Status)
		}
		if len(reason) == 1 && condition.Reason != reason[0] {
			return fmt.Errorf("unexpected reason; expected %v, got %v", reason[0], condition.Reason)
		}
	}

	return nil
}

func assertBindingReadyTrue(obj runtime.Object) error {
	return assertBindingReadyCondition(obj, v1alpha1.ConditionTrue)
}

func assertBindingReadyFalse(obj runtime.Object, reason ...string) error {
	return assertBindingReadyCondition(obj, v1alpha1.ConditionFalse, reason...)
}

func assertBindingReadyCondition(obj runtime.Object, status v1alpha1.ConditionStatus, reason ...string) error {
	binding, ok := obj.(*v1alpha1.Binding)
	if !ok {
		return fmt.Errorf("Couldn't convert object %+v into a *v1alpha1.Binding", obj)
	}

	for _, condition := range binding.Status.Conditions {
		if condition.Type == v1alpha1.BindingConditionReady && condition.Status != status {
			return fmt.Errorf("ready condition had unexpected status; expected %v, got %v", status, condition.Status)
		}
		if len(reason) == 1 && condition.Reason != reason[0] {
			return fmt.Errorf("unexpected reason; expected %v, got %v", reason[0], condition.Reason)
		}
	}

	return nil
}

func assertEmptyFinalizers(obj runtime.Object) error {
	accessor, err := metav1.ObjectMetaFor(obj)
	if err != nil {
		return fmt.Errorf("Error creating ObjectMetaAccessor for param object %+v: %v", obj, err)
	}

	if len(accessor.GetFinalizers()) != 0 {
		return fmt.Errorf("Unexpected number of finalizers; expected 0, got %v", len(accessor.GetFinalizers()))
	}

	return nil
}
