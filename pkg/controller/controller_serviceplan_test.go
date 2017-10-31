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

	"github.com/golang/mock/gomock"
	mocklisters "github.com/kubernetes-incubator/service-catalog/pkg/client/listers_generated/servicecatalog/v1beta1/mocks"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/test/fake"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
	"k8s.io/client-go/util/workqueue"
)

func TestReconcileClusterServicePlanRemovedFromCatalog(t *testing.T) {
	getRemovedPlan := func() *v1beta1.ClusterServicePlan {
		p := getTestClusterServicePlan()
		p.Status.RemovedFromBrokerCatalog = true
		return p
	}

	cases := []struct {
		name                    string
		plan                    *v1beta1.ClusterServicePlan
		catalogClientPrepFunc   func(*fake.Clientset)
		shouldError             bool
		errText                 *string
		catalogActionsCheckFunc func(t *testing.T, name string, actions []clientgotesting.Action)
	}{
		{
			name:        "not removed from catalog",
			plan:        getTestClusterServicePlan(),
			shouldError: false,
			catalogClientPrepFunc: func(client *fake.Clientset) {
				client.AddReactor("list", "serviceinstances", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					return true, &v1beta1.ServiceInstanceList{Items: nil}, nil
				})
			},
		},
		{
			name: "removed from catalog, instances left",
			plan: getRemovedPlan(),

			shouldError: false,
			catalogClientPrepFunc: func(client *fake.Clientset) {
				instances := []v1beta1.ServiceInstance{*getTestServiceInstance()}
				client.AddReactor("list", "serviceinstances", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					return true, &v1beta1.ServiceInstanceList{Items: instances}, nil
				})
			},
			catalogActionsCheckFunc: func(t *testing.T, name string, actions []clientgotesting.Action) {
				listRestrictions := clientgotesting.ListRestrictions{
					Labels: labels.Everything(),
					Fields: fields.OneTermEqualSelector("spec.clusterServicePlanRef.name", "PGUID"),
				}

				expectNumberOfActions(t, name, actions, 1)
				assertList(t, actions[0], &v1beta1.ServiceInstance{}, listRestrictions)
			},
		},
		{
			name:        "removed from catalog, no instances left",
			plan:        getRemovedPlan(),
			shouldError: false,
			catalogClientPrepFunc: func(client *fake.Clientset) {
				client.AddReactor("list", "serviceinstances", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					return true, &v1beta1.ServiceInstanceList{Items: nil}, nil
				})
			},
			catalogActionsCheckFunc: func(t *testing.T, name string, actions []clientgotesting.Action) {
				listRestrictions := clientgotesting.ListRestrictions{
					Labels: labels.Everything(),
					Fields: fields.OneTermEqualSelector("spec.clusterServicePlanRef.name", "PGUID"),
				}

				expectNumberOfActions(t, name, actions, 2)
				assertList(t, actions[0], &v1beta1.ServiceInstance{}, listRestrictions)
				assertDelete(t, actions[1], getRemovedPlan())
			},
		},
		{
			name: "removed from catalog, no instances left, delete fails", plan: getRemovedPlan(),
			shouldError: true,
			catalogClientPrepFunc: func(client *fake.Clientset) {
				client.AddReactor("list", "serviceinstances", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					return true, &v1beta1.ServiceInstanceList{Items: nil}, nil
				})
				client.AddReactor("delete", "clusterserviceplans", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewBadRequest("oops")
				})
			},
			errText: strPtr("oops"),
			catalogActionsCheckFunc: func(t *testing.T, name string, actions []clientgotesting.Action) {
				listRestrictions := clientgotesting.ListRestrictions{
					Labels: labels.Everything(),
					Fields: fields.OneTermEqualSelector("spec.clusterServicePlanRef.name", "PGUID"),
				}

				expectNumberOfActions(t, name, actions, 2)
				assertList(t, actions[0], &v1beta1.ServiceInstance{}, listRestrictions)
				assertDelete(t, actions[1], getRemovedPlan())
			},
		},
		{
			name: "plan not found, reconcile fails", plan: getRemovedPlan(),
			shouldError: true,
			errText:     strPtr("error on list"),
			catalogClientPrepFunc: func(client *fake.Clientset) {
				client.AddReactor("list", "serviceinstances", func(action clientgotesting.Action) (bool, runtime.Object, error) {
					return true, nil, fmt.Errorf("error on list")
				})
			},
			catalogActionsCheckFunc: func(t *testing.T, name string, actions []clientgotesting.Action) {
				listRestrictions := clientgotesting.ListRestrictions{
					Labels: labels.Everything(),
					Fields: fields.OneTermEqualSelector("spec.clusterServicePlanRef.name", "PGUID"),
				}

				expectNumberOfActions(t, name, actions, 1)
				assertList(t, actions[0], &v1beta1.ServiceInstance{}, listRestrictions)
			},
		},
	}

	for _, tc := range cases {
		_, fakeCatalogClient, _, testController, _ := newTestController(t, noFakeActions())

		if tc.catalogClientPrepFunc != nil {
			tc.catalogClientPrepFunc(fakeCatalogClient)
		}

		err := testController.reconcileClusterServicePlan(tc.plan)
		if err != nil {
			if !tc.shouldError {
				t.Errorf("%v: unexpected error from method under test: %v", tc.name, err)
				continue
			} else if tc.errText != nil && *tc.errText != err.Error() {
				t.Errorf("%v: unexpected error text from method under test; expected %v, got %v", tc.name, tc.errText, err.Error())
				continue
			}
		}

		actions := fakeCatalogClient.Actions()

		if tc.catalogActionsCheckFunc != nil {
			tc.catalogActionsCheckFunc(t, tc.name, actions)
		} else {
			expectNumberOfActions(t, tc.name, actions, 0)
		}
	}
}

func TestServicePlanAdd(t *testing.T) {
	// setup
	servicePlanQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-plan")

	// create controller
	testController := controller{
		servicePlanQueue: servicePlanQueue,
	}

	// perform test
	testController.servicePlanAdd(getTestClusterServicePlan())

	if servicePlanQueue.Len() != 1 {
		t.Fatalf("servicePlanQueue length error: %s", expectedGot(1, servicePlanQueue.Len()))
	}

	planName, _ := servicePlanQueue.Get()

	if planName != testClusterServicePlanGUID {
		t.Fatalf("Wrong plan queued: %s", expectedGot(testClusterServicePlanGUID, planName))
	}
}

func TestServicePlanAddFail(t *testing.T) {
	// setup
	servicePlanQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-plan")

	// create controller
	testController := controller{
		servicePlanQueue: servicePlanQueue,
	}

	// perform test
	testController.servicePlanAdd(nil)

	if servicePlanQueue.Len() != 0 {
		t.Fatalf("servicePlanQueue length error: %s", expectedGot(0, servicePlanQueue.Len()))
	}
}

func TestServicePlanUpdate(t *testing.T) {
	// TODO(n3wscott): it looks like servicePlanUpdate is not implemented yet
}

func TestServicePlanDelete(t *testing.T) {
	_, _, _, testController, _ := newTestController(t, noFakeActions())

	testController.servicePlanDelete(nil)
	testController.servicePlanDelete(getTestClusterServicePlan())
	// TODO(#1407): Nothing to test yet
}

func TestReconcileClusterServicePlanKeyThrowsError(t *testing.T) {
	// setup mocks
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClusterServicePlansLister := mocklisters.NewMockClusterServicePlanLister(mockCtrl)
	mockClusterServicePlansLister.EXPECT().Get("key").Return(nil, fmt.Errorf("error"))

	// create controller
	testController := controller{servicePlanLister: mockClusterServicePlansLister}

	// perform test
	if err := testController.reconcileClusterServicePlanKey("key"); err == nil {
		t.Fatalf("Should have returned an error.")
	}
}

func TestReconcileClusterServicePlanKeyNotFound(t *testing.T) {
	// setup mocks
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClusterServicePlansLister := mocklisters.NewMockClusterServicePlanLister(mockCtrl)
	expectedErr := errors.NewNotFound(v1beta1.Resource("serviceplan"), "key")
	mockClusterServicePlansLister.EXPECT().Get("key").Return(nil, expectedErr)

	// create controller
	testController := controller{servicePlanLister: mockClusterServicePlansLister}

	// perform test
	if err := testController.reconcileClusterServicePlanKey("key"); err != nil {
		t.Fatalf("Should have not returned an error.")
	}
}
