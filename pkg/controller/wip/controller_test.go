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

package wip

import (
	"testing"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	fakebrokerapi "github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	servicecataloginformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers"

	"k8s.io/client-go/1.5/kubernetes/fake"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/testing/core"
)

func TestReconcileBroker(t *testing.T) {
	// create a fake kube client
	fakeKubeClient := &fake.Clientset{}
	// create a fake sc client
	fakeCatalogClient := &servicecatalogclientset.Clientset{}

	catalogCl := &fakebrokerapi.CatalogClient{
		RetCatalog: &brokerapi.Catalog{
			Services: []*brokerapi.Service{{
				Name:        "test-service",
				ID:          "12345",
				Description: "a test service",
				Plans: []brokerapi.ServicePlan{{
					Name:        "test-plan",
					Free:        true,
					ID:          "34567",
					Description: "a test plan",
				}},
			}},
		},
	}
	instanceCl := fakebrokerapi.NewInstanceClient()
	bindingCl := fakebrokerapi.NewBindingClient()
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
		glog.Fatal(err)
	}

	broker := &v1alpha1.Broker{
		ObjectMeta: v1.ObjectMeta{Name: "test-name"},
		Spec: v1alpha1.BrokerSpec{
			URL:     "https://example.com",
			OSBGUID: "OSBGUID field",
		},
	}
	stopCh := make(chan struct{})
	informerFactory.Start(stopCh)

	// inject a broker resource into broker informer
	testController.(*controller).reconcileBroker(broker)

	actions := filterActions(fakeCatalogClient.Actions())
	if e, a := 2, len(actions); e != a {
		t.Logf("%+v\n", actions)
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	// first action should be a create action for a service class
	createAction := actions[0].(core.CreateAction)
	if e, a := "create", createAction.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[0]; expected %v, got %v", e, a)
	}

	createActionObject := createAction.GetObject().(*v1alpha1.ServiceClass)
	if e, a := "test-service", createActionObject.Name; e != a {
		t.Fatalf("Unexpected name of serviceClass created: expected %v, got %v", e, a)
	}

	// second action should be an update action for broker status subresource
	createAction2 := actions[1].(core.CreateAction)
	if e, a := "update", createAction2.GetVerb(); e != a {
		t.Fatalf("Unexpected verb on actions[1]; expected %v, got %v", e, a)
	}

	createActionObject2 := createAction2.GetObject().(*v1alpha1.Broker)
	if e, a := "test-name", createActionObject2.Name; e != a {
		t.Fatalf("Unexpected name of serviceClass created: expected %v, got %v", e, a)
	}

	// verify no kube resources created
	kubeActions := fakeKubeClient.Actions()
	if e, a := 0, len(kubeActions); e != a {
		t.Fatalf("Unexpected number of actions: expected %v, got %v", e, a)
	}

	stopCh <- struct{}{}
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
