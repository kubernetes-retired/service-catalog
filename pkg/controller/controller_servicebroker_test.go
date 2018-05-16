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

package controller

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scfeatures "github.com/kubernetes-incubator/service-catalog/pkg/features"
	"github.com/kubernetes-incubator/service-catalog/test/fake"
	"k8s.io/apimachinery/pkg/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientgotesting "k8s.io/client-go/testing"
)

func TestReconcileServiceClassFromServiceBrokerCatalog(t *testing.T) {
	updatedClass := func() *v1beta1.ServiceClass {
		p := getTestServiceClass()
		p.Spec.Description = "new-description"
		p.Spec.ExternalName = "new-value"
		p.Spec.Bindable = false
		p.Spec.ExternalMetadata = &runtime.RawExtension{Raw: []byte(`{"field1": "value1"}`)}
		return p
	}

	cases := []struct {
		name                    string
		newServiceClass         *v1beta1.ServiceClass
		existingServiceClass    *v1beta1.ServiceClass
		listerServiceClass      *v1beta1.ServiceClass
		shouldError             bool
		errText                 *string
		catalogClientPrepFunc   func(*fake.Clientset)
		catalogActionsCheckFunc func(t *testing.T, name string, actions []clientgotesting.Action)
	}{
		{
			name:            "new class",
			newServiceClass: getTestServiceClass(),
			shouldError:     false,
			catalogActionsCheckFunc: func(t *testing.T, name string, actions []clientgotesting.Action) {
				expectNumberOfActions(t, name, actions, 1)
				expectCreate(t, name, actions[0], getTestServiceClass())
			},
		},
		{
			name:                 "exists, but for a different broker",
			newServiceClass:      getTestServiceClass(),
			existingServiceClass: getTestServiceClass(),
			listerServiceClass: func() *v1beta1.ServiceClass {
				p := getTestServiceClass()
				p.Spec.ServiceBrokerName = "something-else"
				return p
			}(),
			shouldError: true,
			errText:     strPtr(`ServiceBroker "test-servicebroker": ServiceClass "test-serviceclass" already exists for Broker "something-else"`),
		},
		{
			name:                 "class update",
			newServiceClass:      updatedClass(),
			existingServiceClass: getTestServiceClass(),
			shouldError:          false,
			catalogActionsCheckFunc: func(t *testing.T, name string, actions []clientgotesting.Action) {
				expectNumberOfActions(t, name, actions, 1)
				expectUpdate(t, name, actions[0], updatedClass())
			},
		},
		{
			name:                 "class update - failure",
			newServiceClass:      updatedClass(),
			existingServiceClass: getTestServiceClass(),
			catalogClientPrepFunc: func(client *fake.Clientset) {
				client.AddReactor("update", "serviceclasss", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("oops")
				})
			},
			shouldError: true,
			errText:     strPtr("oops"),
		},
	}

	broker := getTestServiceBroker()

	for _, tc := range cases {
		err := utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.NamespacedServiceBroker))
		if err != nil {
			t.Fatalf("Failed to enable namespaced service broker feature: %v", err)
		}
		defer utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.NamespacedServiceBroker))

		_, fakeCatalogClient, _, testController, sharedInformers := newTestController(t, noFakeActions())
		if tc.catalogClientPrepFunc != nil {
			tc.catalogClientPrepFunc(fakeCatalogClient)
		}

		if tc.listerServiceClass != nil {
			sharedInformers.ServiceClasses().Informer().GetStore().Add(tc.listerServiceClass)
		}

		err = testController.reconcileServiceClassFromServiceBrokerCatalog(broker, tc.newServiceClass, tc.existingServiceClass)
		if err != nil {
			if !tc.shouldError {
				t.Errorf("%v: unexpected error from method under test: %v", tc.name, err)
				continue
			} else if tc.errText != nil && *tc.errText != err.Error() {
				t.Errorf("%v: unexpected error text from method under test; %s", tc.name, expectedGot(tc.errText, err.Error()))
				continue
			}
		}

		if tc.catalogActionsCheckFunc != nil {
			actions := fakeCatalogClient.Actions()
			tc.catalogActionsCheckFunc(t, tc.name, actions)
		}
	}
}

func TestReconcileServicePlanFromServiceBrokerCatalog(t *testing.T) {
	updatedPlan := func() *v1beta1.ServicePlan {
		p := getTestServicePlan()
		p.Spec.Description = "new-description"
		p.Spec.ExternalName = "new-value"
		p.Spec.Free = false
		p.Spec.ExternalMetadata = &runtime.RawExtension{Raw: []byte(`{"field1": "value1"}`)}
		p.Spec.ServiceInstanceCreateParameterSchema = &runtime.RawExtension{Raw: []byte(`{"field1": "value1"}`)}
		p.Spec.ServiceInstanceUpdateParameterSchema = &runtime.RawExtension{Raw: []byte(`{"field1": "value1"}`)}
		p.Spec.ServiceBindingCreateParameterSchema = &runtime.RawExtension{Raw: []byte(`{"field1": "value1"}`)}

		return p
	}

	cases := []struct {
		name                    string
		newServicePlan          *v1beta1.ServicePlan
		existingServicePlan     *v1beta1.ServicePlan
		listerServicePlan       *v1beta1.ServicePlan
		shouldError             bool
		errText                 *string
		catalogClientPrepFunc   func(*fake.Clientset)
		catalogActionsCheckFunc func(t *testing.T, name string, actions []clientgotesting.Action)
	}{
		{
			name:           "new plan",
			newServicePlan: getTestServicePlan(),
			shouldError:    false,
			catalogActionsCheckFunc: func(t *testing.T, name string, actions []clientgotesting.Action) {
				expectNumberOfActions(t, name, actions, 1)
				expectCreate(t, name, actions[0], getTestServicePlan())
			},
		},
		{
			name:                "exists, but for a different broker",
			newServicePlan:      getTestServicePlan(),
			existingServicePlan: getTestServicePlan(),
			listerServicePlan: func() *v1beta1.ServicePlan {
				p := getTestServicePlan()
				p.Spec.ServiceBrokerName = "something-else"
				return p
			}(),
			shouldError: true,
			errText:     strPtr(`ServiceBroker "test-servicebroker": ServicePlan "test-serviceplan" already exists for Broker "something-else"`),
		},
		{
			name:                "plan update",
			newServicePlan:      updatedPlan(),
			existingServicePlan: getTestServicePlan(),
			shouldError:         false,
			catalogActionsCheckFunc: func(t *testing.T, name string, actions []clientgotesting.Action) {
				expectNumberOfActions(t, name, actions, 1)
				expectUpdate(t, name, actions[0], updatedPlan())
			},
		},
		{
			name:                "plan update - failure",
			newServicePlan:      updatedPlan(),
			existingServicePlan: getTestServicePlan(),
			catalogClientPrepFunc: func(client *fake.Clientset) {
				client.AddReactor("update", "serviceplans", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("oops")
				})
			},
			shouldError: true,
			errText:     strPtr("oops"),
		},
	}

	broker := getTestServiceBroker()

	for _, tc := range cases {
		err := utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.NamespacedServiceBroker))
		if err != nil {
			t.Fatalf("Failed to enable namespaced service broker feature: %v", err)
		}
		defer utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.NamespacedServiceBroker))

		_, fakeCatalogClient, _, testController, sharedInformers := newTestController(t, noFakeActions())
		if tc.catalogClientPrepFunc != nil {
			tc.catalogClientPrepFunc(fakeCatalogClient)
		}

		if tc.listerServicePlan != nil {
			sharedInformers.ServicePlans().Informer().GetStore().Add(tc.listerServicePlan)
		}

		err = testController.reconcileServicePlanFromServiceBrokerCatalog(broker, tc.newServicePlan, tc.existingServicePlan)
		if err != nil {
			if !tc.shouldError {
				t.Errorf("%v: unexpected error from method under test: %v", tc.name, err)
				continue
			} else if tc.errText != nil && *tc.errText != err.Error() {
				t.Errorf("%v: unexpected error text from method under test; %s", tc.name, expectedGot(tc.errText, err.Error()))
				continue
			}
		}

		if tc.catalogActionsCheckFunc != nil {
			actions := fakeCatalogClient.Actions()
			tc.catalogActionsCheckFunc(t, tc.name, actions)
		}
	}
}
