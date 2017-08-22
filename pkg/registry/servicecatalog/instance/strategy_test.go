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
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func instanceWithOldSpec() *servicecatalog.ServiceInstance {
	return &servicecatalog.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
		},
		Spec: servicecatalog.ServiceInstanceSpec{
			ServiceClassName: "test-serviceclass",
			PlanName:         "test-plan",
		},
		Status: servicecatalog.ServiceInstanceStatus{
			Conditions: []servicecatalog.ServiceInstanceCondition{
				{
					Type:   servicecatalog.ServiceInstanceConditionReady,
					Status: servicecatalog.ConditionFalse,
				},
			},
		},
	}
}

// TODO: Un-comment "spec-change" test case when there is a field
// in the spec to which the reconciler allows a change.

//func instanceWithNewSpec() *servicecatalog.ServiceInstance {
//	i := instanceWithOldSpec()
//	i.Spec.ServiceClassName = "new-serviceclass"
//	return i
//}

// TestInstanceUpdate tests that generation is incremented correctly when the
// spec of a Instance is updated.
func TestInstanceUpdate(t *testing.T) {
	cases := []struct {
		name                      string
		older                     *servicecatalog.ServiceInstance
		newer                     *servicecatalog.ServiceInstance
		shouldGenerationIncrement bool
	}{
		{
			name:  "no spec change",
			older: instanceWithOldSpec(),
			newer: instanceWithOldSpec(),
		},
		//{
		//	name:  "spec change",
		//	older: instanceWithOldSpec(),
		//	newer: instanceWithNewSpec(),
		//	shouldGenerationIncrement: true,
		//},
	}

	for _, tc := range cases {
		instanceRESTStrategies.PrepareForUpdate(nil, tc.newer, tc.older)

		expectedGeneration := tc.older.Generation
		if tc.shouldGenerationIncrement {
			expectedGeneration = expectedGeneration + 1
		}
		if e, a := expectedGeneration, tc.newer.Generation; e != a {
			t.Errorf("%v: expected %v, got %v for generation", tc.name, e, a)
		}
	}
}
