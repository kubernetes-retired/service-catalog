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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api/v1"
)

// +genclient=true
// +nonNamespaced=true

// Broker represents an entity that provides ServiceClasses for use in the
// service catalog.
type Broker struct {
	metav1.TypeMeta `json:",inline"`
	// Non-namespaced.  The name of this resource in etcd is in ObjectMeta.Name.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BrokerSpec   `json:"spec"`
	Status BrokerStatus `json:"status"`
}

// BrokerList is a list of Brokers.
type BrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Broker `json:"items"`
}

// BrokerSpec represents a description of a Broker.
type BrokerSpec struct {
	// The URL to communicate with the Broker via..
	URL string `json:"url"`

	// AuthSecret is a reference to a Secret containing auth information the
	// catalog should use to authenticate to this Broker.
	AuthSecret *v1.ObjectReference `json:"authSecret,omitempty"`
}

// BrokerStatus represents the current status of a Broker.
type BrokerStatus struct {
	Conditions []BrokerCondition `json:"conditions"`
}

// BrokerCondition contains condition information for a Broker.
type BrokerCondition struct {
	// Type of the condition, currently ('Ready').
	Type BrokerConditionType `json:"type"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string `json:"reason"`
	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string `json:"message"`
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
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceClass `json:"items"`
}

// +genclient=true
// +nonNamespaced=true

// ServiceClass represents an offering in the service catalog.
type ServiceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// BrokerName is the reference to the Broker that provides this service.
	// Immutable.
	BrokerName string `json:"brokerName"`

	// Bindable indicates whether a user can create bindings to an instance of
	// this service. ServicePlan has an optional field called Bindable which
	// overrides the value of this field.
	Bindable      bool          `json:"bindable"`
	Plans         []ServicePlan `json:"plans"`
	PlanUpdatable bool          `json:"planUpdatable"` // Do we support this?

	// ExternalID is the identity of this object for use with the OSB API.
	// Immutable.
	ExternalID string `json:"externalID"`

	// OSB-specific
	AlphaTags []string `json:"alphaTags,omitempty"`
	Requires  []string `json:"requires,omitempty"`
	// Description is a short description of the service.
	Description string `json:"description"`

	// ExternalMetadata fields
	ExternalMetadata *runtime.RawExtension `json:"externalMetadata, omitempty"`
}

// ServicePlan represents a tier of a ServiceClass.
type ServicePlan struct {
	// CLI-friendly name of this plan
	Name string `json:"name"`

	// Description is a short description of the plan.
	Description string `json:"description"`

	// ExternalID is the identity of this object for use with the OSB API.
	// Immutable.
	ExternalID string `json:"externalID"`

	// Bindable indicates whether this users can create bindings to an
	// Instance using this plan.  If set, overrides the value of the
	// ServiceClass.Bindable field.
	Bindable         *bool                 `json:"bindable,omitempty"`
	Free             bool                  `json:"free"`
	ExternalMetadata *runtime.RawExtension `json:"externalMetadata, omitempty"`
}

// InstanceList is a list of instances
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Instance `json:"items"`
}

// +genclient=true

// Instance represents a provisioned instance of a ServiceClass.
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec"`
	Status InstanceStatus `json:"status"`
}

// InstanceSpec represents a description of an Instance.
type InstanceSpec struct {
	// ServiceClassName is the reference to the ServiceClass this is an
	// instance of.  Immutable.
	ServiceClassName string `json:"serviceClassName"`
	// ServicePlanName is the reference to the ServicePlan for this instance.
	PlanName string `json:"planName"`

	// Parameters is a YAML representation of the properties to be
	// passed to the underlying broker.
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`

	// ExternalID is the identity of this object for use with the OSB SB API.
	// Immutable.
	ExternalID string `json:"externalID"`

	// Checksum is the checksum of the InstanceSpec that was last successfully
	// reconciled against the broker.
	Checksum *string `json:"checksum,omitempty"`
}

// InstanceStatus represents the current status of an Instance.
type InstanceStatus struct {
	Conditions []InstanceCondition `json:"conditions"`

	// AsyncOpInProgress is set to true if there is an ongoing async operation
	// against this Service Instance in progress.
	AsyncOpInProgress bool `json:"asyncOpInProgress"`

	// LastOperation is the string that the broker may have returned when
	// an async operation started, it should be sent back to the broker
	// on poll requests as a query param.
	LastOperation *string `json:"lastOperation,omitempty"`

	// DashboardURL is the URL of a web-based management user interface for
	// the service instance
	DashboardURL *string `json:"dashboardURL,omitempty"`
}

// InstanceCondition contains condition information for an Instance.
type InstanceCondition struct {
	// Type of the condition, currently ('Ready').
	Type InstanceConditionType `json:"type"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string `json:"reason"`
	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string `json:"message"`
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
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Binding `json:"items"`
}

// +genclient=true

// Binding represents a "used by" relationship between an application and an
// Instance.
type Binding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BindingSpec   `json:"spec"`
	Status BindingStatus `json:"status"`
}

// BindingSpec represents a description of a Binding.
type BindingSpec struct {
	// InstanceRef is the reference to the Instance this binding is to.
	// Immutable.
	InstanceRef v1.LocalObjectReference `json:"instanceRef"`

	// Parameters is a YAML representation of the properties to be
	// passed to the underlying broker.
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`

	// Names of subordinate objects to create
	SecretName string `json:"secretName"`

	// ExternalID is the identity of this object for use with the OSB API.
	// Immutable.
	ExternalID string `json:"externalID"`

	// Checksum is the checksum of the BindingSpec that was last successfully
	// reconciled against the broker.
	Checksum *string `json:"checksum,omitempty"`
}

// BindingStatus represents the current status of a Binding.
type BindingStatus struct {
	Conditions []BindingCondition `json:"conditions"`
}

// BindingCondition condition information for a Binding.
type BindingCondition struct {
	// Type of the condition, currently ('Ready').
	Type BindingConditionType `json:"type"`
	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`
	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string `json:"reason"`
	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string `json:"message"`
}

// BindingConditionType represents a binding condition value
type BindingConditionType string

const (
	// BindingConditionReady represents a binding condition is in ready state
	BindingConditionReady BindingConditionType = "Ready"
	// BindingConditionFailed represents a binding condition is in failed state
	BindingConditionFailed BindingConditionType = "Failed"
)
