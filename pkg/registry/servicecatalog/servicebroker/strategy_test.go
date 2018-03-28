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

package clusterservicebroker

import (
	"testing"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func clusterServiceBrokerWithOldSpec() *sc.ClusterServiceBroker {
	return &sc.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
		},
		Spec: sc.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: sc.CommonServiceBrokerSpec{
				URL: "https://kubernetes.default.svc:443/brokers/template.k8s.io",
			},
		},
		Status: sc.ClusterServiceBrokerStatus{
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

func clusterServiceBrokerWithNewSpec() *sc.ClusterServiceBroker {
	b := clusterServiceBrokerWithOldSpec()
	b.Spec.URL = "new"
	return b
}

// TestClusterServiceBrokerStrategyTrivial is the testing of the trivial hardcoded
// boolean flags.
func TestClusterServiceBrokerStrategyTrivial(t *testing.T) {
	if clusterServiceBrokerRESTStrategies.NamespaceScoped() {
		t.Errorf("clusterservicebroker create must not be namespace scoped")
	}
	if clusterServiceBrokerRESTStrategies.NamespaceScoped() {
		t.Errorf("clusterservicebroker update must not be namespace scoped")
	}
	if clusterServiceBrokerRESTStrategies.AllowCreateOnUpdate() {
		t.Errorf("clusterservicebroker should not allow create on update")
	}
	if clusterServiceBrokerRESTStrategies.AllowUnconditionalUpdate() {
		t.Errorf("clusterservicebroker should not allow unconditional update")
	}
}

// TestClusterServiceBrokerCreate
func TestClusterServiceBroker(t *testing.T) {
	// Create a clusterservicebroker or clusterservicebrokers
	broker := &sc.ClusterServiceBroker{
		Spec: sc.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: sc.CommonServiceBrokerSpec{
				URL: "abcd",
			},
		},
		Status: sc.ClusterServiceBrokerStatus{
			CommonServiceBrokerStatus: sc.CommonServiceBrokerStatus{
				Conditions: nil,
			},
		},
	}

	// Canonicalize the broker
	clusterServiceBrokerRESTStrategies.PrepareForCreate(nil, broker)

	if broker.Status.Conditions == nil {
		t.Fatalf("Fresh clusterservicebroker should have empty status")
	}
	if len(broker.Status.Conditions) != 0 {
		t.Fatalf("Fresh clusterservicebroker should have empty status")
	}
}

// TestClusterServiceBrokerUpdate tests that generation is incremented
// correctly when the spec of a ClusterServiceBroker is updated.
func TestClusterServiceBrokerUpdate(t *testing.T) {
	cases := []struct {
		name                      string
		older                     *sc.ClusterServiceBroker
		newer                     *sc.ClusterServiceBroker
		shouldGenerationIncrement bool
	}{
		{
			name:  "no spec change",
			older: clusterServiceBrokerWithOldSpec(),
			newer: clusterServiceBrokerWithOldSpec(),
			shouldGenerationIncrement: false,
		},
		{
			name:  "spec change",
			older: clusterServiceBrokerWithOldSpec(),
			newer: clusterServiceBrokerWithNewSpec(),
			shouldGenerationIncrement: true,
		},
	}

	for i := range cases {
		clusterServiceBrokerRESTStrategies.PrepareForUpdate(nil, cases[i].newer, cases[i].older)

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

// TestClusterServiceBrokerUpdateForRelistRequests tests that the RelistRequests field is
// ignored during updates when it is the default value.
func TestClusterServiceBrokerUpdateForRelistRequests(t *testing.T) {
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
		oldBroker := clusterServiceBrokerWithOldSpec()
		oldBroker.Spec.RelistRequests = tc.oldValue

		newClusterServiceBroker := clusterServiceBrokerWithOldSpec()
		newClusterServiceBroker.Spec.RelistRequests = tc.newValue

		clusterServiceBrokerRESTStrategies.PrepareForUpdate(nil, newClusterServiceBroker, oldBroker)

		if e, a := tc.expectedValue, newClusterServiceBroker.Spec.RelistRequests; e != a {
			t.Errorf("%s: got unexpected RelistRequests: expected %v, got %v", tc.name, e, a)
		}
	}
}
