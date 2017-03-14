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
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	fakebrokerapi "github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake"
	servicecataloginformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated"
	v1alpha1informers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/servicecatalog/v1alpha1"

	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"k8s.io/kubernetes/pkg/api/v1"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/runtime"

	clientgofake "k8s.io/client-go/1.5/kubernetes/fake"
	clientgoruntime "k8s.io/client-go/1.5/pkg/runtime"
	clientgotesting "k8s.io/client-go/1.5/testing"
)

// TLDR
// For the time being, everything related to verifying actions on the k8s core
// client should use the clientgo* packages. Everything related to verifying
// expectations on the service catalog API fake should use the
// k8s.io/kubernetes packages.
//
// NOTE:
//
// There are two different 'testing' packages imported here from kubernetes
// projects:
//
// - k8s.io/kubernetes/client/testing/core
// - k8s.io/client-go/1.5/testing
//
// These are the same package, but we have to import them from both locations
// because our API is written using packages from kubernetes directly, while
// for the kubernetes core API, we use client-go.
//
// We _have_ to do this for now because the version of kubernetes we vendor in
// for the API server guts uses the kubernetes packages.  Once we rebase onto
// the latest kubernetes repo, we'll be able to stop using the kubernetes/...
// packages entirely and _just_ consume types from client-go.
//
// See issue: https://github.com/kubernetes-incubator/service-catalog/issues/413

const (
	serviceClassGUID = "SCGUID"
	planGUID         = "PGUID"
	instanceGUID     = "IGUID"
	bindingGUID      = "BGUID"

	testBrokerName       = "test-broker"
	testServiceClassName = "test-serviceclass"
	testPlanName         = "test-plan"
	testInstanceName     = "test-instance"
	testBindingName      = "test-binding"
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
		ObjectMeta: v1.ObjectMeta{Name: testBrokerName},
		Spec: v1alpha1.BrokerSpec{
			URL: "https://example.com",
		},
	}
}

func getTestServiceClass() *v1alpha1.ServiceClass {
	return &v1alpha1.ServiceClass{
		ObjectMeta: v1.ObjectMeta{Name: testServiceClassName},
		BrokerName: testBrokerName,
		Plans: []v1alpha1.ServicePlan{{
			Name:    testPlanName,
			OSBFree: true,
			OSBGUID: planGUID,
		}},
	}
}

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

type instanceParameters struct {
	Name string            `json:"name"`
	Args map[string]string `json:"args"`
}

type bindingParameters struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

func TestReconcileBroker(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, _ := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	testController.reconcileBroker(getTestBroker())

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 2, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}

	// first action should be a create action for a service class
	createAction := actions[0].(core.CreateAction)
	if e, a := "create", createAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}

	createActionObject := createAction.GetObject().(*v1alpha1.ServiceClass)
	if e, a := testServiceClassName, createActionObject.Name; e != a {
		t.Fatalf("Unexpected name of serviceClass created: expected %v, got %v", e, a)
	}

	// second action should be an update action for broker status subresource
	createAction2 := actions[1].(core.CreateAction)
	if e, a := "update", createAction2.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}

	createActionObject2 := createAction2.GetObject().(*v1alpha1.Broker)
	if e, a := testBrokerName, createActionObject2.Name; e != a {
		t.Fatalf("Unexpected name of broker created: expected %v, got %v", e, a)
	}

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	if e, a := 0, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
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
	if e, a := 0, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	actions := filterActions(fakeCatalogClient.Actions())
	// The three actions should be:
	// 0. Deleting the associated ServiceClass
	// 1. Updating the ready condition
	// 2. Removing the finalizer
	if e, a := 3, len(actions); e != a {
		t.Logf("%+v\n", actions)
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	deleteAction := actions[0].(core.DeleteActionImpl)
	if e, a := "delete", deleteAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}

	if e, a := testServiceClass.Name, deleteAction.Name; e != a {
		t.Fatalf("Unexpected name of serviceclass: expected %v, got %v", e, a)
	}

	updateAction := actions[1].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}

	updatedBroker := updateAction.GetObject().(*v1alpha1.Broker)
	if e, a := broker.Name, updatedBroker.Name; e != a {
		t.Fatalf("Unexpected name of broker: expected %v, got %v", e, a)
	}

	if e, a := 1, len(updatedBroker.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of status conditions: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.BrokerConditionReady, updatedBroker.Status.Conditions[0].Type; e != a {
		t.Fatalf("Unexpected condition type: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.ConditionFalse, updatedBroker.Status.Conditions[0].Status; e != a {
		t.Fatalf("Unexpected condition status: expected %v, got %v", e, a)
	}

	updateAction = actions[2].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[2]; expected %v, got %v", e, a)
	}

	updatedBroker = updateAction.GetObject().(*v1alpha1.Broker)
	if e, a := broker.Name, updatedBroker.Name; e != a {
		t.Fatalf("Unexpected name of broker: expected %v, got %v", e, a)
	}

	if e, a := 0, len(updatedBroker.Finalizers); e != a {
		t.Fatalf("Unexpected number of finalizers: expected %v, got %v", e, a)
	}
}

func TestReconcileBrokerErrorFetchingCatalog(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, _ := newTestController(t)

	fakeBrokerClient.CatalogClient.RetErr = fakebrokerapi.ErrInstanceNotFound
	broker := getTestBroker()

	testController.reconcileBroker(broker)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}
	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}
	updateObject := updateAction.GetObject().(*v1alpha1.Broker)
	if e, a := broker.Name, updateObject.Name; e != a {
		t.Fatalf("Unexpected name of broker: expected %v, got %v", e, a)
	}
	if e, a := v1alpha1.ConditionFalse, updateObject.Status.Conditions[0].Status; e != a {
		t.Fatalf("Unexpected condition status: expected %v, got %v", e, a)
	}

	kubeActions := fakeKubeClient.Actions()
	if e, a := 0, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}
}

func TestReconcileBrokerWithAuthError(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, _, testController, _ := newTestController(t)

	broker := getTestBroker()
	broker.Spec.AuthSecret = &v1.ObjectReference{
		Namespace: "does_not_exist",
		Name:      "auth-name",
	}

	fakeKubeClient.AddReactor("get", "secrets", func(action clientgotesting.Action) (bool, clientgoruntime.Object, error) {
		return true, nil, errors.New("no secret defined")
	})

	testController.reconcileBroker(broker)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}
	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on action; expected %v, got %v", e, a)
	}
	updateObject := updateAction.GetObject().(*v1alpha1.Broker)
	if e, a := broker.Name, updateObject.Name; e != a {
		t.Fatalf("Unexpected name of broker: expected %v, got %v", e, a)
	}
	if e, a := v1alpha1.ConditionFalse, updateObject.Status.Conditions[0].Status; e != a {
		t.Fatalf("Unexpected condition status: expected %v, got %v", e, a)
	}

	// verify one kube action occurred
	kubeActions := fakeKubeClient.Actions()
	if e, a := 1, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}
	getAction := kubeActions[0].(clientgotesting.GetAction)
	if e, a := "get", getAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on action; expected %v, got %v", e, a)
	}
	if e, a := "secrets", getAction.GetResource().Resource; e != a {
		t.Fatalf("Unexpected resource on action; expected %v, got %v", e, a)
	}
}

func TestReconcileBrokerWithReconcileError(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, _, testController, _ := newTestController(t)

	broker := getTestBroker()
	broker.Spec.AuthSecret = &v1.ObjectReference{
		Namespace: "does_not_exist",
		Name:      "auth-name",
	}

	fakeCatalogClient.AddReactor("create", "serviceclasses", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("error creating serviceclass")
	})

	testController.reconcileBroker(broker)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}
	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}
	updateObject := updateAction.GetObject().(*v1alpha1.Broker)
	if e, a := broker.Name, updateObject.Name; e != a {
		t.Fatalf("Unexpected name of broker: expected %v, got %v", e, a)
	}
	if e, a := v1alpha1.ConditionFalse, updateObject.Status.Conditions[0].Status; e != a {
		t.Fatalf("Unexpected condition status: expected %v, got %v", e, a)
	}

	kubeActions := fakeKubeClient.Actions()
	if e, a := 1, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}
	getAction := kubeActions[0].(clientgotesting.GetAction)
	if e, a := "get", getAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on action; expected %v, got %v", e, a)
	}
}

func TestReconcileInstanceNonExistentServiceClass(t *testing.T) {
	_, fakeCatalogClient, _, testController, _ := newTestController(t)

	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: testInstanceName},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: "nothere",
			PlanName:         "nothere",
			OSBGUID:          instanceGUID,
		},
	}

	testController.reconcileInstance(instance)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}

	// There should only be one action that says it failed because no such class exists.
	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}
	updateActionObject := updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := testInstanceName, updateActionObject.Name; e != a {
		t.Fatalf("Unexpected name of instance created: expected %v, got %v", e, a)
	}
	if e, a := 1, len(updateActionObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of conditions: expected %v, got %v", e, a)
	}
	if e, a := "ReferencesNonexistentServiceClass", updateActionObject.Status.Conditions[0].Reason; e != a {
		t.Fatalf("Unexpected condition reason: expected %v, got %v", e, a)
	}
}

func TestReconcileInstanceNonExistentBroker(t *testing.T) {
	_, fakeCatalogClient, _, testController, sharedInformers := newTestController(t)

	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: testInstanceName},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}

	testController.reconcileInstance(instance)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: +%v", e, a, actions)
	}

	// There should only be one action that says it failed because no such broker exists.
	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}
	updateActionObject := updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := testInstanceName, updateActionObject.Name; e != a {
		t.Fatalf("Unexpected name of instance created: expected %v, got %v", e, a)
	}
	if e, a := 1, len(updateActionObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of conditions: expected %v, got %v", e, a)
	}
	if e, a := "ReferencesNonexistentBroker", updateActionObject.Status.Conditions[0].Reason; e != a {
		t.Fatalf("Unexpected condition reason: expected %v, got %v", e, a)
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

	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: testInstanceName},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}

	fakeKubeClient.AddReactor("get", "secrets", func(action clientgotesting.Action) (bool, clientgoruntime.Object, error) {
		return true, nil, errors.New("no secret defined")
	})

	testController.reconcileInstance(instance)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: +%v", e, a, actions)
	}

	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on action; expected %v, got %v", e, a)
	}
	updateActionObject := updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := testInstanceName, updateActionObject.Name; e != a {
		t.Fatalf("Unexpected name of instance created: expected %v, got %v", e, a)
	}
	if e, a := 1, len(updateActionObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of conditions: expected %v, got %v", e, a)
	}
	if e, a := "ErrorGettingAuthCredentials", updateActionObject.Status.Conditions[0].Reason; e != a {
		t.Fatalf("Unexpected condition reason: expected %v, got %v", e, a)
	}

	// verify one kube action occurred
	kubeActions := fakeKubeClient.Actions()
	if e, a := 1, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}
	getAction := kubeActions[0].(clientgotesting.GetAction)
	if e, a := "get", getAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on action; expected %v, got %v", e, a)
	}
	if e, a := "secrets", getAction.GetResource().Resource; e != a {
		t.Fatalf("Unexpected resource on action; expected %v, got %v", e, a)
	}

}

func TestReconcileInstanceNonExistentServicePlan(t *testing.T) {
	_, fakeCatalogClient, _, testController, sharedInformers := newTestController(t)

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: testInstanceName},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         "nothere",
			OSBGUID:          instanceGUID,
		},
	}

	testController.reconcileInstance(instance)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}

	// There should only be one action that says it failed because no such class exists.
	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}
	updateActionObject := updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := testInstanceName, updateActionObject.Name; e != a {
		t.Fatalf("Unexpected name of instance created: expected %v, got %v", e, a)
	}
	if e, a := 1, len(updateActionObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of conditions: expected %v, got %v", e, a)
	}
	if e, a := "ReferencesNonexistentServicePlan", updateActionObject.Status.Conditions[0].Reason; e != a {
		t.Fatalf("Unexpected condition reason: expected %v, got %v", e, a)
	}
}

func TestReconcileInstanceWithParameters(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: testInstanceName, Namespace: "test-ns"},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}
	parameters := instanceParameters{Name: "test-param", Args: make(map[string]string)}
	parameters.Args["first"] = "first-arg"
	parameters.Args["second"] = "second-arg"

	b, err := json.Marshal(parameters)
	if err != nil {
		t.Fatalf("Failed to marshal parameters %v : %v", parameters, err)
	}
	instance.Spec.Parameters = &runtime.RawExtension{Raw: b}

	testController.reconcileInstance(instance)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	if e, a := 0, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}

	updateObject := updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := instance.Name, updateObject.Name; e != a {
		t.Fatalf("Unexpected name of serviceClass created: expected %v, got %v", e, a)
	}

	if e, a := 1, len(updateObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of status conditions: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.InstanceConditionReady, updateObject.Status.Conditions[0].Type; e != a {
		t.Fatalf("Unexpected condition type: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.ConditionTrue, updateObject.Status.Conditions[0].Status; e != a {
		t.Fatalf("Unexpected condition status: expected %v, got %v", e, a)
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
}

func TestReconcileInstanceWithInvalidParameters(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: testInstanceName, Namespace: "test-ns"},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}
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

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	if e, a := 0, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}

	updateObject := updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := instance.Name, updateObject.Name; e != a {
		t.Fatalf("Unexpected name of instance created: expected %v, got %v", e, a)
	}

	if e, a := 1, len(updateObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of status conditions: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.InstanceConditionReady, updateObject.Status.Conditions[0].Type; e != a {
		t.Fatalf("Unexpected condition type: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.ConditionFalse, updateObject.Status.Conditions[0].Status; e != a {
		t.Fatalf("Unexpected condition status: expected %v, got %v", e, a)
	}

	if si, notOK := fakeBrokerClient.InstanceClient.Instances[instanceGUID]; notOK {
		t.Fatalf("Unexpectedly found created Instance: %+v in fakeInstanceClient after creation", si)
	}
}

func TestReconcileInstanceWithInstanceError(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: testInstanceName, Namespace: "test-ns"},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}
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
	kubeActions := fakeKubeClient.Actions()
	if e, a := 0, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}
	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on action; expected %v, got %v", e, a)
	}

	updateObject := updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := instance.Name, updateObject.Name; e != a {
		t.Fatalf("Unexpected name of instance created: expected %v, got %v", e, a)
	}

	if e, a := 1, len(updateObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of status conditions: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.InstanceConditionReady, updateObject.Status.Conditions[0].Type; e != a {
		t.Fatalf("Unexpected condition type: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.ConditionFalse, updateObject.Status.Conditions[0].Status; e != a {
		t.Fatalf("Unexpected condition status: expected %v, got %v", e, a)
	}

	if si, notOK := fakeBrokerClient.InstanceClient.Instances[instanceGUID]; notOK {
		t.Fatalf("Unexpectedly found created Instance: %+v in fakeInstanceClient after creation", si)
	}
}

func TestReconcileInstance(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: testInstanceName, Namespace: "test-ns"},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}

	testController.reconcileInstance(instance)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	if e, a := 0, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}

	updateObject := updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := instance.Name, updateObject.Name; e != a {
		t.Fatalf("Unexpected name of serviceClass created: expected %v, got %v", e, a)
	}

	if e, a := 1, len(updateObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of status conditions: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.InstanceConditionReady, updateObject.Status.Conditions[0].Type; e != a {
		t.Fatalf("Unexpected condition type: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.ConditionTrue, updateObject.Status.Conditions[0].Status; e != a {
		t.Fatalf("Unexpected condition status: expected %v, got %v", e, a)
	}

	if si, ok := fakeBrokerClient.InstanceClient.Instances[instanceGUID]; !ok {
		t.Fatalf("Did not find the created Instance in fakeInstanceClient after creation")
	} else {
		if len(si.Parameters) > 0 {
			t.Fatalf("Unexpected parameters, expected none, got %+v", si.Parameters)
		}
	}
}

func TestReconcileInstanceDelete(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.InstanceClient.Instances = map[string]*brokerapi.ServiceInstance{
		instanceGUID: {},
	}

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{
			Name:              testInstanceName,
			Namespace:         "test-ns",
			DeletionTimestamp: &metav1.Time{},
			Finalizers:        []string{"kubernetes"},
		},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}

	testController.reconcileInstance(instance)

	// Verify no core kube actions occurred
	kubeActions := fakeKubeClient.Actions()
	if e, a := 0, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	actions := filterActions(fakeCatalogClient.Actions())
	// The two actions should be:
	// 0. Updating the ready condition
	// 1. Removing the finalizer
	if e, a := 2, len(actions); e != a {
		t.Logf("%+v\n", actions)
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}

	updatedObject := updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := instance.Name, updatedObject.Name; e != a {
		t.Fatalf("Unexpected name of instance: expected %v, got %v", e, a)
	}

	if e, a := 1, len(updatedObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of status conditions: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.InstanceConditionReady, updatedObject.Status.Conditions[0].Type; e != a {
		t.Fatalf("Unexpected condition type: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.ConditionFalse, updatedObject.Status.Conditions[0].Status; e != a {
		t.Fatalf("Unexpected condition status: expected %v, got %v", e, a)
	}

	if _, ok := fakeBrokerClient.InstanceClient.Instances[instanceGUID]; ok {
		t.Fatalf("Found the deleted Instance in fakeInstanceClient after deletion")
	}

	updateAction = actions[1].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}

	updatedObject = updateAction.GetObject().(*v1alpha1.Instance)
	if e, a := instance.Name, updatedObject.Name; e != a {
		t.Fatalf("Unexpected name of instance: expected %v, got %v", e, a)
	}

	if e, a := 0, len(updatedObject.Finalizers); e != a {
		t.Fatalf("Unexpected number of finalizers: expected %v, got %v", e, a)
	}
}

func TestReconcileBindingNonExistingInstance(t *testing.T) {
	_, fakeCatalogClient, _, testController, _ := newTestController(t)

	binding := &v1alpha1.Binding{
		ObjectMeta: v1.ObjectMeta{Name: testBindingName},
		Spec: v1alpha1.BindingSpec{
			InstanceRef: v1.LocalObjectReference{Name: "nothere"},
			OSBGUID:     bindingGUID,
		},
	}

	testController.reconcileBinding(binding)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}

	// There should only be one action that says it failed because no such instance exists.
	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}
	updateActionObject := updateAction.GetObject().(*v1alpha1.Binding)
	if e, a := testBindingName, updateActionObject.Name; e != a {
		t.Fatalf("Unexpected name of binding created: expected %v, got %v", e, a)
	}
	if e, a := 1, len(updateActionObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of conditions: expected %v, got %v", e, a)
	}
	if e, a := "ReferencesNonexistentInstance", updateActionObject.Status.Conditions[0].Reason; e != a {
		t.Fatalf("Unexpected condition reason: expected %v, got %v", e, a)
	}
}

func TestReconcileBindingNonExistingServiceClass(t *testing.T) {
	_, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())
	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: testInstanceName, Namespace: "test-ns"},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: "nothere",
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}
	sharedInformers.Instances().Informer().GetStore().Add(instance)

	binding := &v1alpha1.Binding{
		ObjectMeta: v1.ObjectMeta{Name: testBindingName, Namespace: "test-ns"},
		Spec: v1alpha1.BindingSpec{
			InstanceRef: v1.LocalObjectReference{Name: testInstanceName},
			OSBGUID:     bindingGUID,
		},
	}

	testController.reconcileBinding(binding)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}

	// There should only be one action that says it failed because no such service class.
	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}
	updateActionObject := updateAction.GetObject().(*v1alpha1.Binding)
	if e, a := testBindingName, updateActionObject.Name; e != a {
		t.Fatalf("Unexpected name of binding created: expected %v, got %v", e, a)
	}
	if e, a := 1, len(updateActionObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of conditions: expected %v, got %v", e, a)
	}
	if e, a := "ReferencesNonexistentServiceClass", updateActionObject.Status.Conditions[0].Reason; e != a {
		t.Fatalf("Unexpected condition reason: expected %v, got %v", e, a)
	}
}

func TestReconcileBindingWithParameters(t *testing.T) {
	_, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	fakeBrokerClient.CatalogClient.RetCatalog = getTestCatalog()

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())
	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: testInstanceName, Namespace: "test-ns"},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}
	sharedInformers.Instances().Informer().GetStore().Add(instance)

	binding := &v1alpha1.Binding{
		ObjectMeta: v1.ObjectMeta{Name: testBindingName, Namespace: "test-ns"},
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

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 1, len(actions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v. Actions: %+v", e, a, actions)
	}

	// There should only be one action that says binding was created
	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}
	updateObject := updateAction.GetObject().(*v1alpha1.Binding)
	if e, a := testBindingName, updateObject.Name; e != a {
		t.Fatalf("Unexpected name of binding created: expected %v, got %v", e, a)
	}
	if e, a := 1, len(updateObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of conditions: expected %v, got %v", e, a)
	}
	if e, a := "InjectedBindResult", updateObject.Status.Conditions[0].Reason; e != a {
		t.Fatalf("Unexpected condition reason: expected %v, got %v", e, a)
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

}

func TestReconcileBindingDelete(t *testing.T) {
	fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController, sharedInformers := newTestController(t)

	bindingsMapKey := fakebrokerapi.BindingsMapKey(instanceGUID, bindingGUID)

	fakeBrokerClient.BindingClient.Bindings = map[string]*brokerapi.ServiceBinding{bindingsMapKey: {}}

	sharedInformers.Brokers().Informer().GetStore().Add(getTestBroker())
	sharedInformers.ServiceClasses().Informer().GetStore().Add(getTestServiceClass())

	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{
			Name:      testInstanceName,
			Namespace: "test-ns",
		},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: testServiceClassName,
			PlanName:         testPlanName,
			OSBGUID:          instanceGUID,
		},
	}

	sharedInformers.Instances().Informer().GetStore().Add(instance)

	binding := &v1alpha1.Binding{
		ObjectMeta: v1.ObjectMeta{
			Name:              testBindingName,
			Namespace:         "test-ns",
			DeletionTimestamp: &metav1.Time{},
			Finalizers:        []string{"kubernetes"},
		},
		Spec: v1alpha1.BindingSpec{
			InstanceRef: v1.LocalObjectReference{Name: testInstanceName},
			OSBGUID:     bindingGUID,
			SecretName:  "test-secret",
		},
	}

	testController.reconcileBinding(binding)

	kubeActions := fakeKubeClient.Actions()
	// The two actions should be:
	// 0. Getting the secret
	// 1. Deleting the secret
	if e, a := 2, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
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

	actions := filterActions(fakeCatalogClient.Actions())
	// The two actions should be:
	// 0. Updating the ready condition
	// 1. Removing the finalizer
	if e, a := 2, len(actions); e != a {
		t.Logf("%+v\n", actions)
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	updateAction := actions[0].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}

	updatedObject := updateAction.GetObject().(*v1alpha1.Binding)
	if e, a := binding.Name, updatedObject.Name; e != a {
		t.Fatalf("Unexpected name of binding: expected %v, got %v", e, a)
	}

	if e, a := 1, len(updatedObject.Status.Conditions); e != a {
		t.Fatalf("Unexpected number of status conditions: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.BindingConditionReady, updatedObject.Status.Conditions[0].Type; e != a {
		t.Fatalf("Unexpected condition type: expected %v, got %v", e, a)
	}

	if e, a := v1alpha1.ConditionFalse, updatedObject.Status.Conditions[0].Status; e != a {
		t.Fatalf("Unexpected condition status: expected %v, got %v", e, a)
	}

	if _, ok := fakeBrokerClient.BindingClient.Bindings[bindingsMapKey]; ok {
		t.Fatalf("Found the deleted Binding in fakeBindingClient after deletion")
	}

	updateAction = actions[1].(core.UpdateAction)
	if e, a := "update", updateAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}

	updatedObject = updateAction.GetObject().(*v1alpha1.Binding)
	if e, a := binding.Name, updatedObject.Name; e != a {
		t.Fatalf("Unexpected name of binding: expected %v, got %v", e, a)
	}

	if e, a := 0, len(updatedObject.Finalizers); e != a {
		t.Fatalf("Unexpected number of finalizers: expected %v, got %v", e, a)
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
		t.Fatalf("Expected 1 serviceclasses for empty catalog, but got: %d", len(serviceClasses))
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
	informerFactory := servicecataloginformers.NewSharedInformerFactory(nil, fakeCatalogClient, 0)
	serviceCatalogSharedInformers := informerFactory.Servicecatalog().V1alpha1()

	// create a test controller
	testController, err := NewController(
		fakeKubeClient,
		fakeCatalogClient.ServicecatalogV1alpha1(),
		serviceCatalogSharedInformers.Brokers(),
		serviceCatalogSharedInformers.ServiceClasses(),
		serviceCatalogSharedInformers.Instances(),
		serviceCatalogSharedInformers.Bindings(),
		brokerClFunc,
	)
	if err != nil {
		t.Fatal(err)
	}

	return fakeKubeClient, fakeCatalogClient, fakeBrokerClient, testController.(*controller), serviceCatalogSharedInformers
}

// filterActions filters the list/watch actions on service catalog resources
// from an array of core.Action.  This is so that we can write tests without
// worrying about the list/watching that the informer infrastructure might
// have done.
func filterActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "brokers") ||
				action.Matches("list", "serviceclasses") ||
				action.Matches("list", "instances") ||
				action.Matches("list", "bindings") ||
				action.Matches("watch", "brokers") ||
				action.Matches("watch", "serviceclasses") ||
				action.Matches("watch", "instances") ||
				action.Matches("watch", "bindings")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}
