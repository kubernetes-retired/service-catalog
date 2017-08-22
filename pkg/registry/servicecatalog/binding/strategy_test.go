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

package binding

import (
	"testing"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func instanceCredentialWithOldSpec() *servicecatalog.ServiceInstanceCredential {
	return &servicecatalog.ServiceInstanceCredential{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
		},
		Spec: servicecatalog.ServiceInstanceCredentialSpec{
			ServiceInstanceRef: v1.LocalObjectReference{
				Name: "some-string",
			},
		},
		Status: servicecatalog.ServiceInstanceCredentialStatus{
			Conditions: []servicecatalog.ServiceInstanceCredentialCondition{
				{
					Type:   servicecatalog.ServiceInstanceCredentialConditionReady,
					Status: servicecatalog.ConditionFalse,
				},
			},
		},
	}
}

// TODO: Un-comment "spec-change" test case when there is a field
// in the spec to which the reconciler allows a change.

//func instanceCredentialWithNewSpec() *servicecatalog.ServiceInstanceCredential {
//	ic := instanceCredentialWithOldSpec()
//	ic.Spec.ServiceInstanceRef = v1.LocalObjectReference{
//		Name: "new-string",
//	}
//	return ic
//}

// TestInstanceCredentialUpdate tests that generation is incremented correctly when the
// spec of a ServiceInstanceCredential is updated.
func TestInstanceCredentialUpdate(t *testing.T) {
	cases := []struct {
		name                      string
		older                     *servicecatalog.ServiceInstanceCredential
		newer                     *servicecatalog.ServiceInstanceCredential
		shouldGenerationIncrement bool
	}{
		{
			name:  "no spec change",
			older: instanceCredentialWithOldSpec(),
			newer: instanceCredentialWithOldSpec(),
		},
		//{
		//	name:  "spec change",
		//	older: instanceCredentialWithOldSpec(),
		//	newer: instanceCredentialWithOldSpec(),
		//	shouldGenerationIncrement: true,
		//},
	}

	for _, tc := range cases {
		bindingRESTStrategies.PrepareForUpdate(nil, tc.newer, tc.older)

		expectedGeneration := tc.older.Generation
		if tc.shouldGenerationIncrement {
			expectedGeneration = expectedGeneration + 1
		}
		if e, a := expectedGeneration, tc.newer.Generation; e != a {
			t.Errorf("%v: expected %v, got %v for generation", tc.name, e, a)
		}
	}
}
