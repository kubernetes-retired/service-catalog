# Proposal: Controlling access to Services and Plans

## Abstract

Proposes changes to Service Catalog to facilitate controlling access to certain services and plans.

## Motivation

Not all services and plans should be available to all users. The existing cluster-scoped resources for brokers, services, and plans are not sufficient to implement access control to ensure that users have access only to the service and plans that they should.

Additionally, when developers are creating new services that are exposed by brokers, they want to be able to iterate on those services without exposing them to all users in a cluster.

There are two ingredients to successfully controlling access to services and plans:

- Namespaced versions of these resources are required to control access to services along the boundaries of namespaces
- API surfaces to allow black/whitelisting which services and plans from a broker's catalog have k8s resources created for them

## Use Cases

- As a cluster operator, I want to control access to certain services and plans so that only certain namespaces are allowed to use them
- As a developer, I want to be able to add a broker just to my namespace, so that I can iterate on new service I am developing without exposing services and plans to users of the cluster

### Use Case: Controlling access to services and plans 

Certain services or plans are not suitable for consumption by every user of a cluster. For example, a service may have a monetary cost associated with it or grant the user a high degree of privilege. In order to prevent users from gaining access to services and/or plans that they should not be able to use, we must be able to:

1. Keep the services and plans in the cluster-scoped catalog limited to those that anyone can use
2. Allow services and plans that are only for certain users to be used only by the users that should have access to them

For example: a broker may offer a highly-privileged service that ordinary users should not even be aware of, let alone allowed to use. In this case, the cluster administrator should be able to keep that service from appearing in the cluster-scoped catalog of services and plans but also make it available to users with the appropriate level of privilege. 

### Use Case: Developing new services

Similar to highly privileged services, there are a class of users who are service developers, and would be interested in publishing their services via a broker to the rest of the cluster. As part of their development cycle, they would like to work within their own private namespace while iterating on their service.

## Goals and non-goals

There are many related problems in Service Catalog that users are interested in solutions for, but we need to keep the scope of this proposal controlled. In that light, let's establish the goals of this proposal:

- Make it possible to keep the cluster-scoped catalog limited to services and plans that everyone should be able to use
- Make it possible to add services and plans into specific namespaces

The following are valid goals but outside the scope of this proposal:

- Make it possible to use a service and plan from namespace X to provision a `ServiceInstance` in namespace Y
- Allow creating a `ServiceBinding` in namespace X to a `ServiceInstance` in namespace Y
- Expose a service and plan in namespace X to the cluster scope
- Expose a service and plan in namespace X to namespace Y
- Create a policy that adds certain services and plans to multiple namespaces
- Creating a virtual resource that allows users to see all services or plans available to them across the cluster and namespace scopes

We should take care to achieve the goals of this proposal without preventing further progress on other issues that are out of scope.

## Analysis

### Why namespaces?

Unfortunately, it is not possible in Kubernetes to create an ACL (access control list) filtering scheme that shows users only the resources in a collection that they are allowed to see. The fundamental gaps here are:

1. Not all authorizers are able to provide a list of subjects with access to a resource
2. An external authorizer may have its state changed at any time out of band to kubernetes, making it impossible to do implement a correct `LIST` or `WATCH` operation from a certain resource version

Since ACL-filtering the cluster-scoped list of services and plans is not a realistic option, we must find another method of controlling read/write access to resources. In Kubernetes, namespaces are the defacto way of performing this type of access control.

Adding namespaced resources for service brokers (`ClusterServiceBroker`), services (`ClusterServiceClass`) and plans (`ClusterServicePlans`) allows us to take advantage of the existing namespace concept to perform access control.

### Filtering services and plans from a broker

Adding namespaced resources for brokers, services and plans is necessary but not sufficient to control access to services and plans. A single broker may offer a mix of services that all users should be able to access and services that should only be usable by some users.

In order to prevent a broker that offers a mix of unprivileged and privileged services to the cluster-scoped catalog, there must be a way to filter the services and plans exposed by a broker. This can be accomplished through the use of white/black lists that control which services and plans in a broker's catalog have Service Catalog resources created for them. For example:

- A cluster administrator should be able to prevent privileged services from appearing in cluster-scoped catalog
- A cluster administrator should be able to add certain privileged services to a namespace

## Design

In this proposal we'll focus on adding the namespaced `ServiceBroker`, `ServiceClass`, and `ServicePlan` resources. For details on filtering which services and plans in a broker's catalog have k8s resources created for them, see https://github.com/kubernetes-incubator/Service Catalog/pull/1773.

### Namespaced resources

The namespaced resources for brokers, services, and plans should have the same behaviors as their cluster-scoped cousins. To a great degree, we can reuse the same API fields in the namespaced resources, but there are some exceptions:

#### `ServiceBroker` resource

The API for the `ServiceBroker` resource should differ from `ClusterServiceBroker` in exactly one area:

- A user should only be able to specify a secret within the same namespace to hold the auth information

#### `ServiceClass` resource

Differences between `ClusterServiceClass` and `ServiceClass`:

- `ServiceClass.Spec` should have `ServiceBrokerName` instead of `ClusterServiceBrokerName`

#### `ServicePlan` resource

Differences between `ClusterServicePlan` and `ServicePlan`

- `ServicePlan.Spec` should have `ServiceBrokerName` instead of `ClusterServiceBrokerName`
- `ServicePlan.Spec` should have `ServiceClassRef` instead of `ClusterServiceClassRef`

### Changes to `ServiceInstance`

The `ServiceInstance` resource should be changed to allow users to unambiguously specify a `ServiceClass` and `ServicePlan` instead of the cluster-level resources.

- Add fields to `PlanReference` to represent the external and k8s names of `ServiceClass` and `ServicePlan` (as opposed to the cluster-scoped versions)
- Add reference fields to `ServiceInstanceSpec` that represent the references to the namespaced resources

## Implementation plan

The implementation of this proposal is too large for a single PR, so we'll break it into stages:

### Extracting shared fields from existing resources

To extract the shared fields, we will:

- Extract the identified shared fields onto embedable types: `SharedServiceBrokerSpec`, `SharedServiceClassSpec`, and `SharedServicePlanSpec`.
- Embed these shared types within their respective cluster scoped specs.
- Minor controller changes to reflect the fact that these fields now belong to an embedded type.
- Required updates to the fuzzer, validations, and defaults where necessary.

This PR will result in no behavioral changes, and should remain entirely transparent to users. It is effectively a no-op.

Relevant API types after this step:

```go

type ClusterServiceBrokerSpec struct {
    SharedServiceBrokerSpec `json:",inline"`
    
    AuthInfo *ClusterServiceBrokerAuthInfo `json:"authInfo,omitempty"`
}

type ClusterServiceBrokerStatus struct {
    SharedServiceBrokerStatus `json:",inline"`
}

type ClusterServiceClassSpec struct {
    SharedServiceClassSpec `json:",inline"`
    
    ClusterServiceBrokerName string `json:"clusterServiceBrokerName"`
}

type ClusterServiceClassStatus struct {
    SharedServiceClassStatus `json:",inline"`
}

type ClusterServicePlanSpec struct {
    SharedServicePlanSpec `json:",inline"`
    
    ClusterServiceBrokerName string `json:"clusterServiceBrokerName"`
    
    ClusterServiceClassRef ClusterObjectReference `json:"clusterServiceClassRef"`
}

type ClusterServicePlanStatus struct {
    SharedServicePlanStatus `json:",inline"`
}
```

### Add API surface for ns-scoped resources

After shared fields are extracted, we will add:

- API resources for the new namespace-scoped resources
- Associated validations / fuzzers / client changes / etc
- A feature gate that controls whether the ns-scoped resources are enabled

The feature gate is necessary to ensure that users do not see resources that aren't fully functional and will be removed after the proposal is completely implemented.

The new resources will look as follows:

```go
type ServiceBrokerSpec struct {
    SharedServiceBrokerSpec `json:",inline"`
    
    AuthInfo *ServiceBrokerAuthInfo `json:"authInfo,omitempty"`
}

type ServiceBrokerStatus struct {
    SharedServiceBrokerStatus `json:",inline"`
}

type ServiceClassSpec struct {
    SharedServiceClassSpec `json:",inline"`
    
    ServiceBrokerName string `json:"serviceBrokerName"`
}

type ServiceClassStatus struct {
    SharedServiceClassStatus `json:",inline"`
}

type ServicePlanSpec struct {
    SharedServicePlanSpec `json:",inline"`
    
    ServiceBrokerName `json:"serviceBrokerName"`
    
    ServiceClassRef LocalObjectReference `json:"serviceClassRef"`
}

type ServicePlanStatus struct {
    SharedServicePlanStatus `json:",inline"`
}
```

After this step, we will have the namespaced resources for brokers, services, and plans, but they will not be functional yet.

### Add control loops for ns-scoped resources

Next, we will add control loops for the new namespaced resources for brokers, services, and plans.

We'll also add associated tests (similar to broker integration tests that already exist).

After this step, users will be able to add a broker to a namespace and see the namespaced services and plans populated, but they won't be able to provision an instance of a namespaced service/plan yet.

### Make it possible to provision instances of ns-scoped services and plans

Next, we will make it possible to provision an instance of a namespaced service and plan. This involves:

- Adding new fields to `ServiceInstanceSpec` for users to indicate which ns-scoped services they want
- Adding new fields to `ServiceInstanceSpec` for references to ns-scoped resources
- Associated validations / fuzzers for those fields
- Modifying the reference subresource to set the new reference fields appropriately
- Modifying the controller to resolve the new specification fields and set reference fields

Relevant API resources at this step:

```go
type PlanReference struct {
    // existing fields omitted
    
    ServiceClassExternalName string `json:"serviceClassExternalName"`
    ServiceClassName         string `json:"serviceClassName"`
    ServicePlanExternalName  string `json:"servicePlanExternalName"`
    ServicePlanName          string `json:"servicePlanName"`
}

type ServiceInstanceSpec struct {
    // existing fields omitted

    ServiceClassRef *LocalObjectReference `json:"serviceClassRef,omitempty"`
    ServicePlanRef  *LocalObjectReference `json:"servicePlanRef,omitempty"`
}
```

After this step, users will be able to provision (but not bind to) ns-scoped services and plans.

### Make it possible to bind to instances of ns-scoped services and plans

Next, we will make the required changes to the binding controller loop to make it possible to create bindings against service instances that are associated with the ns-scoped variants of serviceclasses and plans.

### Remove feature gate

Finally, the feature gate will be removed to graduate this feature as enabled by default.