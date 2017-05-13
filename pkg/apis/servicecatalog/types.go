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

package servicecatalog

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api/v1"
)

// TODO: all metadata and parametersfields need to be refactored to real
// types; skipping for now to get very large generation PR in.

// +genclient=true
// +nonNamespaced=true

// Broker represents an entity that provides ServiceClasses for use in the
// service catalog.
type Broker struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   BrokerSpec
	Status BrokerStatus
}

// BrokerList is a list of Brokers.
type BrokerList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Broker
}

// BrokerSpec represents a description of a Broker.
type BrokerSpec struct {
	// The URL to communicate with the Broker via..
	URL string

	// AuthSecret is a reference to a Secret containing auth information the
	// catalog should use to authenticate to this Broker.
	AuthSecret *v1.ObjectReference
}

// BrokerStatus represents the current status of a Broker.
type BrokerStatus struct {
	Conditions []BrokerCondition
}

// BrokerCondition contains condition information for a Broker.
type BrokerCondition struct {
	// Type of the condition, currently ('Ready').
	Type BrokerConditionType
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus
	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time
	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string
	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string
}

// BrokerConditionType represents a broker condition value
type BrokerConditionType string

const (
	// BrokerConditionReady represents the fact that a given broker condition is in ready state
	BrokerConditionReady BrokerConditionType = "Ready"
)

// ConditionStatus represents a condition's status
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in
// the condition; "ConditionFalse" means a resource is not in the condition;
// "ConditionUnknown" means kubernetes can't decide if a resource is in the
// condition or not. In the future, we could add other intermediate
// conditions, e.g. ConditionDegraded.
const (
	// ConditionTrue represents the fact that a given condition is true
	ConditionTrue ConditionStatus = "True"
	// ConditionFalse represents the fact that a given condition is false
	ConditionFalse ConditionStatus = "False"
	// ConditionUnknown represents the fact that a given condition is unknown
	ConditionUnknown ConditionStatus = "Unknown"
)

// ServiceClassList is a list of ServiceClasses
type ServiceClassList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []ServiceClass
}

// +genclient=true
// +nonNamespaced=true

// ServiceClass represents an offering in the service catalog.
type ServiceClass struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// BrokerName is the reference to the Broker that provides this service.
	// Immutable.
	BrokerName string

	// Bindable indicates whether a user can create bindings to an instance of
	// this service. ServicePlan has an optional field called Bindable which
	// overrides the value of this field.
	Bindable      bool
	Plans         []ServicePlan
	PlanUpdatable bool // Do we support this?

	// ExternalID is the identity of this object for use with the OSB API.
	// Immutable.
	ExternalID string

	// OSB-specific
	OSBTags                    []string
	OSBRequires                []string
	OSBMaxDBPerNode            *string
	OSBDashboardOAuth2ClientID *string
	OSBDashboardSecret         *string
	OSBDashboardRedirectURI    *string
	// Description is a short description of the service.
	Description string
	OSBMetadata *runtime.RawExtension
}

// ServicePlan represents a tier of a ServiceClass.
type ServicePlan struct {
	// CLI-friendly name of this plan
	Name string

	// ExternalID is the identity of this object for use with the OSB API.
	// Immutable.
	ExternalID string

	// Bindable indicates whether this users can create bindings to an
	// Instance using this plan.  If set, overrides the value of the
	// ServiceClass.Bindable field.
	Bindable *bool
	OSBFree  bool
	// Description is a short description of the plan.
	Description string
	OSBMetadata *runtime.RawExtension
}

// InstanceList is a list of instances.
type InstanceList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Instance
}

// +genclient=true

// Instance represents a provisioned instance of a ServiceClass.
type Instance struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   InstanceSpec
	Status InstanceStatus
}

// InstanceSpec represents a description of an Instance.
type InstanceSpec struct {
	// ServiceClassName is the reference to the ServiceClass this is an
	// instance of.  Immutable.
	ServiceClassName string
	// ServicePlanName is the reference to the ServicePlan for this instance.
	PlanName string

	// Parameters is a YAML representation of the properties to be
	// passed to the underlying broker.
	Parameters *runtime.RawExtension

	// ExternalID is the identity of this object for use with the OSB API.
	// Immutable.
	ExternalID string

	// Checksum is the checksum of the InstanceSpec that was last successfully
	// reconciled against the broker.
	Checksum *string
}

// InstanceStatus represents the current status of an Instance.
type InstanceStatus struct {
	Conditions []InstanceCondition
	// AsyncOpInProgress is set to true if there is an ongoing async operation
	// against this Service Instance in progress.
	AsyncOpInProgress bool

	// LastOperation is the string that the broker may have returned when
	// an async operation started, it should be sent back to the broker
	// on poll requests as a query param.
	LastOperation *string

	// DashboardURL is the URL of a web-based management user interface for
	// the service instance
	DashboardURL *string
}

// InstanceCondition contains condition information for an Instance.
type InstanceCondition struct {
	// Type of the condition, currently ('Ready').
	Type InstanceConditionType
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus
	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time
	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string
	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string
}

// InstanceConditionType represents a instance condition value
type InstanceConditionType string

const (
	// InstanceConditionReady represents that a given instance condition is in
	// ready state
	InstanceConditionReady InstanceConditionType = "Ready"
)

// BindingList is a list of Bindings
type BindingList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Binding
}

// +genclient=true

// Binding represents a "used by" relationship between an application and an
// Instance.
type Binding struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   BindingSpec
	Status BindingStatus
}

// BindingSpec represents a description of a Binding.
type BindingSpec struct {
	// InstanceRef is the reference to the Instance this binding is to.
	// Immutable.
	InstanceRef v1.LocalObjectReference

	// Parameters is a YAML representation of the properties to be
	// passed to the underlying broker.
	Parameters *runtime.RawExtension

	// Names of subordinate objects to create
	SecretName string

	// ExternalID is the identity of this object for use with the OSB API.
	// Immutable.
	ExternalID string

	// Checksum is the checksum of the BindingSpec that was last successfully
	// reconciled against the broker.
	Checksum *string
}

// BindingStatus represents the current status of a Binding.
type BindingStatus struct {
	Conditions []BindingCondition
}

// BindingCondition condition information for a Binding.
type BindingCondition struct {
	// Type of the condition, currently ('Ready').
	Type BindingConditionType
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus
	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time
	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string
	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string
}

// BindingConditionType represents a binding condition value
type BindingConditionType string

const (
	// BindingConditionReady represents a binding condition is in ready state
	BindingConditionReady BindingConditionType = "Ready"
)
