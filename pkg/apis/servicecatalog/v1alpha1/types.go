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
	kunversioned "k8s.io/client-go/1.5/pkg/api/unversioned"
	kapi "k8s.io/client-go/1.5/pkg/api/v1"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
)

// +nonNamespaced=true

// Broker represents an entity that provides ServiceClasses for use in the
// service catalog.
type Broker struct {
	metav1.TypeMeta
	// Non-namespaced.  The name of this resource in etcd is in ObjectMeta.Name.
	kapi.ObjectMeta

	Spec   BrokerSpec
	Status BrokerStatus
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

	// CF-specific; move to annotation
	// CFGUID is the identity of this object for use with the CF SB API.
	CFGUID string
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

// ServiceClass represents an offering in the service catalog.
type ServiceClass struct {
	kunversioned.TypeMeta
	kapi.ObjectMeta

	// BrokerName is the reference to the Broker that provides this service.
	// Immutable.
	BrokerName string

	Bindable      bool
	Plans         []ServicePlan
	PlanUpdatable bool // Do we support this?

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

// ServicePlan represents a tier of a ServiceClass.
type ServicePlan struct {
	// CLI-friendly name of this plan
	Name string

	// CF-specific; move to annotation
	// CFGUID is the identity of this object for use with the CF SB API.
	// Immutable.
	CFGUID string

	// CF-specific; move to annotations
	CFMetadata interface{}
	CFFree     bool
}

// Instance represents a provisioned instance of a ServiceClass.
type Instance struct {
	kunversioned.TypeMeta
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

	Parameters map[string]interface{}

	// CF-specific; move to annotation
	// CFGUID is the identity of this object for use with the CF SB API.
	// Immutable.
	CFGUID string

	// CF-specific; move to annotations
	CFCredentials   string
	CFDashboardURL  string
	CFInternalID    string
	CFServiceID     string
	CFPlanID        string
	CFType          string
	CFSpaceGUID     string
	CFLastOperation string
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

// Binding represents a "used by" relationship between an application and an
// Instance.
type Binding struct {
	kunversioned.TypeMeta
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
	AppLabelSelector kunversioned.LabelSelector

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
