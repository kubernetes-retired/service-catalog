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

package integration

import (
	"time"

	"k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"

	scfeatures "github.com/kubernetes-sigs/service-catalog/pkg/features"
)

var (
	// How often to poll for conditions
	pollInterval = 2 * time.Second

	// Default time to wait for operations to complete
	defaultTimeout = 30 * time.Second
)

// strPtr, String Pointer, returns the address of s. useful for filling struct
// fields that require a *string (for json decoding purposes).
func strPtr(s string) *string {
	return &s
}

// truePtr, Boolean Pointer with the value of true
func truePtr() *bool {
	b := true
	return &b
}

// falsePtr, Boolean Pointer with the value of false
func falsePtr() *bool {
	b := false
	return &b
}

func enableNamespacedResources() (resetFeaturesFunc func(), err error) {
	previousFeatureGate := feature.DefaultFeatureGate

	newFeatureGate := featuregate.NewFeatureGate()
	if err := newFeatureGate.Add(map[featuregate.Feature]featuregate.FeatureSpec{
		scfeatures.NamespacedServiceBroker: {Default: true, PreRelease: featuregate.Alpha},
	}); err != nil {
		return nil, err
	}
	feature.DefaultFeatureGate = newFeatureGate

	return func() {
		feature.DefaultFeatureGate = previousFeatureGate
	}, nil
}
