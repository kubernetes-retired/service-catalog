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
	"errors"
	"net/http"
	"testing"

	faketypes "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	fakebrokerapi "github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake"
	fakebrokerserver "github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api/v1"
	clientgotesting "k8s.io/client-go/testing"
)

// TestReconcileInstanceAsynchronousUnsupportedBrokerError tests to ensure that, on an asynchronous
// provision, an Instance's conditions get set with a Broker failure that is not one of the
// "expected" response codes in the OSB API spec for provision.
// See https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#response-2 for
// the list of expected codes and a description of what we should do if another code is returned
func TestReconcileInstanceAsynchronousUnsupportedBrokerError(t *testing.T) {
	const (
		brokerUsername = "testbrokeruser"
		brokerPassword = "testbrokerpass"
	)
	controllerItems, err := controller.NewTestControllerWithBrokerServer(brokerUsername, brokerPassword)

	if err != nil {
		t.Fatal(err)
	}
	defer controllerItems.Close()

	fakeKubeClient := controllerItems.FakeKubeClient
	fakeCatalogClient := controllerItems.FakeCatalogClient
	fakeBrokerServerHandler := controllerItems.BrokerServerHandler
	testController := controllerItems.Controller
	sharedInformers := controllerItems.Informers

	fakeBrokerServerHandler.Catalog = fakebrokerserver.ConvertCatalog(fakebrokerapi.GetTestCatalog())

	controller.AddGetNamespaceReaction(fakeKubeClient)

	// create the secret that the Broker resource will use for auth to the fake broker server
	fakeKubeClient.AddReactor("get", "secrets", func(clientgotesting.Action) (bool, runtime.Object, error) {
		return true, &v1.Secret{
			Data: map[string][]byte{
				"username": []byte(brokerUsername),
				"password": []byte(brokerPassword),
			},
		}, nil
	})

	// build the chain from Instance -> ServiceClass -> Broker
	testBroker := faketypes.GetBroker()
	testBroker.Spec.URL = controllerItems.BrokerServer.URL
	testBroker.Spec.AuthInfo = &v1alpha1.BrokerAuthInfo{
		BasicAuthSecret: &v1.ObjectReference{
			Namespace: "test",
			Name:      "test",
		},
	}
	testServiceClass := faketypes.GetServiceClass()
	testInstance := faketypes.GetInstance()

	sharedInformers.Brokers().Informer().GetStore().Add(testBroker)
	sharedInformers.ServiceClasses().Informer().GetStore().Add(testServiceClass)

	fakeCatalogClient.AddReactor("get", "serviceclasses", func(clientgotesting.Action) (bool, runtime.Object, error) {
		return true, testServiceClass, nil
	})
	fakeCatalogClient.AddReactor("get", "brokers", func(clientgotesting.Action) (bool, runtime.Object, error) {
		return true, testBroker, nil
	})

	// Make the provision return an error that is "unexpected"
	fakeBrokerServerHandler.ProvisionRespError = errors.New("test provision error")

	// there should be nothing that is polling since no instances have been reconciled yet
	if testController.PollingQueueLen() != 0 {
		t.Fatalf("Expected the polling queue to be empty")
	}

	reconcileErr := testController.ReconcileInstance(testInstance)
	if reconcileErr == nil {
		t.Fatalf("expected an error from reconcileInstance")
	}
	statusCodeErr, ok := reconcileErr.(osb.HTTPStatusCodeError)
	if !ok {
		t.Fatalf("expected an OSB client HTTPStatusCodeError, got %#v", reconcileErr)
	}
	if statusCodeErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected an internal server error, got %d", statusCodeErr.StatusCode)
	}

	actions := fakeCatalogClient.Actions()
	controller.AssertNumberOfActions(t, actions, 1)

	// verify that 2 kubernetes actions occurred - a GET to the ServiceClass, then a GET to the
	// Broker
	kubeActions := fakeKubeClient.Actions()
	controller.AssertNumberOfActions(t, kubeActions, 2)

	updatedInstance := controller.AssertUpdateStatus(t, actions[0], testInstance)
	controller.AssertInstanceReadyFalse(t, updatedInstance)

	// there should be 1 request to the provision endpoint
	numProvReqs := len(fakeBrokerServerHandler.ProvisionRequests)
	if numProvReqs != 1 {
		t.Fatalf("%d provision requests were made, expected 1", numProvReqs)
	}

	// The item should not have been added to the polling queue for later processing
	if testController.PollingQueueLen() != 0 {
		t.Fatalf("Expected polling queue to be empty")
	}
	controller.AssertAsyncOpInProgressFalse(t, updatedInstance)
	controller.AssertInstanceReadyFalse(t, updatedInstance)
	controller.AssertInstanceReadyCondition(t, updatedInstance, v1alpha1.ConditionFalse)
}
