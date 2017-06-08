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

package fake

const (
	// Namespace is a name used for test namespaces
	Namespace = "test-ns"
	// NamespaceUID is a UID used for test namespaces
	NamespaceUID = "test-ns-uid"

	// BrokerURL is the URL used for test brokers
	BrokerURL = "http://example.com"
	// BrokerName is the name used for test brokers
	BrokerName = "test-broker"

	// ServiceClassName is the name used for test service classes
	ServiceClassName = "test-serviceclass"
	// ServiceClassGUID is the GUID used for test service classes
	ServiceClassGUID = "SCGUID"
	// PlanName is the name used for test plans
	PlanName = "test-plan"
	// PlanGUID is the GUID used for test plans
	PlanGUID = "PGUID"
	//NonBindablePlanName is the name used for test plans that should not be bindable
	NonBindablePlanName = "test-unbindable-plan"
	// NonBindablePlanGUID is the GUID used for test plans that should not be bindable
	NonBindablePlanGUID = "UNBINDABLE-PLAN"

	// InstanceName is a name used for test instances
	InstanceName = "test-instance"
	// InstanceGUID is the GUID used for test instances
	InstanceGUID = "IGUID"
)
