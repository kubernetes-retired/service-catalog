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

package servicebroker

import (
	"testing"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func serviceBrokerWithOldSpec() *sc.ServiceBroker {
	return &sc.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
		},
		Spec: sc.ServiceBrokerSpec{
			CommonServiceBrokerSpec: sc.CommonServiceBrokerSpec{
				URL: "https://kubernetes.default.svc:443/brokers/template.k8s.io",
			},
		},
		Status: sc.ServiceBrokerStatus{
			CommonServiceBrokerStatus: sc.CommonServiceBrokerStatus{
				Conditions: []sc.ServiceBrokerCondition{
					{
						Type:   sc.ServiceBrokerConditionReady,
						Status: sc.ConditionFalse,
					},
				},
			},
		},
	}
}

func serviceBrokerWithNewSpec() *sc.ServiceBroker {
	b := serviceBrokerWithOldSpec()
	b.Spec.URL = "new"
	return b
}

// TestServiceBrokerStrategyTrivial is the testing of the trivial hardcoded
// boolean flags.
func TestServiceBrokerStrategyTrivial(t *testing.T) {
	if serviceBrokerRESTStrategies.NamespaceScoped() {
		t.Errorf("servicebroker create must not be namespace scoped")
	}
	if serviceBrokerRESTStrategies.NamespaceScoped() {
		t.Errorf("servicebroker update must not be namespace scoped")
	}
	if serviceBrokerRESTStrategies.AllowCreateOnUpdate() {
		t.Errorf("servicebroker should not allow create on update")
	}
	if serviceBrokerRESTStrategies.AllowUnconditionalUpdate() {
		t.Errorf("servicebroker should not allow unconditional update")
	}
}

// TestServiceBrokerCreate
func TestServiceBroker(t *testing.T) {
	// Create a servicebroker or servicebrokers
	broker := &sc.ServiceBroker{
		Spec: sc.ServiceBrokerSpec{
			CommonServiceBrokerSpec: sc.CommonServiceBrokerSpec{
				URL: "abcd",
			},
		},
		Status: sc.ServiceBrokerStatus{
			CommonServiceBrokerStatus: sc.CommonServiceBrokerStatus{
				Conditions: nil,
			},
		},
	}

	// Canonicalize the broker
	serviceBrokerRESTStrategies.PrepareForCreate(nil, broker)

	if broker.Status.Conditions == nil {
		t.Fatalf("Fresh servicebroker should have empty status")
	}
	if len(broker.Status.Conditions) != 0 {
		t.Fatalf("Fresh servicebroker should have empty status")
	}
}

// TestServiceBrokerUpdate tests that generation is incremented
// correctly when the spec of a ServiceBroker is updated.
func TestServiceBrokerUpdate(t *testing.T) {
	cases := []struct {
		name                      string
		older                     *sc.ServiceBroker
		newer                     *sc.ServiceBroker
		shouldGenerationIncrement bool
	}{
		{
			name:  "no spec change",
			older: serviceBrokerWithOldSpec(),
			newer: serviceBrokerWithOldSpec(),
			shouldGenerationIncrement: false,
		},
		{
			name:  "spec change",
			older: serviceBrokerWithOldSpec(),
			newer: serviceBrokerWithNewSpec(),
			shouldGenerationIncrement: true,
		},
	}

	for i := range cases {
		serviceBrokerRESTStrategies.PrepareForUpdate(nil, cases[i].newer, cases[i].older)

		if cases[i].shouldGenerationIncrement {
			if e, a := cases[i].older.Generation+1, cases[i].newer.Generation; e != a {
				t.Fatalf("%v: expected %v, got %v for generation", cases[i].name, e, a)
			}
		} else {
			if e, a := cases[i].older.Generation, cases[i].newer.Generation; e != a {
				t.Fatalf("%v: expected %v, got %v for generation", cases[i].name, e, a)
			}
		}
	}
}

// TestServiceBrokerUpdateForRelistRequests tests that the RelistRequests field is
// ignored during updates when it is the default value.
func TestServiceBrokerUpdateForRelistRequests(t *testing.T) {
	cases := []struct {
		name          string
		oldValue      int64
		newValue      int64
		expectedValue int64
	}{
		{
			name:          "both default",
			oldValue:      0,
			newValue:      0,
			expectedValue: 0,
		},
		{
			name:          "old default",
			oldValue:      0,
			newValue:      1,
			expectedValue: 1,
		},
		{
			name:          "new default",
			oldValue:      1,
			newValue:      0,
			expectedValue: 1,
		},
		{
			name:          "neither default",
			oldValue:      1,
			newValue:      2,
			expectedValue: 2,
		},
	}
	for _, tc := range cases {
		oldBroker := serviceBrokerWithOldSpec()
		oldBroker.Spec.RelistRequests = tc.oldValue

		newServiceBroker := serviceBrokerWithOldSpec()
		newServiceBroker.Spec.RelistRequests = tc.newValue

		serviceBrokerRESTStrategies.PrepareForUpdate(nil, newServiceBroker, oldBroker)

		if e, a := tc.expectedValue, newServiceBroker.Spec.RelistRequests; e != a {
			t.Errorf("%s: got unexpected RelistRequests: expected %v, got %v", tc.name, e, a)
		}
	}
}
