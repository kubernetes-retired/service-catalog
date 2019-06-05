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
	"reflect"
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scfeatures "github.com/kubernetes-sigs/service-catalog/pkg/features"
	"github.com/kubernetes-sigs/service-catalog/test/fake"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientgotesting "k8s.io/client-go/testing"
)

// NOTE: This only tests a single test case. Others are tested in TestShouldReconcileClusterServiceBroker.
func TestShouldReconcileServiceBroker(t *testing.T) {
	broker := getTestClusterServiceBroker()
	broker.Spec.RelistDuration = &metav1.Duration{Duration: 3 * time.Minute}

	if !shouldReconcileClusterServiceBroker(broker, time.Now(), 24*time.Hour) {
		t.Error("expected true, bot got false")
	}
}

func TestReconcileServiceBrokerUpdatesBrokerClient(t *testing.T) {
	broker := getTestServiceBroker()
	broker.Name = broker.Name + "not-predefined"
	_, _, _, testController, _ := newTestController(t, noFakeActions())
	testController.reconcileServiceBroker(broker)

	_, found := testController.brokerClientManager.BrokerClient(NewServiceBrokerKey(broker.Namespace, broker.Name))
	if !found {
		t.Error("expected predefined OSB client")
	}
}

func getServiceBrokerReactor(broker *v1beta1.ServiceBroker) (string, string, clientgotesting.ReactionFunc) {
	return "get", "servicebrokers", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, broker, nil
	}
}

func listServiceClassesReactor(classes []v1beta1.ServiceClass) (string, string, clientgotesting.ReactionFunc) {
	return "list", "serviceclasses", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, &v1beta1.ServiceClassList{
			Items: classes,
		}, nil
	}
}

func listServicePlansReactor(plans []v1beta1.ServicePlan) (string, string, clientgotesting.ReactionFunc) {
	return "list", "serviceplans", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, &v1beta1.ServicePlanList{
			Items: plans,
		}, nil
	}
}

func reconcileServiceBroker(t *testing.T, testController *controller, broker *v1beta1.ServiceBroker) error {
	clone := broker.DeepCopy()
	err := testController.reconcileServiceBroker(broker)
	if !reflect.DeepEqual(broker, clone) {
		t.Errorf("reconcileServiceBroker shouldn't mutate input, but it does: %s", expectedGot(clone, broker))
	}
	return err
}

// TestReconcileServiceBrokerDelete simulates a broker reconciliation where broker was marked for deletion.
// Results in service class and broker both being deleted.
func TestReconcileServiceBrokerDelete(t *testing.T) {
	cases := []struct {
		name     string
		authInfo *v1beta1.ServiceBrokerAuthInfo
		secret   *corev1.Secret
	}{
		{
			name:     "no auth",
			authInfo: nil,
			secret:   nil,
		},
		{
			name:     "basic auth",
			authInfo: getTestBrokerBasicAuthInfo(),
			secret:   getTestBasicAuthSecret(),
		},
		{
			name:     "bearer auth",
			authInfo: getTestBrokerBearerAuthInfo(),
			secret:   getTestBearerAuthSecret(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			fakeKubeClient, fakeCatalogClient, fakeServiceBrokerClient, testController, _ := newTestController(t, getTestCatalogConfig())

			testServiceClass := getTestServiceClass()
			testServicePlan := getTestServicePlan()

			addGetSecretReaction(fakeKubeClient, tc.secret)

			broker := getTestServiceBrokerWithAuth(tc.authInfo)
			broker.DeletionTimestamp = &metav1.Time{}
			broker.Finalizers = []string{v1beta1.FinalizerServiceCatalog}

			updateBrokerClientCalled := false
			testController.brokerClientManager = NewBrokerClientManager(func(_ *osb.ClientConfiguration) (osb.Client, error) {
				updateBrokerClientCalled = true
				return nil, nil
			})

			fakeCatalogClient.AddReactor(getServiceBrokerReactor(broker))
			fakeCatalogClient.AddReactor(listServiceClassesReactor([]v1beta1.ServiceClass{*testServiceClass}))
			fakeCatalogClient.AddReactor(listServicePlansReactor([]v1beta1.ServicePlan{*testServicePlan}))

			// when
			err := reconcileServiceBroker(t, testController, broker)
			if err != nil {
				t.Fatalf("This should not fail : %v", err)
			}

			// then
			if updateBrokerClientCalled {
				t.Errorf("Unexpected broker client update action")
			}

			brokerActions := fakeServiceBrokerClient.Actions()
			assertNumberOfBrokerActions(t, brokerActions, 0)

			kubeActions := fakeKubeClient.Actions()
			assertNumberOfActions(t, kubeActions, 0)

			catalogActions := fakeCatalogClient.Actions()
			// The actions should be:
			// - list serviceplans
			// - delete serviceplans
			// - list serviceclasses
			// - delete serviceclass
			// - update the ready condition
			// - get the broker
			// - remove the finalizer
			assertNumberOfActions(t, catalogActions, 7)

			listRestrictions := clientgotesting.ListRestrictions{
				Labels: labels.Everything(),
				Fields: fields.OneTermEqualSelector("spec.serviceBrokerName", broker.Name),
			}
			assertList(t, catalogActions[0], &v1beta1.ServiceClass{}, listRestrictions)
			assertList(t, catalogActions[1], &v1beta1.ServicePlan{}, listRestrictions)
			assertDelete(t, catalogActions[2], testServicePlan)
			assertDelete(t, catalogActions[3], testServiceClass)
			updatedServiceBroker := assertUpdateStatus(t, catalogActions[4], broker)
			assertServiceBrokerReadyFalse(t, updatedServiceBroker)

			assertGet(t, catalogActions[5], broker)

			updatedServiceBroker = assertUpdateStatus(t, catalogActions[6], broker)
			assertEmptyFinalizers(t, updatedServiceBroker)

			events := getRecordedEvents(testController)

			expectedEvent := normalEventBuilder(successServiceBrokerDeletedReason).msg(
				"The servicebroker test-servicebroker was deleted successfully.",
			)
			if err := checkEvents(events, expectedEvent.stringArr()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

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
		catalogActionsCheckFunc func(t *testing.T, actions []clientgotesting.Action)
	}{
		{
			name:            "new class",
			newServiceClass: getTestServiceClass(),
			shouldError:     false,
			catalogActionsCheckFunc: func(t *testing.T, actions []clientgotesting.Action) {
				assertNumberOfActions(t, actions, 1)
				assertCreate(t, actions[0], getTestServiceClass())
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
			catalogActionsCheckFunc: func(t *testing.T, actions []clientgotesting.Action) {
				assertNumberOfActions(t, actions, 1)
				assertUpdate(t, actions[0], updatedClass())
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
		t.Run(tc.name, func(t *testing.T) {
			err := utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.NamespacedServiceBroker))
			if err != nil {
				t.Fatalf("Failed to enable namespaced service broker feature: %v", err)
			}
			defer utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.NamespacedServiceBroker))

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
					t.Fatalf("unexpected error from method under test: %v", err)
				} else if tc.errText != nil && *tc.errText != err.Error() {
					t.Fatalf("unexpected error text from method under test; %s", expectedGot(tc.errText, err.Error()))
				}
			}

			if tc.catalogActionsCheckFunc != nil {
				actions := fakeCatalogClient.Actions()
				tc.catalogActionsCheckFunc(t, actions)
			}
		})
	}
}

func TestReconcileServicePlanFromServiceBrokerCatalog(t *testing.T) {
	updatedPlan := func() *v1beta1.ServicePlan {
		p := getTestServicePlan()
		p.Spec.Description = "new-description"
		p.Spec.ExternalName = "new-value"
		p.Spec.Free = false
		p.Spec.ExternalMetadata = &runtime.RawExtension{Raw: []byte(`{"field1": "value1"}`)}
		p.Spec.InstanceCreateParameterSchema = &runtime.RawExtension{Raw: []byte(`{"field1": "value1"}`)}
		p.Spec.InstanceUpdateParameterSchema = &runtime.RawExtension{Raw: []byte(`{"field1": "value1"}`)}
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
		catalogActionsCheckFunc func(t *testing.T, actions []clientgotesting.Action)
	}{
		{
			name:           "new plan",
			newServicePlan: getTestServicePlan(),
			shouldError:    false,
			catalogActionsCheckFunc: func(t *testing.T, actions []clientgotesting.Action) {
				assertNumberOfActions(t, actions, 1)
				assertCreate(t, actions[0], getTestServicePlan())
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
			catalogActionsCheckFunc: func(t *testing.T, actions []clientgotesting.Action) {
				assertNumberOfActions(t, actions, 1)
				assertUpdate(t, actions[0], updatedPlan())
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
		t.Run(tc.name, func(t *testing.T) {
			err := utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.NamespacedServiceBroker))
			if err != nil {
				t.Fatalf("Failed to enable namespaced service broker feature: %v", err)
			}
			defer utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.NamespacedServiceBroker))

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
					t.Fatalf("unexpected error from method under test: %v", err)
				} else if tc.errText != nil && *tc.errText != err.Error() {
					t.Fatalf("unexpected error text from method under test; %s", expectedGot(tc.errText, err.Error()))
				}
			}

			if tc.catalogActionsCheckFunc != nil {
				actions := fakeCatalogClient.Actions()
				tc.catalogActionsCheckFunc(t, actions)
			}
		})
	}
}
