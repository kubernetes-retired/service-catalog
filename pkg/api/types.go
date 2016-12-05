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

package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	kunversioned "k8s.io/kubernetes/pkg/api/unversioned"
)

// +nonNamespaced=true

// Broker represents the broker resource
type Broker struct {
	kunversioned.TypeMeta
	// Non-namespaced.  The name of this resource in etcd is in ObjectMeta.Name.
	kapi.ObjectMeta

	Spec   BrokerSpec
	Status BrokerStatus
}

// BrokerSpec represents the data under the 'spec' section of a broker
type BrokerSpec struct {
	URL string

	// Auth credentials should live in an api.Secret that
	// is documented to have "username" and "password" keys
	AuthUsername string
	AuthPassword string

	// CF-specific; move to annotation
	// CFGUID is the identity of this object for use with the CF SB API.
	CFGUID string
}

// BrokerStatus represents the data under the 'status' section of a broker
type BrokerStatus struct {
	Conditions []BrokerCondition
}

// BrokerCondition represents a single condition inside a broker status
type BrokerCondition struct {
	Type    BrokerConditionType
	Status  ConditionStatus
	Reason  string
	Message string
}

// BrokerConditionType is the type that defines a broker's condition type
type BrokerConditionType string

const (
	// BrokerConditionReady is the a broker's readiness condition
	BrokerConditionReady BrokerConditionType = "Ready"
)

// ConditionStatus is the type that defines a broker's condition status
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition;
// "ConditionFalse" means a resource is not in the condition; "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	// ConditionTrue indicates that a broker's condition is true
	ConditionTrue ConditionStatus = "True"
	// ConditionFalse indicates that a broker's condition is false
	ConditionFalse ConditionStatus = "False"
	// ConditionUnknown indicates that a broker's condition is unknown
	ConditionUnknown ConditionStatus = "Unknown"
)

// ServiceClass represents the service class resource
type ServiceClass struct {
	kunversioned.TypeMeta
	kapi.ObjectMeta

	// BrokerName is the reference to the Broker that provides this service.
	BrokerName string

	Bindable      bool
	Plans         []ServicePlan
	PlanUpdatable bool // Do we support this?

	// Move to annotation
	Description string

	// CF-specific; move to annotation
	// CFGUID is the identity of this object for use with the CF SB API.
	// Immutable.
	CFGUID string

	// CF-specific; move to annotations
	CFTags                    []string
	CFRequires                []string
	CFMaxDBPerNode            string
	CFMetadata                interface{}
	CFDashboardOAuth2ClientID string
	CFDashboardSecret         string
	CFDashboardRedirectURI    string
}

// ServicePlan is a single service plan inside a service class
type ServicePlan struct {
	// CLI-friendly name of this plan
	Name string

	// Move to annotation
	Description string

	// CF-specific; move to annotation
	// CFGUID is the identity of this object for use with the CF SB API.
	// Immutable.
	CFGUID string

	// CF-specific; move to annotations
	CFMetadata interface{}
	CFFree     bool
}

// Instance represents an instance resource
type Instance struct {
	kunversioned.TypeMeta
	kapi.ObjectMeta

	Spec   InstanceSpec
	Status InstanceStatus
}

// InstanceSpec represents the data under an instance's 'spec' field
type InstanceSpec struct {
	// ServiceClassName is the reference to the ServiceClass this is an instance of.
	ServiceClassName string
	// ServicePlanName is the reference to the ServicePlan for this instance.
	PlanName string

	Parameters map[string]interface{}

	// CF-specific; move to annotation
	// CFGUID is the identity of this object for use with the CF SB API.
	// Immutable.
	CFGUID string

	// CF-specific; move to annotations
	CFCredentials   string
	CFDashboardURL  string
	CFInternalID    string // came from ville. remove? nobody likes that guy
	CFServiceID     string
	CFPlanID        string
	CFType          string
	CFSpaceGUID     string
	CFLastOperation string
}

// InstanceStatus represents the data under an instance's 'status' field
type InstanceStatus struct {
	Conditions []InstanceCondition
}

// InstanceCondition represents a condition under an instance's 'status' field
type InstanceCondition struct {
	Type    InstanceConditionType
	Status  ConditionStatus
	Reason  string
	Message string
}

// InstanceConditionType represents the name of an instance's condition
type InstanceConditionType string

const (
	// InstanceConditionProvisioning represents an instance's provisioning condition
	InstanceConditionProvisioning InstanceConditionType = "Provisioning"
	// InstanceConditionReady represents an instance's ready condition
	InstanceConditionReady InstanceConditionType = "Ready"
	// InstanceConditionProvisionFailed represents an instance's provisioning failed condition
	InstanceConditionProvisionFailed InstanceConditionType = "ProvisionFailed"
	// InstanceConditionDeprovisioning represents an instance's deprovisioning condition
	InstanceConditionDeprovisioning InstanceConditionType = "Deprovisioning"
	// InstanceConditionDeprovisionFailed represents an instance's deprovisioning failed condition
	InstanceConditionDeprovisionFailed InstanceConditionType = "DeprovisioningFailed"
)

// Binding represents a binding resource
type Binding struct {
	kunversioned.TypeMeta
	kapi.ObjectMeta

	Spec   BindingSpec
	Status BindingStatus
}

// BindingSpec represents the data under a binding's 'spec' field
type BindingSpec struct {
	// InstanceRef is the reference to the Instance this binding is to.
	InstanceRef kapi.ObjectReference
	// AppLabelSelector selects the pods in the Binding's namespace that should be injected with the results of the binding
	AppLabelSelector kapi.LabelSelector

	Parameters map[string]interface{}

	// References to objects to create
	SecretRef                 string
	ServiceRef                string
	ConfigMapRef              string
	ServiceInjectionPolicyRef string

	// CF-specific; move to annotation
	// CFGUID is the identity of this object for use with the CF SB API.
	// Immutable.
	CFGUID string

	// TODO: allow the svc consumer to tell the SIP how to expose CM and secret (env or volume)
}

// BindingStatus represents the data under a binding's 'status' field
type BindingStatus struct {
	Conditions []BindingCondition
}

// BindingCondition represents a condition under the binding's 'status' field
type BindingCondition struct {
	Type    BindingConditionType
	Status  ConditionStatus
	Reason  string
	Message string
}

// BindingConditionType represents a single binding condition
type BindingConditionType string

const (
	// BindingConditionReady represents a binding's readiness condition
	BindingConditionReady BindingConditionType = "Ready"
	// BindingConditionFailed represents a binding's failure condition
	BindingConditionFailed BindingConditionType = "Failed"
)

// ServiceInjectionPolicy represents a service injection policy resource. It is a core resource, but
// prototyped here for now
type ServiceInjectionPolicy struct {
	kunversioned.TypeMeta
	kapi.ObjectMeta

	// ServiceRef is the reference to the core Service this InjectPolicy sources.
	ServiceRef string
	// AppLabelSelector selects the pods in the SIP's namespace to apply the injection policy to.
	AppLabelSelector kapi.LabelSelector

	// TODO: the service consumer's preference on how to expose CM and secret (env or volume)
}
