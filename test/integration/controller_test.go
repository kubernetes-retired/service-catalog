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
	"testing"
	"time"

	"k8s.io/client-go/1.5/kubernetes/fake"
	"k8s.io/kubernetes/pkg/api/v1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	fakebrokerapi "github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	scinformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated"
	informers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	"github.com/kubernetes-incubator/service-catalog/test/util"
)

// TestAddAndRemoveBroker tests the add broker flow.  It creates a new Broker
// resource and waits for that Broker to have a ready condition with true
// status, establishes that the controller created the expected ServiceClass,
// then deletes the Broker and establishes that the controller cleaned up the
// ServiceClass.
func TestAddAndRemoveBroker(t *testing.T) {
	_, catalogClient, fakeBrokerCatalog, _, _, _, _, shutdownServer := newTestController(t)
	defer shutdownServer()

	client := catalogClient.ServicecatalogV1alpha1()

	const (
		testBrokerName       = "test-broker"
		testServiceClassName = "test-service"
	)

	fakeBrokerCatalog.RetCatalog = &brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				Name:        testServiceClassName,
				ID:          "12345",
				Description: "a test service",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "test-plan",
						Free:        true,
						ID:          "34567",
						Description: "a test plan",
					},
				},
			},
		},
	}
	broker := &v1alpha1.Broker{
		ObjectMeta: v1.ObjectMeta{Name: testBrokerName},
		Spec: v1alpha1.BrokerSpec{
			URL: "https://example.com",
		},
	}

	_, err := client.Brokers().Create(broker)
	if nil != err {
		t.Fatalf("error creating the broker %q (%q)", broker, err)
	}

	err = util.WaitForBrokerCondition(client,
		testBrokerName,
		v1alpha1.BrokerCondition{
			Type:   v1alpha1.BrokerConditionReady,
			Status: v1alpha1.ConditionTrue,
		})
	if err != nil {
		t.Fatal(err)
	}

	err = util.WaitForServiceClassToExist(client, testServiceClassName)
	if nil != err {
		t.Fatalf("error waiting from ServiceClass to exist: %v", err)
	}

	// Delete the broker
	err = client.Brokers().Delete(testBrokerName, &v1.DeleteOptions{})
	if nil != err {
		t.Fatalf("broker should be deleted (%s)", err)
	}

	err = util.WaitForServiceClassToNotExist(client, testServiceClassName)
	if err != nil {
		t.Fatalf("error waiting for ServiceClass to not exist: %v", err)
	}

	err = util.WaitForBrokerToNotExist(client, testBrokerName)
	if err != nil {
		t.Fatalf("error waiting for Broker to not exist: %v", err)
	}
}

func newTestController(t *testing.T) (
	*fake.Clientset,
	clientset.Interface,
	*fakebrokerapi.CatalogClient,
	*fakebrokerapi.InstanceClient,
	*fakebrokerapi.BindingClient,
	controller.Controller,
	informers.Interface,
	func(),
) {
	// create a fake kube client
	fakeKubeClient := &fake.Clientset{}
	// create an sc client and running server
	catalogClient, shutdownServer := getFreshApiserverAndClient(t, server.StorageTypeEtcd.String())

	catalogCl := &fakebrokerapi.CatalogClient{}
	instanceCl := fakebrokerapi.NewInstanceClient()
	bindingCl := fakebrokerapi.NewBindingClient()
	brokerClFunc := fakebrokerapi.NewClientFunc(catalogCl, instanceCl, bindingCl)

	// create informers
	resync, _ := time.ParseDuration("1m")
	informerFactory := scinformers.NewSharedInformerFactory(nil, catalogClient, resync)
	serviceCatalogSharedInformers := informerFactory.Servicecatalog().V1alpha1()

	// create a test controller
	testController, err := controller.NewController(
		fakeKubeClient,
		catalogClient.ServicecatalogV1alpha1(),
		serviceCatalogSharedInformers.Brokers(),
		serviceCatalogSharedInformers.ServiceClasses(),
		serviceCatalogSharedInformers.Instances(),
		serviceCatalogSharedInformers.Bindings(),
		brokerClFunc,
	)
	if err != nil {
		t.Fatal(err)
	}

	stopCh := make(chan struct{})
	informerFactory.Start(stopCh)

	return fakeKubeClient, catalogClient, catalogCl, instanceCl, bindingCl,
		testController, serviceCatalogSharedInformers, shutdownServer
}
