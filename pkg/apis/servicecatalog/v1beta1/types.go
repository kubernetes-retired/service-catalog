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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServiceBroker represents an entity that provides
// ClusterServiceClasses for use in the service catalog.
// +k8s:openapi-gen=x-kubernetes-print-columns:custom-columns=NAME:.metadata.name,URL:.spec.url
type ClusterServiceBroker struct {
	metav1.TypeMeta `json:",inline"`

	// Non-namespaced.  The name of this resource in etcd is in ObjectMeta.Name.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of the broker.
	// +optional
	Spec ClusterServiceBrokerSpec `json:"spec,omitempty"`

	// Status represents the current status of a broker.
	// +optional
	Status ClusterServiceBrokerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServiceBrokerList is a list of Brokers.
type ClusterServiceBrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterServiceBroker `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBroker represents an entity that provides
// ServiceClasses for use in the service catalog.
// +k8s:openapi-gen=x-kubernetes-print-columns:custom-columns=NAME:.metadata.name,URL:.spec.url
type ServiceBroker struct {
	metav1.TypeMeta `json:",inline"`

	// The name of this resource in etcd is in ObjectMeta.Name.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of the broker.
	// +optional
	Spec ServiceBrokerSpec `json:"spec,omitempty"`

	// Status represents the current status of a broker.
	// +optional
	Status ServiceBrokerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBrokerList is a list of Brokers.
type ServiceBrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceBroker `json:"items"`
}

// CommonServiceBrokerSpec represents a description of a Broker.
type CommonServiceBrokerSpec struct {
	// URL is the address used to communicate with the ServiceBroker.
	URL string `json:"url"`

	// InsecureSkipTLSVerify disables TLS certificate verification when communicating with this Broker.
	// This is strongly discouraged.  You should use the CABundle instead.
	// +optional
	InsecureSkipTLSVerify bool `json:"insecureSkipTLSVerify,omitempty"`

	// CABundle is a PEM encoded CA bundle which will be used to validate a Broker's serving certificate.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`

	// RelistBehavior specifies the type of relist behavior the catalog should
	// exhibit when relisting ServiceClasses available from a broker.
	// +optional
	RelistBehavior ServiceBrokerRelistBehavior `json:"relistBehavior"`

	// RelistDuration is the frequency by which a controller will relist the
	// broker when the RelistBehavior is set to ServiceBrokerRelistBehaviorDuration.
	// Users are cautioned against configuring low values for the RelistDuration,
	// as this can easily overload the controller manager in an environment with
	// many brokers. The actual interval is intrinsically governed by the
	// configured resync interval of the controller, which acts as a minimum bound.
	// For example, with a resync interval of 5m and a RelistDuration of 2m, relists
	// will occur at the resync interval of 5m.
	RelistDuration *metav1.Duration `json:"relistDuration,omitempty"`

	// RelistRequests is a strictly increasing, non-negative integer counter that
	// can be manually incremented by a user to manually trigger a relist.
	// +optional
	RelistRequests int64 `json:"relistRequests"`

	// CatalogRestrictions is a set of restrictions on which of a broker's services
	// and plans have resources created for them.
	// +optional
	CatalogRestrictions *CatalogRestrictions `json:"catalogRestrictions,omitempty"`
}

// CatalogRestrictions is a set of restrictions on which of a broker's services
// and plans have resources created for them.
//
// Some examples of this object are as follows:
//
// This is an example of a whitelist on service externalName.
// Goal: Only list Services with the externalName of FooService and BarService,
// Solution: restrictions := ServiceCatalogRestrictions{
// 		ServiceClass: ["spec.externalName in (FooService, BarService)"]
// }
//
// This is an example of a blacklist on service externalName.
// Goal: Allow all services except the ones with the externalName of FooService and BarService,
// Solution: restrictions := ServiceCatalogRestrictions{
// 		ServiceClass: ["spec.externalName notin (FooService, BarService)"]
// }
//
// This whitelists plans called "Demo", and blacklists (but only a single element in
// the list) a service and a plan.
// Goal: Allow all plans with the externalName demo, but not AABBCC, and not a specific service by name,
// Solution: restrictions := ServiceCatalogRestrictions{
// 		ServiceClass: ["name!=AABBB-CCDD-EEGG-HIJK"]
// 		ServicePlan: ["spec.externalName in (Demo)", "name!=AABBCC"]
// }
//
// CatalogRestrictions strings have a special format similar to Label Selectors,
// except the catalog supports only a very specific property set.
//
// The predicate format is expected to be `<property><conditional><requirement>`
// Check the *Requirements type definition for which <property> strings will be allowed.
// <conditional> is allowed to be one of the following: ==, !=, in, notin
// <requirement> will be a string value if `==` or `!=` are used.
// <requirement> will be a set of string values if `in` or `notin` are used.
// Multiple predicates are allowed to be chained with a comma (,)
//
// ServiceClass allowed property names:
//   name - the value set to [Cluster]ServiceClass.Name
//   spec.externalName - the value set to [Cluster]ServiceClass.Spec.ExternalName
//   spec.externalID - the value set to [Cluster]ServiceClass.Spec.ExternalID
// ServicePlan allowed property names:
//   name - the value set to [Cluster]ServicePlan.Name
//   spec.externalName - the value set to [Cluster]ServicePlan.Spec.ExternalName
//   spec.externalID - the value set to [Cluster]ServicePlan.Spec.ExternalID
//   spec.free - the value set to [Cluster]ServicePlan.Spec.Free
//   spec.serviceClass.name - the value set to ServicePlan.Spec.ServiceClassRef.Name
//   spec.clusterServiceClass.name - the value set to ClusterServicePlan.Spec.ClusterServiceClassRef.Name
type CatalogRestrictions struct {
	// ServiceClass represents a selector for plans, used to filter catalog re-lists.
	// +listType=set
	ServiceClass []string `json:"serviceClass,omitempty"`
	// ServicePlan represents a selector for classes, used to filter catalog re-lists.
	// +listType=set
	ServicePlan []string `json:"servicePlan,omitempty"`
}

// ClusterServiceBrokerSpec represents a description of a Broker.
type ClusterServiceBrokerSpec struct {
	CommonServiceBrokerSpec `json:",inline"`

	// AuthInfo contains the data that the service catalog should use to authenticate
	// with the ClusterServiceBroker.
	AuthInfo *ClusterServiceBrokerAuthInfo `json:"authInfo,omitempty"`
}

// ServiceBrokerSpec represents a description of a Broker.
type ServiceBrokerSpec struct {
	CommonServiceBrokerSpec `json:",inline"`

	// AuthInfo contains the data that the service catalog should use to authenticate
	// with the ServiceBroker.
	AuthInfo *ServiceBrokerAuthInfo `json:"authInfo,omitempty"`
}

// ServiceBrokerRelistBehavior represents a type of broker relist behavior.
type ServiceBrokerRelistBehavior string

const (
	// ServiceBrokerRelistBehaviorDuration indicates that the broker will be
	// relisted automatically after the specified duration has passed.
	ServiceBrokerRelistBehaviorDuration ServiceBrokerRelistBehavior = "Duration"

	// ServiceBrokerRelistBehaviorManual indicates that the broker is only
	// relisted when the spec of the broker changes.
	ServiceBrokerRelistBehaviorManual ServiceBrokerRelistBehavior = "Manual"
)

// ClusterServiceBrokerAuthInfo is a union type that contains information on
// one of the authentication methods the service catalog and brokers may
// support, according to the OpenServiceBroker API specification
// (https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md).
type ClusterServiceBrokerAuthInfo struct {
	// ClusterBasicAuthConfigprovides configuration for basic authentication.
	Basic *ClusterBasicAuthConfig `json:"basic,omitempty"`
	// ClusterBearerTokenAuthConfig provides configuration to send an opaque value as a bearer token.
	// The value is referenced from the 'token' field of the given secret.  This value should only
	// contain the token value and not the `Bearer` scheme.
	Bearer *ClusterBearerTokenAuthConfig `json:"bearer,omitempty"`
}

// ClusterBasicAuthConfig provides config for the basic authentication of
// cluster scoped brokers.
type ClusterBasicAuthConfig struct {
	// SecretRef is a reference to a Secret containing information the
	// catalog should use to authenticate to this ServiceBroker.
	//
	// Required at least one of the fields:
	// - Secret.Data["username"] - username used for authentication
	// - Secret.Data["password"] - password or token needed for authentication
	SecretRef *ObjectReference `json:"secretRef,omitempty"`
}

// ClusterBearerTokenAuthConfig provides config for the bearer token
// authentication of cluster scoped brokers.
type ClusterBearerTokenAuthConfig struct {
	// SecretRef is a reference to a Secret containing information the
	// catalog should use to authenticate to this ServiceBroker.
	//
	// Required field:
	// - Secret.Data["token"] - bearer token for authentication
	SecretRef *ObjectReference `json:"secretRef,omitempty"`
}

// ServiceBrokerAuthInfo is a union type that contains information on
// one of the authentication methods the service catalog and brokers may
// support, according to the OpenServiceBroker API specification
// (https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md).
type ServiceBrokerAuthInfo struct {
	// BasicAuthConfig provides configuration for basic authentication.
	Basic *BasicAuthConfig `json:"basic,omitempty"`
	// BearerTokenAuthConfig provides configuration to send an opaque value as a bearer token.
	// The value is referenced from the 'token' field of the given secret.  This value should only
	// contain the token value and not the `Bearer` scheme.
	Bearer *BearerTokenAuthConfig `json:"bearer,omitempty"`
}

// BasicAuthConfig provides config for the basic authentication of
// cluster scoped brokers.
type BasicAuthConfig struct {
	// SecretRef is a reference to a Secret containing information the
	// catalog should use to authenticate to this ServiceBroker.
	//
	// Required at least one of the fields:
	// - Secret.Data["username"] - username used for authentication
	// - Secret.Data["password"] - password or token needed for authentication
	SecretRef *LocalObjectReference `json:"secretRef,omitempty"`
}

// BearerTokenAuthConfig provides config for the bearer token
// authentication of cluster scoped brokers.
type BearerTokenAuthConfig struct {
	// SecretRef is a reference to a Secret containing information the
	// catalog should use to authenticate to this ServiceBroker.
	//
	// Required field:
	// - Secret.Data["token"] - bearer token for authentication
	SecretRef *LocalObjectReference `json:"secretRef,omitempty"`
}

const (
	// BasicAuthUsernameKey is the key of the username for SecretTypeBasicAuth secrets
	BasicAuthUsernameKey = "username"
	// BasicAuthPasswordKey is the key of the password or token for SecretTypeBasicAuth secrets
	BasicAuthPasswordKey = "password"

	// BearerTokenKey is the key of the bearer token for SecretTypeBearerTokenAuth secrets
	BearerTokenKey = "token"
)

// CommonServiceBrokerStatus represents the current status of a Broker.
type CommonServiceBrokerStatus struct {
	Conditions []ServiceBrokerCondition `json:"conditions"`

	// ReconciledGeneration is the 'Generation' of the ClusterServiceBrokerSpec that
	// was last processed by the controller. The reconciled generation is updated
	// even if the controller failed to process the spec.
	ReconciledGeneration int64 `json:"reconciledGeneration"`

	// OperationStartTime is the time at which the current operation began.
	OperationStartTime *metav1.Time `json:"operationStartTime,omitempty"`

	// LastCatalogRetrievalTime is the time the Catalog was last fetched from
	// the Service Broker
	LastCatalogRetrievalTime *metav1.Time `json:"lastCatalogRetrievalTime,omitempty"`

	// LastConditionState aggregates state from the Conditions array
	// It is used for printing in a kubectl output via additionalPrinterColumns
	LastConditionState string `json:"lastConditionState"`
}

// ClusterServiceBrokerStatus represents the current status of a
// ClusterServiceBroker.
type ClusterServiceBrokerStatus struct {
	CommonServiceBrokerStatus `json:",inline"`
}

// ServiceBrokerStatus the current status of a ServiceBroker.
type ServiceBrokerStatus struct {
	CommonServiceBrokerStatus `json:",inline"`
}

// ServiceBrokerCondition contains condition information for a Broker.
type ServiceBrokerCondition struct {
	// Type of the condition, currently ('Ready').
	Type ServiceBrokerConditionType `json:"type"`

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

// ServiceBrokerConditionType represents a broker condition value.
type ServiceBrokerConditionType string

const (
	// ServiceBrokerConditionReady represents the fact that a given broker condition
	// is in ready state.
	ServiceBrokerConditionReady ServiceBrokerConditionType = "Ready"

	// ServiceBrokerConditionFailed represents information about a final failure
	// that should not be retried.
	ServiceBrokerConditionFailed ServiceBrokerConditionType = "Failed"
)

// ConditionStatus represents a condition's status.
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServiceClassList is a list of ClusterServiceClasses.
type ClusterServiceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterServiceClass `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServiceClass represents an offering in the service catalog.
// +k8s:openapi-gen=x-kubernetes-print-columns:custom-columns=NAME:.metadata.name,EXTERNAL NAME:.spec.externalName,BROKER:.spec.clusterServiceBrokerName,BINDABLE:.spec.bindable,PLAN UPDATABLE:.spec.planUpdatable
type ClusterServiceClass struct {
	metav1.TypeMeta `json:",inline"`

	// Non-namespaced.  The name of this resource in etcd is in ObjectMeta.Name.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of the cluster service class.
	// +optional
	Spec ClusterServiceClassSpec `json:"spec,omitempty"`

	// Status represents the current status of the cluster service class.
	// +optional
	Status ClusterServiceClassStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceClassList is a list of ServiceClasses.
type ServiceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceClass `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceClass represents a namespaced offering in the service catalog.
type ServiceClass struct {
	metav1.TypeMeta `json:",inline"`

	// The name of this resource in etcd is in ObjectMeta.Name.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of the service class.
	// +optional
	Spec ServiceClassSpec `json:"spec,omitempty"`

	// Status represents the current status of a service class.
	// +optional
	Status ServiceClassStatus `json:"status,omitempty"`
}

// ServiceClassStatus represents status information about a ServiceClass.
type ServiceClassStatus struct {
	CommonServiceClassStatus `json:",inline"`
}

// ClusterServiceClassStatus represents status information about a
// ClusterServiceClass.
type ClusterServiceClassStatus struct {
	CommonServiceClassStatus `json:",inline"`
}

// CommonServiceClassStatus represents common status information between
// cluster scoped and namespace scoped ServiceClasses.
type CommonServiceClassStatus struct {
	// RemovedFromBrokerCatalog indicates that the broker removed the service from its
	// catalog.
	RemovedFromBrokerCatalog bool `json:"removedFromBrokerCatalog"`
}

// CommonServiceClassSpec represents details about a ServiceClass
type CommonServiceClassSpec struct {
	// ExternalName is the name of this object that the Service Broker
	// exposed this Service Class as. Mutable.
	ExternalName string `json:"externalName"`

	// ExternalID is the identity of this object for use with the OSB API.
	//
	// Immutable.
	ExternalID string `json:"externalID"`

	// Description is a short description of this ServiceClass.
	Description string `json:"description"`

	// Bindable indicates whether a user can create bindings to an
	// ServiceInstance provisioned from this service. ServicePlan
	// has an optional field called Bindable which overrides the value of
	// this field.
	Bindable bool `json:"bindable"`

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// BindingRetrievable indicates whether fetching a binding via a GET on
	// its endpoint is supported for all plans.
	BindingRetrievable bool `json:"bindingRetrievable"`

	// PlanUpdatable indicates whether instances provisioned from this
	// ServiceClass may change ServicePlans after being
	// provisioned.
	PlanUpdatable bool `json:"planUpdatable"`

	// ExternalMetadata is a blob of information about the
	// ServiceClass, meant to be user-facing content and display
	// instructions. This field may contain platform-specific conventional
	// values.
	ExternalMetadata *runtime.RawExtension `json:"externalMetadata,omitempty"`

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// Tags is a list of strings that represent different classification
	// attributes of the ServiceClass.  These are used in Cloud
	// Foundry in a way similar to Kubernetes labels, but they currently
	// have no special meaning in Kubernetes.
	Tags []string `json:"tags,omitempty"`

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// Requires exposes a list of Cloud Foundry-specific 'permissions'
	// that must be granted to an instance of this service within Cloud
	// Foundry.  These 'permissions' have no meaning within Kubernetes and an
	// ServiceInstance provisioned from this ServiceClass will not
	// work correctly.
	Requires []string `json:"requires,omitempty"`

	// DefaultProvisionParameters are default parameters passed to the broker
	// when an instance of this class is provisioned. Any parameters defined on
	// the plan and instance are merged with these defaults, with
	// plan and then instance-defined parameters taking precedence over the class
	// defaults.
	DefaultProvisionParameters *runtime.RawExtension `json:"defaultProvisionParameters,omitempty"`
}

// ClusterServiceClassSpec represents the details about a ClusterServiceClass
type ClusterServiceClassSpec struct {
	CommonServiceClassSpec `json:",inline"`

	// ClusterServiceBrokerName is the reference to the Broker that provides this
	// ClusterServiceClass.
	//
	// Immutable.
	ClusterServiceBrokerName string `json:"clusterServiceBrokerName"`
}

// ServiceClassSpec represents the details about a ServiceClass
type ServiceClassSpec struct {
	CommonServiceClassSpec `json:",inline"`

	// ServiceBrokerName is the reference to the Broker that provides this
	// ServiceClass.
	//
	// Immutable.
	ServiceBrokerName string `json:"serviceBrokerName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServicePlanList is a list of ClusterServicePlans.
type ClusterServicePlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterServicePlan `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterServicePlan represents a tier of a ServiceClass.
// +k8s:openapi-gen=x-kubernetes-print-columns:custom-columns=NAME:.metadata.name,EXTERNAL NAME:.spec.externalName,BROKER:.spec.clusterServiceBrokerName,CLASS:.spec.clusterServiceClassRef.name
type ClusterServicePlan struct {
	metav1.TypeMeta `json:",inline"`

	// Non-namespaced.  The name of this resource in etcd is in ObjectMeta.Name.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of the service plan.
	// +optional
	Spec ClusterServicePlanSpec `json:"spec,omitempty"`

	// Status represents the current status of the service plan.
	// +optional
	Status ClusterServicePlanStatus `json:"status,omitempty"`
}

// CommonServicePlanSpec represents details that are shared by both
// a ClusterServicePlan and a namespaced ServicePlan
type CommonServicePlanSpec struct {
	// ExternalName is the name of this object that the Service Broker
	// exposed this Service Plan as. Mutable.
	ExternalName string `json:"externalName"`

	// ExternalID is the identity of this object for use with the OSB API.
	//
	// Immutable.
	ExternalID string `json:"externalID"`

	// Description is a short description of this ServicePlan.
	Description string `json:"description"`

	// Bindable indicates whether a user can create bindings to an
	// ServiceInstance using this ServicePlan.  If set, overrides
	// the value of the corresponding ServiceClassSpec Bindable field.
	Bindable *bool `json:"bindable,omitempty"`

	// Free indicates whether this plan is available at no cost.
	Free bool `json:"free"`

	// ExternalMetadata is a blob of information about the plan, meant to be
	// user-facing content and display instructions.  This field may contain
	// platform-specific conventional values.
	ExternalMetadata *runtime.RawExtension `json:"externalMetadata,omitempty"`

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// InstanceCreateParameterSchema is the schema for the parameters
	// that may be supplied when provisioning a new ServiceInstance on this plan.
	InstanceCreateParameterSchema *runtime.RawExtension `json:"instanceCreateParameterSchema,omitempty"`

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// InstanceUpdateParameterSchema is the schema for the parameters
	// that may be updated once an ServiceInstance has been provisioned on
	// this plan. This field only has meaning if the corresponding ServiceClassSpec is
	// PlanUpdatable.
	InstanceUpdateParameterSchema *runtime.RawExtension `json:"instanceUpdateParameterSchema,omitempty"`

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// ServiceBindingCreateParameterSchema is the schema for the parameters that
	// may be supplied binding to a ServiceInstance on this plan.
	ServiceBindingCreateParameterSchema *runtime.RawExtension `json:"serviceBindingCreateParameterSchema,omitempty"`

	// DefaultProvisionParameters are default parameters passed to the broker
	// when an instance of this plan is provisioned. Any parameters defined on
	// the instance are merged with these defaults, with instance-defined
	// parameters taking precedence over defaults.
	DefaultProvisionParameters *runtime.RawExtension `json:"defaultProvisionParameters,omitempty"`
}

// ClusterServicePlanSpec represents details about a ClusterServicePlan.
type ClusterServicePlanSpec struct {
	// CommonServicePlanSpec contains the common details of this ClusterServicePlan
	CommonServicePlanSpec `json:",inline"`

	// ClusterServiceBrokerName is the name of the ClusterServiceBroker
	// that offers this ClusterServicePlan.
	ClusterServiceBrokerName string `json:"clusterServiceBrokerName"`

	// ClusterServiceClassRef is a reference to the service class that
	// owns this plan.
	ClusterServiceClassRef ClusterObjectReference `json:"clusterServiceClassRef"`
}

// ClusterServicePlanStatus represents status information about a
// ClusterServicePlan.
type ClusterServicePlanStatus struct {
	CommonServicePlanStatus `json:",inline"`
}

// CommonServicePlanStatus represents status information about a
// ClusterServicePlan or a ServicePlan.
type CommonServicePlanStatus struct {
	// RemovedFromBrokerCatalog indicates that the broker removed the plan
	// from its catalog.
	RemovedFromBrokerCatalog bool `json:"removedFromBrokerCatalog"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServicePlanList is a list of rServicePlans.
type ServicePlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServicePlan `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServicePlan represents a tier of a ServiceClass.
// +k8s:openapi-gen=x-kubernetes-print-columns:custom-columns=NAME:.metadata.name,EXTERNAL NAME:.spec.externalName,BROKER:.spec.serviceBrokerName,CLASS:.spec.serviceClassRef.name
type ServicePlan struct {
	metav1.TypeMeta `json:",inline"`

	// Non-namespaced.  The name of this resource in etcd is in ObjectMeta.Name.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of the service plan.
	// +optional
	Spec ServicePlanSpec `json:"spec,omitempty"`

	// Status represents the current status of the service plan.
	// +optional
	Status ServicePlanStatus `json:"status,omitempty"`
}

// ServicePlanSpec represents details about a ServicePlan.
type ServicePlanSpec struct {
	// CommonServicePlanSpec contains the common details of this ServicePlan
	CommonServicePlanSpec `json:",inline"`

	// ServiceBrokerName is the name of the ServiceBroker
	// that offers this ServicePlan.
	ServiceBrokerName string `json:"serviceBrokerName"`

	// ServiceClassRef is a reference to the service class that
	// owns this plan.
	ServiceClassRef LocalObjectReference `json:"serviceClassRef"`
}

// ServicePlanStatus represents status information about a
// ServicePlan.
type ServicePlanStatus struct {
	CommonServicePlanStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceInstanceList is a list of instances.
type ServiceInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceInstance `json:"items"`
}

// UserInfo holds information about the user that last changed a resource's spec.
type UserInfo struct {
	Username string                `json:"username"`
	UID      string                `json:"uid"`
	Groups   []string              `json:"groups,omitempty"`
	Extra    map[string]ExtraValue `json:"extra,omitempty"`
}

// ExtraValue contains additional information about a user that may be
// provided by the authenticator.
type ExtraValue []string

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceInstance represents a provisioned instance of a ServiceClass.
// Currently, the spec field cannot be changed once a ServiceInstance is
// created.  Spec changes submitted by users will be ignored.
//
// In the future, this will be allowed and will represent the intention that
// the ServiceInstance should have the plan and/or parameters updated at the
// ClusterServiceBroker.
// +k8s:openapi-gen=x-kubernetes-print-columns:custom-columns=NAME:.metadata.name,CLASS:.spec.clusterServiceClassExternalName,PLAN:.spec.clusterServicePlanExternalName
type ServiceInstance struct {
	metav1.TypeMeta `json:",inline"`

	// The name of this resource in etcd is in ObjectMeta.Name.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of the service instance.
	// +optional
	Spec ServiceInstanceSpec `json:"spec,omitempty"`

	// Status represents the current status of a service instance.
	// +optional
	Status ServiceInstanceStatus `json:"status,omitempty"`
}

// PlanReference defines the user specification for the desired
// (Cluster)ServicePlan and (Cluster)ServiceClass. Because there are
// multiple ways to specify the desired Class/Plan, this structure specifies the
// allowed ways to specify the intent. Note: a user may specify either cluster
// scoped OR namespace scoped identifiers, but NOT both, as they are mutually
// exclusive.
//
// Currently supported ways:
//  - ClusterServiceClassExternalName and ClusterServicePlanExternalName
//  - ClusterServiceClassExternalID and ClusterServicePlanExternalID
//  - ClusterServiceClassName and ClusterServicePlanName
//  - ServiceClassExternalName and ServicePlanExternalName
//  - ServiceClassExternalID and ServicePlanExternalID
//  - ServiceClassName and ServicePlanName
//
// For any of these ways, if a ClusterServiceClass only has one plan
// then the corresponding service plan field is optional.
type PlanReference struct {
	// ClusterServiceClassExternalName is the human-readable name of the
	// service as reported by the ClusterServiceBroker. Note that if the
	// ClusterServiceBroker changes the name of the ClusterServiceClass,
	// it will not be reflected here, and to see the current name of the
	// ClusterServiceClass, you should follow the ClusterServiceClassRef below.
	//
	// Immutable.
	ClusterServiceClassExternalName string `json:"clusterServiceClassExternalName,omitempty"`
	// ClusterServicePlanExternalName is the human-readable name of the plan
	// as reported by the ClusterServiceBroker. Note that if the
	// ClusterServiceBroker changes the name of the ClusterServicePlan, it will
	// not be reflected here, and to see the current name of the
	// ClusterServicePlan, you should follow the ClusterServicePlanRef below.
	ClusterServicePlanExternalName string `json:"clusterServicePlanExternalName,omitempty"`

	// ClusterServiceClassExternalID is the ClusterServiceBroker's external id
	// for the class.
	//
	// Immutable.
	ClusterServiceClassExternalID string `json:"clusterServiceClassExternalID,omitempty"`

	// ClusterServicePlanExternalID is the ClusterServiceBroker's external id for
	// the plan.
	ClusterServicePlanExternalID string `json:"clusterServicePlanExternalID,omitempty"`

	// ClusterServiceClassName is the kubernetes name of the ClusterServiceClass.
	//
	// Immutable.
	ClusterServiceClassName string `json:"clusterServiceClassName,omitempty"`
	// ClusterServicePlanName is kubernetes name of the ClusterServicePlan.
	ClusterServicePlanName string `json:"clusterServicePlanName,omitempty"`

	// ServiceClassExternalName is the human-readable name of the
	// service as reported by the ServiceBroker. Note that if the ServiceBroker
	// changes the name of the ServiceClass, it will not be reflected here,
	// and to see the current name of the ServiceClass, you should
	// follow the ServiceClassRef below.
	//
	// Immutable.
	ServiceClassExternalName string `json:"serviceClassExternalName,omitempty"`
	// ServicePlanExternalName is the human-readable name of the plan
	// as reported by the ServiceBroker. Note that if the ServiceBroker changes
	// the name of the ServicePlan, it will not be reflected here, and to see
	// the current name of the ServicePlan, you should follow the
	// ServicePlanRef below.
	ServicePlanExternalName string `json:"servicePlanExternalName,omitempty"`

	// ServiceClassExternalID is the ServiceBroker's external id for the class.
	//
	// Immutable.
	ServiceClassExternalID string `json:"serviceClassExternalID,omitempty"`

	// ServicePlanExternalID is the ServiceBroker's external id for the plan.
	ServicePlanExternalID string `json:"servicePlanExternalID,omitempty"`

	// ServiceClassName is the kubernetes name of the ServiceClass.
	//
	// Immutable.
	ServiceClassName string `json:"serviceClassName,omitempty"`
	// ServicePlanName is kubernetes name of the ServicePlan.
	ServicePlanName string `json:"servicePlanName,omitempty"`
}

// ServiceInstanceSpec represents the desired state of an Instance.
type ServiceInstanceSpec struct {
	// Specification of what ServiceClass/ServicePlan is being provisioned.
	PlanReference `json:",inline"`

	// ClusterServiceClassRef is a reference to the ClusterServiceClass
	// that the user selected. This is set by the controller based on the
	// cluster-scoped values specified in the PlanReference.
	ClusterServiceClassRef *ClusterObjectReference `json:"clusterServiceClassRef,omitempty"`
	// ClusterServicePlanRef is a reference to the ClusterServicePlan
	// that the user selected. This is set by the controller based on the
	// cluster-scoped values specified in the PlanReference.
	ClusterServicePlanRef *ClusterObjectReference `json:"clusterServicePlanRef,omitempty"`

	// ServiceClassRef is a reference to the ServiceClass that the user selected.
	// This is set by the controller based on the namespace-scoped values
	// specified in the PlanReference.
	ServiceClassRef *LocalObjectReference `json:"serviceClassRef,omitempty"`
	// ServicePlanRef is a reference to the ServicePlan that the user selected.
	// This is set by the controller based on the namespace-scoped values
	// specified in the PlanReference.
	ServicePlanRef *LocalObjectReference `json:"servicePlanRef,omitempty"`

	// Parameters is a set of the parameters to be passed to the underlying
	// broker. The inline YAML/JSON payload to be translated into equivalent
	// JSON object. If a top-level parameter name exists in multiples sources
	// among `Parameters` and `ParametersFrom` fields, it is considered to be
	// a user error in the specification.
	//
	// The Parameters field is NOT secret or secured in any way and should
	// NEVER be used to hold sensitive information. To set parameters that
	// contain secret information, you should ALWAYS store that information
	// in a Secret and use the ParametersFrom field.
	//
	// +optional
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`

	// List of sources to populate parameters.
	// If a top-level parameter name exists in multiples sources among
	// `Parameters` and `ParametersFrom` fields, it is
	// considered to be a user error in the specification
	// +optional
	ParametersFrom []ParametersFromSource `json:"parametersFrom,omitempty"`

	// ExternalID is the identity of this object for use with the OSB SB API.
	//
	// Immutable.
	// +optional
	ExternalID string `json:"externalID"`

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// UserInfo contains information about the user that last modified this
	// instance. This field is set by the API server and not settable by the
	// end-user. User-provided values for this field are not saved.
	// +optional
	UserInfo *UserInfo `json:"userInfo,omitempty"`

	// UpdateRequests is a strictly increasing, non-negative integer counter that
	// can be manually incremented by a user to manually trigger an update. This
	// allows for parameters to be updated with any out-of-band changes that have
	// been made to the secrets from which the parameters are sourced.
	// +optional
	UpdateRequests int64 `json:"updateRequests"`
}

// ServiceInstanceStatus represents the current status of an Instance.
type ServiceInstanceStatus struct {
	// Conditions is an array of ServiceInstanceConditions capturing aspects of an
	// ServiceInstance's status.
	Conditions []ServiceInstanceCondition `json:"conditions"`

	// AsyncOpInProgress is set to true if there is an ongoing async operation
	// against this Service Instance in progress.
	AsyncOpInProgress bool `json:"asyncOpInProgress"`

	// OrphanMitigationInProgress is set to true if there is an ongoing orphan
	// mitigation operation against this ServiceInstance in progress.
	OrphanMitigationInProgress bool `json:"orphanMitigationInProgress"`

	// LastOperation is the string that the broker may have returned when
	// an async operation started, it should be sent back to the broker
	// on poll requests as a query param.
	LastOperation *string `json:"lastOperation,omitempty"`

	// DashboardURL is the URL of a web-based management user interface for
	// the service instance.
	DashboardURL *string `json:"dashboardURL,omitempty"`

	// CurrentOperation is the operation the Controller is currently performing
	// on the ServiceInstance.
	CurrentOperation ServiceInstanceOperation `json:"currentOperation,omitempty"`

	// ReconciledGeneration is the 'Generation' of the serviceInstanceSpec that
	// was last processed by the controller. The reconciled generation is updated
	// even if the controller failed to process the spec.
	// Deprecated: use ObservedGeneration with conditions set to true to find
	// whether generation was reconciled.
	ReconciledGeneration int64 `json:"reconciledGeneration"`

	// ObservedGeneration is the 'Generation' of the serviceInstanceSpec that
	// was last processed by the controller. The observed generation is updated
	// whenever the status is updated regardless of operation result.
	ObservedGeneration int64 `json:"observedGeneration"`

	// OperationStartTime is the time at which the current operation began.
	OperationStartTime *metav1.Time `json:"operationStartTime,omitempty"`

	// InProgressProperties is the properties state of the ServiceInstance when
	// a Provision, Update or Deprovision is in progress.
	InProgressProperties *ServiceInstancePropertiesState `json:"inProgressProperties,omitempty"`

	// ExternalProperties is the properties state of the ServiceInstance which the
	// broker knows about.
	ExternalProperties *ServiceInstancePropertiesState `json:"externalProperties,omitempty"`

	// ProvisionStatus describes whether the instance is in the provisioned state.
	ProvisionStatus ServiceInstanceProvisionStatus `json:"provisionStatus"`

	// DeprovisionStatus describes what has been done to deprovision the
	// ServiceInstance.
	DeprovisionStatus ServiceInstanceDeprovisionStatus `json:"deprovisionStatus"`

	// DefaultProvisionParameters are the default parameters applied to this
	// instance.
	DefaultProvisionParameters *runtime.RawExtension `json:"defaultProvisionParameters,omitempty"`

	// LastConditionState aggregates state from the Conditions array
	// It is used for printing in a kubectl output via additionalPrinterColumns
	LastConditionState string `json:"lastConditionState"`

	// UserSpecifiedPlanName aggregates cluster or namespace PlanName
	// It is used for printing in a kubectl output via additionalPrinterColumns
	UserSpecifiedPlanName string `json:"userSpecifiedPlanName"`

	// UserSpecifiedClassName aggregates cluster or namespace ClassName
	// It is used for printing in a kubectl output via additionalPrinterColumns
	UserSpecifiedClassName string `json:"userSpecifiedClassName"`
}

// ServiceInstanceCondition contains condition information about an Instance.
type ServiceInstanceCondition struct {
	// Type of the condition, currently ('Ready').
	Type ServiceInstanceConditionType `json:"type"`

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

// ServiceInstanceConditionType represents a instance condition value.
type ServiceInstanceConditionType string

const (
	// ServiceInstanceConditionReady represents that a given InstanceCondition is in
	// ready state.
	ServiceInstanceConditionReady ServiceInstanceConditionType = "Ready"

	// ServiceInstanceConditionFailed represents information about a final failure
	// that should not be retried.
	ServiceInstanceConditionFailed ServiceInstanceConditionType = "Failed"

	// ServiceInstanceConditionOrphanMitigation represents information about an
	// orphan mitigation that is required after failed provisioning.
	ServiceInstanceConditionOrphanMitigation ServiceInstanceConditionType = "OrphanMitigation"
)

// ServiceInstanceOperation represents a type of operation the controller can
// be performing for a service instance in the OSB API.
type ServiceInstanceOperation string

const (
	// ServiceInstanceOperationProvision indicates that the ServiceInstance is
	// being Provisioned.
	ServiceInstanceOperationProvision ServiceInstanceOperation = "Provision"
	// ServiceInstanceOperationUpdate indicates that the ServiceInstance is
	// being Updated.
	ServiceInstanceOperationUpdate ServiceInstanceOperation = "Update"
	// ServiceInstanceOperationDeprovision indicates that the ServiceInstance is
	// being Deprovisioned.
	ServiceInstanceOperationDeprovision ServiceInstanceOperation = "Deprovision"
)

// ServiceInstancePropertiesState is the state of a ServiceInstance that
// the ClusterServiceBroker knows about.
type ServiceInstancePropertiesState struct {
	// ClusterServicePlanExternalName is the name of the plan that the
	// broker knows this ServiceInstance to be on. This is the human
	// readable plan name from the OSB API.
	ClusterServicePlanExternalName string `json:"clusterServicePlanExternalName"`

	// ClusterServicePlanExternalID is the external ID of the plan that the
	// broker knows this ServiceInstance to be on.
	ClusterServicePlanExternalID string `json:"clusterServicePlanExternalID"`

	// ServicePlanExternalName is the name of the plan that the broker knows this
	// ServiceInstance to be on. This is the human readable plan name from the
	// OSB API.
	ServicePlanExternalName string `json:"servicePlanExternalName,omitempty"`

	// ServicePlanExternalID is the external ID of the plan that the
	// broker knows this ServiceInstance to be on.
	ServicePlanExternalID string `json:"servicePlanExternalID,omitempty"`

	// Parameters is a blob of the parameters and their values that the broker
	// knows about for this ServiceInstance.  If a parameter was sourced from
	// a secret, its value will be "<redacted>" in this blob.
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`

	// ParameterChecksum is the checksum of the parameters that were sent.
	ParameterChecksum string `json:"parameterChecksum,omitempty"`

	// UserInfo is information about the user that made the request.
	UserInfo *UserInfo `json:"userInfo,omitempty"`
}

// ServiceInstanceDeprovisionStatus is the status of deprovisioning a
// ServiceInstance
type ServiceInstanceDeprovisionStatus string

const (
	// ServiceInstanceDeprovisionStatusNotRequired indicates that a provision
	// request has not been sent for the ServiceInstance, so no deprovision
	// request needs to be made.
	ServiceInstanceDeprovisionStatusNotRequired ServiceInstanceDeprovisionStatus = "NotRequired"
	// ServiceInstanceDeprovisionStatusRequired indicates that a provision
	// request has been sent for the ServiceInstance. A deprovision request
	// must be made before deleting the ServiceInstance.
	ServiceInstanceDeprovisionStatusRequired ServiceInstanceDeprovisionStatus = "Required"
	// ServiceInstanceDeprovisionStatusSucceeded indicates that a deprovision
	// request has been sent for the ServiceInstance, and the request was
	// successful.
	ServiceInstanceDeprovisionStatusSucceeded ServiceInstanceDeprovisionStatus = "Succeeded"
	// ServiceInstanceDeprovisionStatusFailed indicates that deprovision
	// requests have been sent for the ServiceInstance but they failed. The
	// controller has given up on sending more deprovision requests.
	ServiceInstanceDeprovisionStatusFailed ServiceInstanceDeprovisionStatus = "Failed"
)

// ServiceInstanceProvisionStatus is the status of provisioning a
// ServiceInstance
type ServiceInstanceProvisionStatus string

const (
	// ServiceInstanceProvisionStatusProvisioned indicates that the instance
	// was provisioned.
	ServiceInstanceProvisionStatusProvisioned ServiceInstanceProvisionStatus = "Provisioned"
	// ServiceInstanceProvisionStatusNotProvisioned indicates that the instance
	// was not ever provisioned or was deprovisioned.
	ServiceInstanceProvisionStatusNotProvisioned ServiceInstanceProvisionStatus = "NotProvisioned"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBindingList is a list of ServiceBindings.
type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceBinding `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBinding represents a "used by" relationship between an application and an
// ServiceInstance.
// +k8s:openapi-gen=x-kubernetes-print-columns:custom-columns=NAME:.metadata.name,INSTANCE:.spec.instanceRef.name,SECRET:.spec.secretName
type ServiceBinding struct {
	metav1.TypeMeta `json:",inline"`

	// The name of this resource in etcd is in ObjectMeta.Name.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec represents the desired state of a ServiceBinding.
	// +optional
	Spec ServiceBindingSpec `json:"spec,omitempty"`

	// Status represents the current status of a ServiceBinding.
	// +optional
	Status ServiceBindingStatus `json:"status,omitempty"`
}

// ServiceBindingSpec represents the desired state of a
// ServiceBinding.
//
// The spec field cannot be changed after a ServiceBinding is
// created.  Changes submitted to the spec field will be ignored.
type ServiceBindingSpec struct {
	// InstanceRef is the reference to the Instance this ServiceBinding is to.
	//
	// Immutable.
	InstanceRef LocalObjectReference `json:"instanceRef"`

	// Parameters is a set of the parameters to be passed to the underlying
	// broker. The inline YAML/JSON payload to be translated into equivalent
	// JSON object. If a top-level parameter name exists in multiples sources
	// among `Parameters` and `ParametersFrom` fields, it is considered to be
	// a user error in the specification.
	//
	// The Parameters field is NOT secret or secured in any way and should
	// NEVER be used to hold sensitive information. To set parameters that
	// contain secret information, you should ALWAYS store that information
	// in a Secret and use the ParametersFrom field.
	//
	// +optional
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`

	// List of sources to populate parameters.
	// If a top-level parameter name exists in multiples sources among
	// `Parameters` and `ParametersFrom` fields, it is
	// considered to be a user error in the specification.
	// +optional
	ParametersFrom []ParametersFromSource `json:"parametersFrom,omitempty"`

	// SecretName is the name of the secret to create in the ServiceBinding's
	// namespace that will hold the credentials associated with the ServiceBinding.
	SecretName string `json:"secretName,omitempty"`

	// SecretKey is used as the key inside the secret to store the credentials
	// returned by the broker encoded as json. If not specified the credentials
	// returned by the broker will be used directly as the secrets data. This
	// can cause problems when using complex data structures.
	// +optional
	SecretKey *string `json:"secretKey,omitempty"`

	// List of transformations that should be applied to the credentials
	// associated with the ServiceBinding before they are inserted into the Secret.
	SecretTransforms []SecretTransform `json:"secretTransforms,omitempty"`

	// ExternalID is the identity of this object for use with the OSB API.
	//
	// Immutable.
	// +optional
	ExternalID string `json:"externalID"`

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// UserInfo contains information about the user that last modified this
	// ServiceBinding. This field is set by the API server and not
	// settable by the end-user. User-provided values for this field are not saved.
	// +optional
	UserInfo *UserInfo `json:"userInfo,omitempty"`
}

// ServiceBindingStatus represents the current status of a ServiceBinding.
type ServiceBindingStatus struct {
	Conditions []ServiceBindingCondition `json:"conditions"`

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// AsyncOpInProgress is set to true if there is an ongoing async operation
	// against this ServiceBinding in progress.
	AsyncOpInProgress bool `json:"asyncOpInProgress"`

	// Currently, this field is ALPHA: it may change or disappear at any time
	// and its data will not be migrated.
	//
	// LastOperation is the string that the broker may have returned when
	// an async operation started, it should be sent back to the broker
	// on poll requests as a query param.
	LastOperation *string `json:"lastOperation,omitempty"`

	// CurrentOperation is the operation the Controller is currently performing
	// on the ServiceBinding.
	CurrentOperation ServiceBindingOperation `json:"currentOperation,omitempty"`

	// ReconciledGeneration is the 'Generation' of the
	// ServiceBindingSpec that was last processed by the controller.
	// The reconciled generation is updated even if the controller failed to
	// process the spec.
	ReconciledGeneration int64 `json:"reconciledGeneration"`

	// OperationStartTime is the time at which the current operation began.
	OperationStartTime *metav1.Time `json:"operationStartTime,omitempty"`

	// InProgressProperties is the properties state of the
	// ServiceBinding when a Bind is in progress. If the current
	// operation is an Unbind, this will be nil.
	InProgressProperties *ServiceBindingPropertiesState `json:"inProgressProperties,omitempty"`

	// ExternalProperties is the properties state of the
	// ServiceBinding which the broker knows about.
	ExternalProperties *ServiceBindingPropertiesState `json:"externalProperties,omitempty"`

	// OrphanMitigationInProgress is a flag that represents whether orphan
	// mitigation is in progress.
	OrphanMitigationInProgress bool `json:"orphanMitigationInProgress"`

	// UnbindStatus describes what has been done to unbind the ServiceBinding.
	UnbindStatus ServiceBindingUnbindStatus `json:"unbindStatus"`

	// LastConditionState aggregates state from the Conditions array
	// It is used for printing in a kubectl output via additionalPrinterColumns
	LastConditionState string `json:"lastConditionState"`
}

// ServiceBindingCondition condition information for a ServiceBinding.
type ServiceBindingCondition struct {
	// Type of the condition, currently ('Ready').
	Type ServiceBindingConditionType `json:"type"`

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

// ServiceBindingConditionType represents a ServiceBindingCondition value.
type ServiceBindingConditionType string

const (
	// ServiceBindingConditionReady represents a binding condition is in ready state.
	ServiceBindingConditionReady ServiceBindingConditionType = "Ready"

	// ServiceBindingConditionFailed represents a ServiceBindingCondition that has failed
	// completely and should not be retried.
	ServiceBindingConditionFailed ServiceBindingConditionType = "Failed"
)

// ServiceBindingOperation represents a type of operation
// the controller can be performing for a binding in the OSB API.
type ServiceBindingOperation string

const (
	// ServiceBindingOperationBind indicates that the
	// ServiceBinding is being bound.
	ServiceBindingOperationBind ServiceBindingOperation = "Bind"
	// ServiceBindingOperationUnbind indicates that the
	// ServiceBinding is being unbound.
	ServiceBindingOperationUnbind ServiceBindingOperation = "Unbind"
)

// ServiceBindingUnbindStatus is the status of unbinding a Binding
type ServiceBindingUnbindStatus string

const (
	// ServiceBindingUnbindStatusNotRequired indicates that a binding request
	// has not been sent for the ServiceBinding, so no unbinding request
	// needs to be made.
	ServiceBindingUnbindStatusNotRequired ServiceBindingUnbindStatus = "NotRequired"
	// ServiceBindingUnbindStatusRequired indicates that a binding request has
	// been sent for the ServiceBinding. An unbind request must be made before
	// deleting the ServiceBinding.
	ServiceBindingUnbindStatusRequired ServiceBindingUnbindStatus = "Required"
	// ServiceBindingUnbindStatusSucceeded indicates that a unbind request has
	// been sent for the ServiceBinding, and the request was successful.
	ServiceBindingUnbindStatusSucceeded ServiceBindingUnbindStatus = "Succeeded"
	// ServiceBindingUnbindStatusFailed indicates that unbind requests have
	// been sent for the ServiceBinding but they failed. The controller has
	// given up on sending more unbind requests.
	ServiceBindingUnbindStatusFailed ServiceBindingUnbindStatus = "Failed"
)

// These are external finalizer values to service catalog, must be qualified name.
const (
	FinalizerServiceCatalog string = "kubernetes-incubator/service-catalog"
)

// ServiceBindingPropertiesState is the state of a
// ServiceBinding that the ClusterServiceBroker knows about.
type ServiceBindingPropertiesState struct {
	// Parameters is a blob of the parameters and their values that the broker
	// knows about for this ServiceBinding.  If a parameter was
	// sourced from a secret, its value will be "<redacted>" in this blob.
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`

	// ParameterChecksum is the checksum of the parameters that were sent.
	ParameterChecksum string `json:"parameterChecksum,omitempty"`

	// UserInfo is information about the user that made the request.
	UserInfo *UserInfo `json:"userInfo,omitempty"`
}

// ParametersFromSource represents the source of a set of Parameters
type ParametersFromSource struct {
	// The Secret key to select from.
	// The value must be a JSON object.
	// +optional
	SecretKeyRef *SecretKeyReference `json:"secretKeyRef,omitempty"`
}

// SecretKeyReference references a key of a Secret.
type SecretKeyReference struct {
	// The name of the secret in the pod's namespace to select from.
	Name string `json:"name"`
	// The key of the secret to select from.  Must be a valid secret key.
	Key string `json:"key"`
}

// ObjectReference contains enough information to let you locate the
// referenced object.
type ObjectReference struct {
	// Namespace of the referent.
	Namespace string `json:"namespace,omitempty"`
	// Name of the referent.
	Name string `json:"name,omitempty"`
}

// LocalObjectReference contains enough information to let you locate the
// referenced object inside the same namespace.
type LocalObjectReference struct {
	// Name of the referent.
	Name string `json:"name,omitempty"`
}

// ClusterObjectReference contains enough information to let you locate the
// cluster-scoped referenced object.
type ClusterObjectReference struct {
	// Name of the referent.
	Name string `json:"name,omitempty"`
}

// Filter path for Properties
const (
	// Name field.
	FilterName = "name"
	// SpecExternalName is the external name of the object.
	FilterSpecExternalName = "spec.externalName"
	// SpecExternalID is the external id of the object.
	FilterSpecExternalID = "spec.externalID"
	// SpecServiceBrokerName is used for ServiceClasses, the parent service broker name.

	FilterSpecServiceBrokerName = "spec.serviceBrokerName"
	// SpecClusterServiceBrokerName is used for ClusterServiceClasses, the parent service broker name.
	FilterSpecClusterServiceBrokerName = "spec.clusterServiceBrokerName"

	// SpecServiceClassName is only used for plans, the parent service class name.
	FilterSpecServiceClassName = "spec.serviceClass.name"
	// SpecClusterServiceClassName is only used for plans, the parent service class name.
	FilterSpecClusterServiceClassName = "spec.clusterServiceClass.name"
	// SpecClusterServiceClassRefName is only used for plans, the parent service class name.
	FilterSpecServiceClassRefName = "spec.serviceClassRef.name"
	// SpecClusterServiceClassRefName is only used for plans, the parent service class name.
	FilterSpecClusterServiceClassRefName = "spec.clusterServiceClassRef.name"

	// SpecServicePlanRefName is only used for instances.
	FilterSpecServicePlanRefName = "spec.servicePlanRef.name"
	// SpecClusterServiceClassRefName is only used for instances.
	FilterSpecClusterServicePlanRefName = "spec.clusterServicePlanRef.name"

	// FilterSpecFree is only used for plans, determines if the plan is free.
	FilterSpecFree = "spec.free"
)

// SecretTransform is a single transformation that is applied to the
// credentials returned from the broker before they are inserted into
// the Secret associated with the ServiceBinding.
// Because different brokers providing the same type of service may
// each return a different credentials structure, users can specify
// the transformations that should be applied to the Secret to adapt
// its entries to whatever the service consumer expects.
// For example, the credentials returned by the broker may include the
// key "USERNAME", but the consumer requires the username to be
// exposed under the key "DB_USER" instead. To have the Service
// Catalog transform the Secret, the following SecretTransform must
// be specified in ServiceBinding.spec.secretTransform:
// - {"renameKey": {"from": "USERNAME", "to": "DB_USER"}}
// Only one of the SecretTransform's members may be specified.
type SecretTransform struct {
	// RenameKey represents a transform that renames a credentials Secret entry's key
	RenameKey *RenameKeyTransform `json:"renameKey,omitempty"`
	// AddKey represents a transform that adds an additional key to the credentials Secret
	AddKey *AddKeyTransform `json:"addKey,omitempty"`
	// AddKeysFrom represents a transform that merges all the entries of an existing Secret
	// into the credentials Secret
	AddKeysFrom *AddKeysFromTransform `json:"addKeysFrom,omitempty"`
	// RemoveKey represents a transform that removes a credentials Secret entry
	RemoveKey *RemoveKeyTransform `json:"removeKey,omitempty"`
}

// RenameKeyTransform specifies that one of the credentials keys returned
// from the broker should be renamed and stored under a different key
// in the Secret.
// For example, given the following credentials entry:
//     "USERNAME": "johndoe"
// and the following RenameKeyTransform:
//     {"from": "USERNAME", "to": "DB_USER"}
// the following entry will appear in the Secret:
//     "DB_USER": "johndoe"
type RenameKeyTransform struct {
	// The name of the key to rename
	From string `json:"from"`
	// The new name for the key
	To string `json:"to"`
}

// AddKeyTransform specifies that Service Catalog should add an
// additional entry to the Secret associated with the ServiceBinding.
// For example, given the following AddKeyTransform:
//     {"key": "CONNECTION_POOL_SIZE", "stringValue": "10"}
// the following entry will appear in the Secret:
//     "CONNECTION_POOL_SIZE": "10"
// Note that this transform should only be used to add non-sensitive
// (non-secret) values. To add sensitive information, the
// AddKeysFromTransform should be used instead.
type AddKeyTransform struct {
	// The name of the key to add
	Key string `json:"key"`
	// The binary value (possibly non-string) to add to the Secret under the specified key. If both
	// value and stringValue are specified, then value is ignored and stringValue is stored.
	Value []byte `json:"value"`
	// The string (non-binary) value to add to the Secret under the specified key.
	StringValue *string `json:"stringValue"`
	// The JSONPath expression, the result of which will be added to the Secret under the specified key.
	// For example, given the following credentials:
	// { "foo": { "bar": "foobar" } }
	// and the jsonPathExpression "{.foo.bar}", the value "foobar" will be
	// stored in the credentials Secret under the specified key.
	JSONPathExpression *string `json:"jsonPathExpression"`
}

// AddKeysFromTransform specifies that Service Catalog should merge
// an existing secret into the Secret associated with the ServiceBinding.
// For example, given the following AddKeysFromTransform:
//     {"secretRef": {"namespace": "foo", "name": "bar"}}
// the entries of the Secret "bar" from Namespace "foo" will be merged into
// the credentials Secret.
type AddKeysFromTransform struct {
	// The reference to the Secret that should be merged into the credentials Secret.
	SecretRef *ObjectReference `json:"secretRef,omitempty"`
}

// RemoveKeyTransform specifies that one of the credentials keys returned
// from the broker should not be included in the credentials Secret.
type RemoveKeyTransform struct {
	// The key to remove from the Secret
	Key string `json:"key"`
}

func init() {
	// SchemaBuilder is used to map go structs to GroupVersionKinds.
	// Solution suggested by the Kubebuilder book: https://book.kubebuilder.io/basics/simple_resource.html - "Scaffolded Boilerplate" section
	SchemeBuilderRuntime.Register(
		&ServiceBinding{},
		&ServiceInstance{},
		&ClusterServiceClass{},
		&ClusterServiceClassList{},
		&ServiceBroker{},
		&ClusterServiceBroker{},
		&ServiceClass{},
		&ServiceClassList{},
		&ServicePlan{},
		&ServicePlanList{},
		&ClusterServicePlan{},
		&ClusterServicePlanList{})
}
