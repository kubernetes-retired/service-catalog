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

type Broker struct {
	kunversioned.TypeMeta
	// Non-namespaced.  The name of this resource in etcd is in ObjectMeta.Name.
	kapi.ObjectMeta

	Spec   BrokerSpec
	Status BrokerStatus
}

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

type BrokerStatus struct {
	State BrokerState
}

type BrokerState string

const (
	BrokerStatePending   BrokerState = "Pending"
	BrokerStateAvailable BrokerState = "Available"
	BrokerStateFailed    BrokerState = "Failed"
)

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

type Instance struct {
	kunversioned.TypeMeta
	kapi.ObjectMeta

	Spec   InstanceSpec
	Status InstanceStatus
}

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

type InstanceStatus struct {
	State InstanceState
}

type InstanceState string

const (
	InstanceStatePending     InstanceState = "Pending"
	InstanceStateProvisioned InstanceState = "Provisioned"
	InstanceStateFailed      InstanceState = "Failed"
)

type Binding struct {
	kunversioned.TypeMeta
	kapi.ObjectMeta

	Spec   BindingSpec
	Status BindingStatus
}

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

type BindingStatus struct {
	State BindingState
}

type BindingState string

const (
	BindingStatePending BindingState = "Pending"
	BindingStateBound   BindingState = "Bound"
	BindingStateFailed  BindingState = "Failed"
)

// Core resource; prototype here for now
type ServiceInjectionPolicy struct {
	kunversioned.TypeMeta
	kapi.ObjectMeta

	// ServiceRef is the reference to the core Service this InjectPolicy sources.
	ServiceRef string
	// AppLabelSelector selects the pods in the SIP's namespace to apply the injection policy to.
	AppLabelSelector kapi.LabelSelector

	// TODO: the service consumer's preference on how to expose CM and secret (env or volume)
}
