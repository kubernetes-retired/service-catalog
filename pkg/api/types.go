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
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type Broker struct {
	unversioned.TypeMeta
	kapi.ObjectMeta // not namespaced. document

	Spec   BrokerSpec
	Status BrokerStatus
}

type BrokerSpec struct {
	BrokerURL    string
	AuthUsername string
	AuthPassword string
	GUID         string
}

type BrokerStatus struct {
	State BrokerState
}

type BrokerState string

const (
	BrokerStatePending   BrokerState = "PENDING"
	BrokerStateAvailable BrokerState = "AVAILABLE"
)

type ServiceClass struct {
	// BOILERPLATE VILLE DOUG NOW YOU ARE HAPPY

	// WATCH THE MAIL FOR MY INVOICE

	ID              string
	Description     string
	Bindable        bool
	PlanUpdatable   bool // TODO: do we support this? document if we don't
	Tags            []string
	Requires        []string
	Metadata        interface{}
	Plans           []ServicePlan
	DashboardClient interface{}
	BrokerName      string // broker object name
}

type ServicePlan struct {
	ID          string
	Name        string
	Description string
	Metadata    interface{}
	Free        bool
	// OTHER?
}

type Instance struct {
	// boilerplate

	Spec   InstanceSpec
	Status InstanceStatus
}

type InstanceSpec struct {
	ServiceClassName string // name of service class resource
	PlanName         string
	InstanceGUID     string // may move this to an annotation
	Credentials      string // this is legacy CF stuff
	DashboardURL     string
	InternalID       string // came from ville. remove? nobody likes that guy
	CFServiceID      string
	CFPlanID         string
	CFType           string // may move this to an annotation
	CFSpaceGUID      string // may move this to an annotation
	CFLastOperation  string // TODO: talk about supporting async provision
	CFParameters     map[string]interface{}
}

type InstanceStatus struct {
	Status string // TODO: make this an "enum" (constant + type)
}

type Binding struct {
	// boilerplate

	Spec   BindingSpec
	Status BindingStatus
}

type BindingSpec struct {
	BindingGUID               string               // may move this to annotation
	AppLabelSelector          kapi.LabelSelector   // this is the "from"
	InstanceRef               kapi.ObjectReference // this is the "to". the controller can follow this pointer to get the instance ID
	Parameters                map[string]interface{}
	SecretRef                 string
	ServiceRef                string
	ConfigMapRef              string
	ServiceInjectionPolicyRef string
	// for later: allow the svc consumer to tell the SIP how to expose CM and secret (env or volume)
}

type BindingStatus struct {
	Status string // either pending or bound
}

type ServiceInjectionPolicy struct {
	ServiceRef       string
	AppLabelSelector kapi.LabelSelector
	// for later: the service consumer's preference on how to expose CM and secret (env or volume)
}
