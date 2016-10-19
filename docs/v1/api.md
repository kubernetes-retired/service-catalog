# Service Catalog API

This document details a standard API, including a schema for standard
Kubernetes-native resources, to:

- Formalize an API around the service broker concept as a first-class Kubernetes
  (or Kubernetes-like) API for all to develop against.
- Follow established patterns in the Kubernetes core API for globally-available
  and namespaced resources
- Facilitate the development of Kubernetes-like controllers in support of the
  API, without undermining competition or choice
- Ensure a smooth, consistent UX for users of the API and implementors of the
  controllers

# API Precedents in Kubernetes

Broadly speaking, the system detailed herein can list services and allow
consumers to provision, bind, unbind and deprovision them.

We've drawn inspiration from the
[Persistent Volumes](http://kubernetes.io/docs/user-guide/persistent-volumes/)
API and the [Ingress Controllers](http://kubernetes.io/docs/user-guide/ingress/)
system. The remainder of this section explains these inspirations in further
detail.

Additionally, we've adhered, wherever possible, to Kubernetes's declarative
API. That is, the resources we've defined herein describe the _desired end
state_ of the applicable systems. Kubernetes and the service controller
implementations (see below) are responsible for fulfilling that end state.

## Persistent Volumes

First, this problem space is similar to that of persistent volumes. We model
the data structures herein after those in persistent volumes primarily because
we believe that consistency is a virtue. More importantly, the community
benefits when similar problems have similar solutions.

Specifically, we believe that the following existing types provide a good model
for the mechanics of how the services offered by a broker can be used in
Kubernetes:

- `StorageClass`
- `PersistentVolume`
- `PersistentVolumeClaim`

## Ingress Controllers

Today, operators choose their own implementation of an ingress controller.
Kubernetes gives the operator the ability to specify any ingress controller,
and the complete freedom to implement it themselves.

Similarly, permitting operators to implement or select their own service catalog
controller would encourage freedom of choice and diversity. The analog to an
ingress controller for the service catalog system is called a _service catalog
controller_. Like ingress, these controllers run inside a standard pod. Anyone
can write and deploy a service catalog controller, with no changes necessary to
the Kubernetes install.

Service catalog controllers vary, however, from ingress controllers, because a
cluster can consist of many different controllers, each of which handles actions
on a subset of the available service-related resources. For example, a cluster
can run a service controller to handle AWS RDS databases, and a completely
separate one to handle AWS Elastic Load Balancer instances.

As we'll detail below, each service controller implementation must to conform
to an implementation spec, but that spec allows for a heterogenous set of
controllers in a cluster. We believe this allowance encourages flexibility in
a cluster's service catalog.

# Kubernetes Resources

Given these inspirations, the following resources will be added to the
Kubernetes API:

- `ServiceClass`
  - Describes a _kind_ of service. This resource is semantically similar to how
    a `StorageClass` defines a _kind_ of persistent volume, but we expect that,
    in practice, the kinds of services represented here will be more coarse-
    grained than the kinds of volumes represented in `StorageClass`es
  - References a service ID and plan ID, which is managed and understood by a
    service controller (see below)
  - Created by an operator or tool
- `ServiceInstance`
  - An instance of a `ServiceClass` -- an in-cluster representation of a
    provisioned service
  - Can be statically provisioned by an operator, to enable multiple consumers
    to bind to it (enables multi-tenancy)
  - Can be dynamically provisioned and bound to by a consumer, to guarantee an
    app’s exclusive use of the resource (guarantees single-tenancy)
- `ServiceClassClaim`
  - Created by a consumer to provision an instance of a `ServiceClass`, then
    bind to the provisioned. Enables dedicated, single-tenant usage of a
    provisioned resource
- `ServiceClassBinding`
  - Created in response to a successfully executed `ServiceClassClaim`
  - References credentials created as a result of the bind operation
  - Holds a record of an application that created a `ServiceClassClaim` which
    was successfully executed
- `ServiceInstanceClaim`
  - Created by a consumer to bind to an existing `ServiceInstance` by name.
    Enables multi-tenant usage of a provisioned resource
- `ServiceInstanceBinding`
  - Created in response to a successfully executed `ServiceInstanceClaim`
  - References credentials created as a result of the bind operation
  - Holds a record of an application that created a `ServiceInstance` which
    was successfully executed


The below diagram shows the interactions between these resources:

![Resource Interactions](./img/partial-flow.png)

# Example Resources

## `ServiceClass`

```yaml
kind: ServiceClass
apiVersion: service-catalog.k8s.io/alpha
metadata:
  name: postgres-small
spec:
  serviceID: c5a0ad8e-b57b…
  planID: c68cd6b1-be9f…
```

## `ServiceInstance`

```yaml
kind: ServiceInstance
apiVersion: service-catalog.k8s.io/alpha
metadata:
  name: zip-code-db
spec:
  serviceClass: postgres-small
```

## Single Tenant / Dedicated Resources

The `ServiceClassClaim` and `ServiceClassBinding` resources are created
to give an application single-tenant, dedicated access to a backing resource.

The developer should create a `ServiceClassClaim` resource, and the system
should do the `provision` operation, then `bind` immediately thereafter.

### `ServiceClassClaim`

```yaml
kind: ServiceClassClaim
apiVersion: service-catalog.k8s.io/alpha
metadata:
  name: my-app-db-claim
  namespace: my-app
status: unknown
spec:
  serviceClass: postgres-large
  targetBinding: my-app-binding
  status: unknown # only written by the service controller
  statusDescription: unknown # only written by the service controller
```

### `ServiceClassBinding`

This example shows the output of a single-tenant `ServiceInstanceClaim`.

```yaml
kind: ServiceInstanceBinding
apiVersion: service-catalog.k8s.io/alpha
metadata:
  name: my-app-binding # same as the previous 'targetBinding' field
  namespace: my-app
status:
  serviceClass: postgres-large
  secrets:
    - my-zip-code-creds-password
  configMaps:
    - my-zip-code-creds-username
    - my-zip-code-creds-conn-info
```

## Multi Tenant / Shared Resources

The `ServiceInstanceClaim` and `ServiceInstanceBinding` resources are created
to give an application multi-tenant, shared access to a backing resource.

The developer should create a `ServiceInstanceClaim` resource, and the system
should do the `bind` operation. It's assumed that an operator will have already
created the `ServiceInstance` to which the `ServiceInstanceClaim` and
`ServiceInstanceBinding` refers.


### `ServiceInstanceClaim`

```yaml
kind: ServiceInstanceClaim
apiVersion: service-catalog.k8s.io/alpha
metadata:
  name: my-zip-code-db-claim
  namespace: my-app
spec:
  serviceInstance: zip-code-db
  targetBinding: my-zip-code-binding
  status: unknown # only written by the service controller
  statusDescription: unknown # only written by the service controller
```

### `ServiceInstanceBinding`

This example shows the output of a multi-tenant `ServiceInstanceClaim`.

```yaml
kind: ServiceInstanceBinding
apiVersion: service-catalog.k8s.io/alpha
metadata:
  name: my-zip-code-binding # same as the above 'targetBinding' field
  namespace: my-app
status:
  serviceInstance: zip-code-db
  secrets:
    - my-zip-code-creds-password
  configMaps:
    - my-zip-code-creds-username
    - my-zip-code-creds-conn-info
```

# The Service Catalog Controller

As indicated above, application developers and cluster operators interact with
the aforementioned Kubernetes resources. The service catalog controller watches
these resources and takes action on them.

Like ingress controllers, the operator is free to choose any service controller
they prefer. All implementations, however, must satisfy the following
requirements:

- Watch the event stream for new or deleted `ServiceInstance`s and
  `ServiceInstanceClaim`s
- On `ServiceInstance` creation, the controller must do a provision
- On `ServiceInstance` deletion, the controller must do a deprovision. The
  controller should not delete any existing `ServiceInstanceClaim`s or
  `ServiceInstanceBinding`s that reference the deleted `ServiceInstance`
- On `ServiceInstanceClaim` creation:
  - If the claim references an existing `ServiceInstance`, the controller must
    do a bind
  - If the claim references an existing `ServiceClass`, the controller must do
    a provision, then bind. The controller should not create a
    publicly-accessible `ServiceInstance`. This ensures that the
    provisioned/bound resource will be single-tenant and dedicated
- On `ServiceInstanceClaim` deletion:
  - If the claim referenced an existing `ServiceInstance`, the controller must
    do an unbind
  - If the claim referenced an existing `ServiceClass`, the controller must do
    an unbind, then a deprovision
- Any successful provision operation must change the `status` field of the
  applicable `ServiceInstanceClaim` to `provisioned`
- Any successful bind operation must:
  - Change the `ServiceInstanceClaim`'s `status` field to `bound`
  - Create the appropriate resources (`Secret`s, `ConfigMap`s) to hold
    credential and other resource-specific bind information
  - Create a new `ServiceInstanceBinding`:
    - With the same name as the claim’s `targetBinding` field
    - In the same namespace as the claim itself
    - With fields that point to all the previously created resources
- Any successful unbind operation must:
  - Change the claim’s `status` field to `unbound`
  - Delete the `ServiceInstanceBinding` that was created in the bind operation
  - Delete all resources that the aforementioned `ServiceInstanceBinding`
    points to
- Any failed operation must change the status field of the
  `ServiceInstanceClaim` to `failed`, and should write the reason of the
  failure into the `statusDescription` field

The below diagram shows all of the interactions that all service catalog
controllers are expected to have.

![Complete System](./img/entire-flow.png)

# Visibility

The visibility of each resource and system described above is important to the
functionality and ease-of-use in this system.

The aforementioned Kubernetes resources share many of the same visibility rules
as those in the persistent volumes and ingress systems.

- `ServiceClasse`s are cluster-global (just as `StorageClass`es are)
- `ServiceInstance`s are cluster-global (just as `PersistentVolume`s are)
- `ServiceInstanceClaim`s are namespace-scoped (just as
  `PersistentVolumeClaim`s are)
- `ServiceInstanceBinding`s are namespace-scoped
- Service catalog controllers may run on any number of namespaces, including
  cluster-global

# Security

We expect that many cluster operators will need to restrict usage of many
provisionable and bindable resources. We provide several mechanism for
implementing these restrictions:

## `ProvisioningPolicy`

Since `ServiceClass`es are cluster-global, operators must be able to restrict
the namespaces that can provision and bind to each `ServiceClass`. A
`ProvisioningPolicy` is a cluster-global resource that holds a blacklist that
contains the namespaces that cannot provision and bind to each `ServiceClass`.
A few additional notes:

- This resource is optional. If it doesn’t exist, there will be no blacklist
  applied
- In the future, we may add a whitelist or other more advanced filtering
  features

## `BindingPolicy`

Since `ServiceInstance`s are cluster-global, operators must be able to restrict
the namespaces that can bind to each `ServiceInstance`. A `BindingPolicy` is a
cluster-global resource that holds a blacklist that contains the namespaces
that cannot bind to each `ServiceInstance`. A few additional notes:

- This resource is optional. If it doesn’t exist, there will be no blacklist
  applied
- In the future, we may add a whitelist or other advanced filtering features

## Quotas

Provision and bind operations may be subject to namespace-scoped quotas,
similar to those already in existence on other resources in Kubernetes. This
section is in need of additional detail.

## Final Note

Service catalog controllers will be responsible for enforcing both the policies
and the quotas.
