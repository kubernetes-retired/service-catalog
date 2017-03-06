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
	"k8s.io/kubernetes/pkg/api/v1"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/runtime"
)

// +genclient=true
// +nonNamespaced=true

// Broker represents an entity that provides ServiceClasses for use in the
// service catalog.
type Broker struct {
	metav1.TypeMeta `json:",inline"`
	// Non-namespaced.  The name of this resource in etcd is in ObjectMeta.Name.
	v1.ObjectMeta `json:"metadata,omitempty"`

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

// BrokerCondition represents an aspect of a Broker's status.
type BrokerCondition struct {
	Type    BrokerConditionType `json:"type"`
	Status  ConditionStatus     `json:"status"`
	Reason  string              `json:"reason"`
	Message string              `json:"message"`
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
	metav1.TypeMeta `json:",inline"`
	v1.ObjectMeta   `json:"metadata,omitempty"`

	// BrokerName is the reference to the Broker that provides this service.
	// Immutable.
	BrokerName string `json:"brokerName"`

	Bindable      bool          `json:"bindable"`
	Plans         []ServicePlan `json:"plans"`
	PlanUpdatable bool          `json:"planUpdatable"` // Do we support this?

	// OSB-specific
	// OSBGUID is the identity of this object for use with the OSB API.
	// Immutable.
	OSBGUID string `json:"osbGuid"`

	// OSB-specific
	OSBTags                    []string `json:"osbTags"`
	OSBRequires                []string `json:"osbRequires"`
	OSBMaxDBPerNode            string   `json:"osbMaxDBPerNode"`
	OSBDashboardOAuth2ClientID string   `json:"osbDashboardOAuth2ClientID"`
	OSBDashboardSecret         string   `json:"osbDashboardSecret"`
	OSBDashboardRedirectURI    string   `json:"osbDashboardRedirectURI"`

	// Metadata fields
	Description         string `json:"description,omitempty"`
	DisplayName         string `json:"displayName,omitempty"`
	ImageURL            string `json:"imageUrl,omitempty"`
	LongDescription     string `json:"longDescription,omitempty"`
	ProviderDisplayName string `json:"providerDisplayName,omitempty"`
	DocumentationURL    string `json:"documentationUrl,omitempty"`
	SupportURL          string `json:"supportUrl,omitempty"`
}

// ServicePlan represents a tier of a ServiceClass.
type ServicePlan struct {
	// CLI-friendly name of this plan
	Name string `json:"name"`

	// OSB-specific
	// OSBGUID is the identity of this object for use with the OSB API.
	// Immutable.
	OSBGUID string `json:"osbGuid"`

	// OSB-specific
	OSBFree     bool     `json:"osbFree"`
	Description string   `json:"description,omitempty"`
	Bullets     []string `json:"bullets,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
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
	metav1.TypeMeta `json:",inline"`
	v1.ObjectMeta   `json:"metadata,omitempty"`

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

	// OSB-specific
	// OSBGUID is the identity of this object for use with the OSB SB API.
	// Immutable.
	OSBGUID string `json:"osbGuid"`

	// OSB-specific
	OSBCredentials   string `json:"osbCredentials"`
	OSBDashboardURL  string `json:"osbDashboardURL"`
	OSBInternalID    string `json:"osbInternalID"`
	OSBServiceID     string `json:"osbServiceID"`
	OSBPlanID        string `json:"osbPlanID"`
	OSBType          string `json:"osbType"`
	OSBSpaceGUID     string `json:"osbSpaceGUID"`
	OSBLastOperation string `json:"osbLastOperation"`
}

// InstanceStatus represents the current status of an Instance.
type InstanceStatus struct {
	Conditions []InstanceCondition `json:"conditions"`
}

// InstanceCondition represents an aspect of an Instance's status.
type InstanceCondition struct {
	Type    InstanceConditionType `json:"type"`
	Status  ConditionStatus       `json:"status"`
	Reason  string                `json:"reason"`
	Message string                `json:"message"`
}

// InstanceConditionType represents a instance condition value
type InstanceConditionType string

const (
	// InstanceConditionProvisioning represents that a given instance condition is in
	// provisioning state
	InstanceConditionProvisioning InstanceConditionType = "Provisioning"
	// InstanceConditionReady represents that a given instance condition is in
	// ready state
	InstanceConditionReady InstanceConditionType = "Ready"
	// InstanceConditionProvisionFailed represents that a given instance condition is in
	// failed state
	InstanceConditionProvisionFailed InstanceConditionType = "ProvisionFailed"
	// InstanceConditionDeprovisioning represents that a given instance condition is in
	// deprovisioning state
	InstanceConditionDeprovisioning InstanceConditionType = "Deprovisioning"
	// InstanceConditionDeprovisionFailed represents that a given instance condition is in
	// deprovision failed state
	InstanceConditionDeprovisionFailed InstanceConditionType = "DeprovisioningFailed"
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
	metav1.TypeMeta `json:",inline"`
	v1.ObjectMeta   `json:"metadata,omitempty"`

	Spec   BindingSpec   `json:"spec"`
	Status BindingStatus `json:"status"`
}

// BindingSpec represents a description of a Binding.
type BindingSpec struct {
	// InstanceRef is the reference to the Instance this binding is to.
	// Immutable.
	InstanceRef v1.ObjectReference `json:"instanceRef"`
	// AppLabelSelector selects the pods in the Binding's namespace that
	// should be injected with the results of the binding.  Immutable.
	AppLabelSelector metav1.LabelSelector `json:"appLabelSelector"`

	// Parameters is a YAML representation of the properties to be
	// passed to the underlying broker.
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`

	// Names of subordinate objects to create
	SecretName    string `json:"secretName"`
	ServiceName   string `json:"serviceName"`
	ConfigMapName string `json:"configMapName"`
	// Placeholder for future SIP support
	// ServiceInjectionPolicyName string `json:"serviceInjectionPolicyName"`

	// OSB-specific
	// OSBGUID is the identity of this object for use with the OSB API.
	// Immutable.
	OSBGUID string `json:"osbGuid"`

	// TODO: allow the svc consumer to tell the SIP how to expose CM and secret (env or volume)
}

// BindingStatus represents the current status of a Binding.
type BindingStatus struct {
	Conditions []BindingCondition `json:"conditions"`
}

// BindingCondition represents an aspect of a Binding's status.
type BindingCondition struct {
	Type    BindingConditionType `json:"type"`
	Status  ConditionStatus      `json:"status"`
	Reason  string               `json:"reason"`
	Message string               `json:"message"`
}

// BindingConditionType represents a binding condition value
type BindingConditionType string

const (
	// BindingConditionReady represents a binding condition is in ready state
	BindingConditionReady BindingConditionType = "Ready"
	// BindingConditionFailed represents a binding condition is in failed state
	BindingConditionFailed BindingConditionType = "Failed"
)
