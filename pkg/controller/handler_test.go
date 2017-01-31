/*
Copyright 2016 The Kubernetes Authors.

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
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient/mem"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/injector"
	"k8s.io/kubernetes/pkg/api"
)

// TestCreateServiceInstanceHelper is a test for handler's createServiceInstance method. Note
// that it _only_ tests the lowercase (unexported) method, not the uppercase (exported) one.
func TestCreateServiceInstanceHelper(t *testing.T) {
	const (
		namespace    = "testNS"
		brokerName   = "testBroker"
		svcClassName = "testSvcClass"
	)
	// set up the mock (in-memory) APIClient implementation
	apiClient := mem.NewPopulatedAPIClient(
		map[string]*servicecatalog.Broker{
			brokerName: {
				ObjectMeta: api.ObjectMeta{
					Namespace: namespace,
					Name:      brokerName,
				},
			},
		},
		map[string]*servicecatalog.ServiceClass{
			svcClassName: {
				ObjectMeta: api.ObjectMeta{
					Namespace: namespace,
					Name:      svcClassName,
				},
				BrokerName: brokerName,
			},
		},
		make(map[string]apiclient.InstanceClient),
		make(map[string]apiclient.BindingClient),
	)
	// set up the mock injector
	// (we aren't exercising any functionality of this mock, so leaving it empty)
	inj := injector.NewFake()

	// set up the mock broker client (which is composed of catalog, instance and binding APIs).
	// we want these all to be "empty" to start because we'll be checking later that they were
	// properly called
	catalogCl := &fake.CatalogClient{}
	instanceCl := fake.NewInstanceClient()
	bindingCl := fake.NewBindingClient()
	brokerClFunc := fake.NewClientFunc(catalogCl, instanceCl, bindingCl)

	// set up the handler with the mocks that we've previously created.
	// we're exercising the handler and ensuring that it interacted with our mocks properly
	hdl := createHandler(apiClient, inj, brokerClFunc)

	// set up the instance that we're creating
	inst := &servicecatalog.Instance{
		Spec: servicecatalog.InstanceSpec{
			ServiceClassName: svcClassName,
		},
		Status: servicecatalog.InstanceStatus{},
	}
	if err := hdl.createServiceInstance(inst); err != nil {
		t.Fatalf("error creating service instance (%s)", err)
	}
	if len(instanceCl.Instances) != 1 {
		t.Fatalf("expected 1 created instance, got %d", len(instanceCl.Instances))
	}
	if len(bindingCl.Bindings) != 0 {
		t.Fatalf("expected 0 bindings, got %d", len(bindingCl.Bindings))
	}
	if len(inj.Injected) != 0 {
		t.Fatalf("expected 0 injected credentials, got %d", len(inj.Injected))
	}

	// check to ensure that the pre-populated broker was not deleted and none were
	// added
	brokersList, err := apiClient.Brokers().List()
	if err != nil {
		t.Fatalf("error getting stored brokers list (%s)", err)
	}
	if len(brokersList) != 1 {
		t.Fatalf("expected a single broker from the apiClient, got %d", len(brokersList))
	}
	broker := brokersList[0]
	if broker.Namespace != namespace {
		t.Fatalf("expected broker to have namespace '%s', got '%s'", namespace, broker.Namespace)
	}
	if broker.Name != brokerName {
		t.Fatalf("expected broker to have name '%s', got '%s'", brokerName, broker.Name)
	}

	// check to ensure that the pre-populated service class was not deleted and
	// none were added
	svcClassList, err := apiClient.ServiceClasses().List()
	if err != nil {
		t.Fatalf("error getting service classes list (%s)", err)
	}
	if len(svcClassList) != 1 {
		t.Fatalf("expected a single service class from the apiClient, got %d", len(svcClassList))
	}
	svcClass := svcClassList[0]
	if svcClass.Namespace != namespace {
		t.Fatalf("expected service class to have namespace '%s', got '%s'", namespace, svcClass.Namespace)
	}
	if svcClass.Name != svcClassName {
		t.Fatalf("expected service class to have name '%s', got '%s'", svcClassName, svcClass.Name)
	}

	// check to ensure that no instances were created and none were added. Note
	// that the createServiceInstance function (lowercase) only calls the OSB
	// service broker client. It should not mutate the underlying storage (the
	// uppercase function does that however)
	instList, err := apiClient.Instances(namespace).List()
	if err != nil {
		t.Fatalf("error getting instances list (%s)", err)
	}
	if len(instList) != 0 {
		t.Fatalf("expected no instances from the apiClient, got %d", len(instList))
	}

	// check to ensure that no bindings were created
	bindingsList, err := apiClient.Bindings(namespace).List()
	if err != nil {
		t.Fatalf("error getting bindings list (%s)", err)
	}
	if len(bindingsList) != 0 {
		t.Fatalf("expected no bindings from the apiClient, got %d", len(bindingsList))
	}
}
