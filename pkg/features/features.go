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

package features

import (
	utilfeature "k8s.io/apiserver/pkg/util/feature"
)

const (
	// Every feature gate should add const here following this template:
	//
	// MyFeature
	// // owner: @username
	// // alpha: v0.xxx
	// MyFeature utilfeature.Feature = "MyFeature"

	// PodPreset controls whether PodPreset resource is enabled or not in the
	// API server.
	// owner: @droot
	// alpha: v0.xx
	PodPreset utilfeature.Feature = "PodPreset"
)

// defaultServiceCatalogFeatureGates consists of all known service catalog specific feature keys.
// To add a new feature, define a key for it above and add it here. The features will be
// available throughout service catalog binaries.
var defaultServiceCatalogFeatureGates = map[utilfeature.Feature]utilfeature.FeatureSpec{
	PodPreset: {Default: false, PreRelease: utilfeature.Alpha},
}

// Initialize initalizes the feature gates for service catalog features.
func Initialize() {
	utilfeature.DefaultFeatureGate.Add(defaultServiceCatalogFeatureGates)
}
