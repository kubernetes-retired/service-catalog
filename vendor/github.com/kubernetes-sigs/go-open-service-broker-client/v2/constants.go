/*
Copyright 2019 The Kubernetes Authors.

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

package v2

const (
	// AcceptsIncomplete is the name of a query parameter that indicates that
	// the client allows a request to complete asynchronously.
	AcceptsIncomplete = "accepts_incomplete"

	// VarKeyInstanceID is the name to use for a mux var representing an
	// instance ID.
	VarKeyInstanceID = "instance_id"

	// VarKeyBindingID is the name to use for a mux var representing a binding
	// ID.
	VarKeyBindingID = "binding_id"

	// VarKeyServiceID is the name to use for a mux var representing a service ID.
	VarKeyServiceID = "service_id"

	// VarKeyPlanID is the name to use for a mux var representing a plan ID.
	VarKeyPlanID = "plan_id"

	// VarKeyOperation is the name to use for a mux var representing an
	// operation.
	VarKeyOperation = "operation"

	// PlatformKubernetes is the name for Kubernetes in the Platform field of
	// OriginatingIdentity.
	PlatformKubernetes = "kubernetes"

	// PlatformCloudFoundry is the name for Cloud Foundry in the Platform field
	// of OriginatingIdentity.
	PlatformCloudFoundry = "cloudfoundry"
)
