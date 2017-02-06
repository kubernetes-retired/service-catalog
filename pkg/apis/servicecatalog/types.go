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
	kapi "k8s.io/kubernetes/pkg/api"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/runtime"
)

// TODO: all metadata and parametersfields need to be refactored to real
// types; skipping for now to get very large generation PR in.

// +genclient=true
// +nonNamespaced=true

// Broker represents an entity that provides ServiceClasses for use in the
// service catalog.
type Broker struct {
	metav1.TypeMeta
	kapi.ObjectMeta

	Spec   BrokerSpec `json:"spec"`
	Status BrokerStatus
}

// BrokerList is a list of Brokers.
type BrokerList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Broker
}

const (
	// DescriptionKey is the key of an annotation that holds the brief
	// description of an API resource
	DescriptionKey = "alpha.service-catalog.kubernetes.io/description"
)

// BrokerSpec represents a description of a Broker.
type BrokerSpec struct {
	// The URL to communicate with the Broker via..
	URL string

	// Auth credentials should live in an api.Secret that
	// is documented to have "username" and "password" keys
	AuthUsername string
	AuthPassword string

	// OSB-specific
	// OSBGUID is the identity of this object for use with the OSB API.
	OSBGUID string
}

// BrokerStatus represents the current status of a Broker.
type BrokerStatus struct {
	Conditions []BrokerCondition
}

// BrokerCondition represents an aspect of a Broker's status.
type BrokerCondition struct {
	Type    BrokerConditionType
	Status  ConditionStatus
	Reason  string
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

// ServiceClass represents an offering in the service catalog.
type ServiceClass struct {
	metav1.TypeMeta
	kapi.ObjectMeta

	// BrokerName is the reference to the Broker that provides this service.
	// Immutable.
	BrokerName string

	Bindable      bool
	Plans         []ServicePlan
	PlanUpdatable bool // Do we support this?

	// OSB-specific
	// OSBGUID is the identity of this object for use with the OSB API.
	// Immutable.
	OSBGUID string

	// OSB-specific
	OSBTags                    []string
	OSBRequires                []string
	OSBMaxDBPerNode            string
	OSBDashboardOAuth2ClientID string
	OSBDashboardSecret         string
	OSBDashboardRedirectURI    string

	// Metadata fields
	Description         string
	DisplayName         string
	ImageURL            string
	LongDescription     string
	ProviderDisplayName string
	DocumentationURL    string
	SupportURL          string
}

// ServicePlan represents a tier of a ServiceClass.
type ServicePlan struct {
	// CLI-friendly name of this plan
	Name string

	// OSB-specific
	// OSBGUID is the identity of this object for use with the OSB API.
	// Immutable.
	OSBGUID string

	// OSB-specific
	OSBFree     bool
	Description string
	Bullets     []string
	DisplayName string

	// TODO: add costs
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
	kapi.ObjectMeta

	Spec   InstanceSpec `json:"spec"`
	Status InstanceStatus
}

// InstanceSpec represents a description of an Instance.
type InstanceSpec struct {
	// ServiceClassName is the reference to the ServiceClass this is an
	// instance of.  Immutable.
	ServiceClassName string `json:"serviceClassName"`
	// ServicePlanName is the reference to the ServicePlan for this instance.
	PlanName string `json:"planName"`

	Parameters map[string]runtime.Object

	// OSB-specific
	// OSBGUID is the identity of this object for use with the OSB API.
	// Immutable.
	OSBGUID string

	// OSB-specific
	OSBCredentials   string
	OSBDashboardURL  string
	OSBInternalID    string
	OSBServiceID     string
	OSBPlanID        string
	OSBType          string
	OSBSpaceGUID     string
	OSBLastOperation string
}

// InstanceStatus represents the current status of an Instance.
type InstanceStatus struct {
	Conditions []InstanceCondition
}

// InstanceCondition represents an aspect of an Instance's status.
type InstanceCondition struct {
	Type    InstanceConditionType
	Status  ConditionStatus
	Reason  string
	Message string
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
	metav1.TypeMeta
	metav1.ListMeta

	Items []Binding
}

// +genclient=true

// Binding represents a "used by" relationship between an application and an
// Instance.
type Binding struct {
	metav1.TypeMeta
	kapi.ObjectMeta

	Spec   BindingSpec `json:"spec"`
	Status BindingStatus
}

// BindingSpec represents a description of a Binding.
type BindingSpec struct {
	// InstanceRef is the reference to the Instance this binding is to.
	// Immutable.
	InstanceRef kapi.ObjectReference `json:"instanceRef"`
	// AppLabelSelector selects the pods in the Binding's namespace that
	// should be injected with the results of the binding.  Immutable.
	AppLabelSelector metav1.LabelSelector

	Parameters map[string]runtime.Object

	// Names of subordinate objects to create
	SecretName    string
	ServiceName   string `json:"serviceName"`
	ConfigMapName string
	// Placeholder for future SIP support
	// ServiceInjectionPolicyName string

	// OSB-specific
	// OSBGUID is the identity of this object for use with the OSB API.
	// Immutable.
	OSBGUID string

	// TODO: allow the svc consumer to tell the SIP how to expose CM and secret (env or volume)
}

// BindingStatus represents the current status of a Binding.
type BindingStatus struct {
	Conditions []BindingCondition
}

// BindingCondition represents an aspect of a Binding's status.
type BindingCondition struct {
	Type    BindingConditionType
	Status  ConditionStatus
	Reason  string
	Message string
}

// BindingConditionType represents a binding condition value
type BindingConditionType string

const (
	// BindingConditionReady represents a binding condition is in ready state
	BindingConditionReady BindingConditionType = "Ready"
	// BindingConditionFailed represents a binding condition is in failed state
	BindingConditionFailed BindingConditionType = "Failed"
	// BindingConditionUninjected represents a binding condition that the binding credentials have
	// been deleted. It is the first condition recorded when a binding is removed
	BindingConditionUninjected BindingConditionType = "Uninjected"
	// BindingConditionUnbound represents a binding condition that the service catalog has performed
	// the unbind operation on the backing CF broker. It is the second condition recorded when a
	// binding is removed
	BindingConditionUnbound BindingConditionType = "Unbound"
	// BindingConditionDeleted represents a binding condition that the service catalog has
	// intentionally deleted the binding. It is the third condition recorded when a binding is
	// removed
	BindingConditionDeleted BindingConditionType = "Deleted"
)
