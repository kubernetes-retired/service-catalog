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
	"fmt"
	"testing"
	"time"

	"k8s.io/client-go/1.5/kubernetes/fake"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/util/wait"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	fakebrokerapi "github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	scinformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated"
	informers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
)

// TestController is a very basic test to start
//
// need etcd running
// start a fresh apiserver for the controller to talk to
func TestController(t *testing.T) {
	fakeKubeClient, catalogClient, fakeBrokerCatalog, _, _, testController, _, stopCh := newTestController(t)
	defer close(stopCh)

	t.Log(fakeKubeClient, catalogClient, fakeBrokerCatalog, testController, stopCh)

	fakeBrokerCatalog.RetCatalog = &brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				Name:        "test-service",
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
	name := "test-name"
	broker := &v1alpha1.Broker{
		ObjectMeta: v1.ObjectMeta{Name: name},
		Spec: v1alpha1.BrokerSpec{
			URL: "https://example.com",
		},
	}
	brokerClient := catalogClient.Servicecatalog().Brokers()

	brokerServer, err := brokerClient.Create(broker)
	if nil != err {
		t.Fatalf("error creating the broker %q (%q)", broker, err)
	}

	if err := wait.PollImmediate(500*time.Millisecond, wait.ForeverTestTimeout,
		func() (bool, error) {
			brokerServer, err = brokerClient.Get(name)
			if nil != err {
				return false,
					fmt.Errorf("error getting broker %s (%s)",
						name, err)
			} else if len(brokerServer.Status.Conditions) > 0 {
				t.Log(brokerServer)
				return true, nil
			} else {
				return false, nil
			}
		},
	); err != nil {
		t.Fatal(err)
	}

	// check
	serviceClassClient := catalogClient.Servicecatalog().ServiceClasses()
	_, err = serviceClassClient.Get("test-service")
	if nil != err {
		t.Fatal("could not find the test service", err)
	}

	// cleanup our broker
	err = brokerClient.Delete(name, &v1.DeleteOptions{})
	if nil != err {
		t.Fatalf("broker should be deleted (%s)", err)
	}

	// uncomment if/when deleting a broker deletes the associated service
	// if class, err := serviceClassClient.Get("test-service"); nil == err {
	// 	t.Fatal("found the test service that should have been deleted", err, class)
	// }
}

func newTestController(t *testing.T) (
	*fake.Clientset,
	clientset.Interface,
	*fakebrokerapi.CatalogClient,
	*fakebrokerapi.InstanceClient,
	*fakebrokerapi.BindingClient,
	controller.Controller,
	informers.Interface,
	chan struct{},
) {
	// create a fake kube client
	fakeKubeClient := &fake.Clientset{}
	// create an sc client and running server
	catalogClient, shutdownServer := getFreshApiserverAndClient(t, server.StorageTypeEtcd.String())
	defer shutdownServer()

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
		testController, serviceCatalogSharedInformers, stopCh
}
