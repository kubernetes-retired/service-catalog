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
)

// +nonNamespaced=true

// Broker represents an entity that provides ServiceClasses for use in the
// service catalog.
type Broker struct {
	metav1.TypeMeta
	kapi.ObjectMeta

	Spec   BrokerSpec
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

	// UserFacingInfo contains supplemental information that should be shown
	// to users in various user interface contexts.
	UserFacingInfo ServiceClassUserFacingInfo

	// DashboardClientInfo contains information necessary to enable dashboard
	// SSO for this service class.
	DashboardClientInfo DashboardClientInfo

	// OpenServiceBrokerServiceClassFields contains fields specific to the OSB
	// API.
	OpenServiceBrokerServiceClassFields
}

// ServiceClassUserFacingInfo contains information that should be displayed to
// users in various user interface contexts.
type ServiceClassUserFacingInfo struct {
	Tags                []string
	Description         string
	DisplayName         string
	ImageURL            string
	LongDescription     string
	ProviderDisplayName string
	DocumentationURL    string
	SupportURL          string
}

// DashboardClientInfo holds the information necessary to enable dashboard SSO
// for a ServiceClass.
type DashboardClientInfo struct {
	ID        string
	SecretRef kapi.ObjectReference
}

// OpenServiceBrokerServiceClassFields contains fields specific to the OSB API.
type OpenServiceBrokerServiceClassFields struct {
	// GUID is the identity of this object for use with the OSB API.
	// Immutable.
	GUID         string
	Requires     []string
	MaxDBPerNode string
}

// ServicePlan represents a tier of a ServiceClass.
type ServicePlan struct {
	// CLI-friendly name of this plan
	Name string

	// UserFacingInfo contains supplemental information that should be shown
	// to users in various user interface contexts.
	UserFacingInfo ServicePlanUserFacingInfo

	// OpenServiceBrokerServicePlanFields contains fields specific to the OSB
	// API.
	OpenServiceBrokerServicePlanFields
}

// ServicePlanUserFacingInfo contains information that should be displayed to
// users in various user interface contexts.
type ServicePlanUserFacingInfo struct {
	// Free indicates whether this plan is free of charge.
	Free bool
	// Description is the user-facing description of this plan.
	Description string
	// DisplayName is the user-facing long-form name to display for this plan.
	DisplayName string
	// Features of this plan, to be displayed in a bulleted list.
	Bullets []string
	// Costs associated with this plan.
	Costs []ServicePlanCost
}

// ServicePlanCost contains information about the costs associated with a
// ServicePlan.
type ServicePlanCost struct {
	Amount string
	Unit   string
}

// TODO: determine correct way of handling currency in a k8s API.

// OpenServiceBrokerServicePlanFields contains fields specific to the OSB API.
type OpenServiceBrokerServicePlanFields struct {
	// GUID is the identity of this object for use with the OSB API.
	// Immutable.
	OSBGUID string
}

// InstanceList is a list of instances.
type InstanceList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Instance
}

// Instance represents a provisioned instance of a ServiceClass.
type Instance struct {
	metav1.TypeMeta
	kapi.ObjectMeta

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
	// Parameters is a set of parameters to pass to the backing API.
	Parameters map[string]string
	// OpenServiceBrokerFields contains fields specific to the OSB API.
	OpenServiceBrokerInstanceFields
}

type OpenServiceBrokerInstanceFields struct {
	// GUID is the identity of this object for use with the OSB API.
	// Immutable.
	GUID             string
	DashboardURL     string
	InternalID       string
	SpaceGUID        string
	OrganizationGUID string
	LastOperation    string
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

// Binding represents a "used by" relationship between an application and an
// Instance.
type Binding struct {
	metav1.TypeMeta
	kapi.ObjectMeta

	Spec   BindingSpec
	Status BindingStatus
}

// BindingSpec represents a description of a Binding.
type BindingSpec struct {
	// InstanceRef is the reference to the Instance this binding is to.
	// Immutable.
	InstanceRef kapi.ObjectReference
	// AppLabelSelector selects the pods in the Binding's namespace that
	// should be injected with the results of the binding.  Immutable.
	AppLabelSelector metav1.LabelSelector

	Parameters map[string]string

	// Names of subordinate objects to create
	SecretName    string
	ServiceName   string
	ConfigMapName string
	// OpenServiceBrokerBindingFields contains fields specific to the OSB API.
	OpenServiceBrokerBindingFields

	// Placeholders for future SIP support
	// ServiceInjectionPolicyName string `json:"serviceInjectionPolicyName"`
	// ServiceInjectionPolicySpec
}

type OpenServiceBrokerBindingFields struct {
	// GUID is the identity of this object for use with the OSB API.
	// Immutable.
	GUID string
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
)
