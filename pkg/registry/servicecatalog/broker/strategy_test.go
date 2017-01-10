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

package broker

import (
	"testing"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

// TestBrokerStrategyTrivial is the testing of the trivial hardcoded
// boolean flags.
func TestBrokerStrategyTrivial(t *testing.T) {
	if createStrategy.NamespaceScoped() {
		t.Errorf("broker create must be namespace scoped")
	}
	if updateStrategy.NamespaceScoped() {
		t.Errorf("broker update must be namespace scoped")
	}
	if updateStrategy.AllowCreateOnUpdate() {
		t.Errorf("Job should not allow create on update")
	}
	if updateStrategy.AllowUnconditionalUpdate() {
		t.Errorf("Job should not allow create on update")
	}
}

// TestBrokerCanonicalize tes
func TestBroker(t *testing.T) {
	// Create a broker or brokers
	broker := &sc.Broker{
		Spec: sc.BrokerSpec{
			URL:          "abcd",
			AuthUsername: "user",
			AuthPassword: "pass",
			OSBGUID:      "guid",
		},
		Status: sc.BrokerStatus{
			Conditions: nil,
		},
	}

	// Canonicalize the broker
	createStrategy.Canonicalize(broker)

	// Check that canonicalize did the appropriate stuff to the broker.

	// Imagine a table driven series of subtests that make sure
	// each individual aspect of canonicalization works.
}
