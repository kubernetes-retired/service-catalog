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

package instance

import (
	"fmt"
	"testing"

	utilfeature "k8s.io/apiserver/pkg/util/feature"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	scfeatures "github.com/kubernetes-incubator/service-catalog/pkg/features"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/pkg/api/v1"
)

func getTestInstance() *servicecatalog.ServiceInstance {
	return &servicecatalog.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
		},
		Spec: servicecatalog.ServiceInstanceSpec{
			ExternalServiceClassName: "test-serviceclass",
			ExternalServicePlanName:  "test-plan",
			ServiceClassRef:          &v1.ObjectReference{},
			ServicePlanRef:           &v1.ObjectReference{},
			UserInfo: &servicecatalog.UserInfo{
				Username: "some-user",
			},
		},
		Status: servicecatalog.ServiceInstanceStatus{
			Conditions: []servicecatalog.ServiceInstanceCondition{
				{
					Type:   servicecatalog.ServiceInstanceConditionReady,
					Status: servicecatalog.ConditionTrue,
				},
			},
		},
	}
}

func contextWithUserName(userName string) genericapirequest.Context {
	ctx := genericapirequest.NewContext()
	userInfo := &user.DefaultInfo{
		Name: userName,
	}
	return genericapirequest.WithUser(ctx, userInfo)
}

// TestInstanceUpdate tests that updates to the spec of an Instance.
func TestInstanceUpdate(t *testing.T) {
	cases := []struct {
		name               string
		older              *servicecatalog.ServiceInstance
		newer              *servicecatalog.ServiceInstance
		shouldSpecUpdate   bool
		shouldPlanRefClear bool
	}{
		{
			name:  "no spec change",
			older: getTestInstance(),
			newer: getTestInstance(),
		},
		{
			name: "UpdateRequest increment",
			older: func() *servicecatalog.ServiceInstance {
				i := getTestInstance()
				i.Spec.UpdateRequests = 1
				return i
			}(),
			newer: func() *servicecatalog.ServiceInstance {
				i := getTestInstance()
				i.Spec.UpdateRequests = 2
				return i
			}(),
			shouldSpecUpdate: true,
		},
		{
			name:  "plan change",
			older: getTestInstance(),
			newer: func() *servicecatalog.ServiceInstance {
				i := getTestInstance()
				i.Spec.ExternalServicePlanName = "new-test-plan"
				return i
			}(),
			shouldSpecUpdate:   true,
			shouldPlanRefClear: true,
		},
	}

	for _, tc := range cases {
		instanceRESTStrategies.PrepareForUpdate(nil, tc.newer, tc.older)

		expectedGeneration := tc.older.Generation
		expectedReadyCondition := servicecatalog.ConditionTrue
		if tc.shouldSpecUpdate {
			expectedGeneration = expectedGeneration + 1
			expectedReadyCondition = servicecatalog.ConditionFalse
		}
		if e, a := expectedGeneration, tc.newer.Generation; e != a {
			t.Errorf("%v: expected %v, got %v for generation", tc.name, e, a)
			continue
		}
		if e, a := 1, len(tc.newer.Status.Conditions); e != a {
			t.Errorf("%v: unexpected number of conditions: expected %v, got %v", tc.name, e, a)
			continue
		}
		if e, a := servicecatalog.ServiceInstanceConditionReady, tc.newer.Status.Conditions[0].Type; e != a {
			t.Errorf("%v: unexpected condition type: expected %v, got %v", tc.name, e, a)
			continue
		}
		if e, a := expectedReadyCondition, tc.newer.Status.Conditions[0].Status; e != a {
			t.Errorf("%v: unexpected ready condition status: expected %v, got %v", tc.name, e, a)
		}
		if tc.shouldPlanRefClear {
			if tc.newer.Spec.ServicePlanRef != nil {
				t.Errorf("%v: expected ServicePlanRef to be nil", tc.name)
			}
		} else {
			if tc.newer.Spec.ServicePlanRef == nil {
				t.Errorf("%v: expected ServicePlanRef to not be nil", tc.name)
			}
		}
	}
}

// TestInstanceUserInfo tests that the user info is set properly
// as the user changes for different modifications of the instance.
func TestInstanceUserInfo(t *testing.T) {
	// Enable the OriginatingIdentity feature
	utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=true", scfeatures.OriginatingIdentity))
	defer utilfeature.DefaultFeatureGate.Set(fmt.Sprintf("%v=false", scfeatures.OriginatingIdentity))

	creatorUserName := "creator"
	createdInstance := getTestInstance()
	createContext := contextWithUserName(creatorUserName)
	instanceRESTStrategies.PrepareForCreate(createContext, createdInstance)

	if e, a := creatorUserName, createdInstance.Spec.UserInfo.Username; e != a {
		t.Errorf("unexpected user info in created spec: expected %v, got %v", e, a)
	}

	updaterUserName := "updater"
	updatedInstance := getTestInstance()
	updatedInstance.Spec.UpdateRequests = updatedInstance.Spec.UpdateRequests + 1
	updateContext := contextWithUserName(updaterUserName)
	instanceRESTStrategies.PrepareForUpdate(updateContext, updatedInstance, createdInstance)

	if e, a := updaterUserName, updatedInstance.Spec.UserInfo.Username; e != a {
		t.Errorf("unexpected user info in updated spec: expected %v, got %v", e, a)
	}

	deleterUserName := "deleter"
	deletedInstance := getTestInstance()
	deleteContext := contextWithUserName(deleterUserName)
	instanceRESTStrategies.CheckGracefulDelete(deleteContext, deletedInstance, nil)

	if e, a := deleterUserName, deletedInstance.Spec.UserInfo.Username; e != a {
		t.Errorf("unexpected user info in deleted spec: expected %v, got %v", e, a)
	}
}
