# Open Questions

This document contains open questions regarding the
[API design document](./design.md).

## 1. Is the `ServiceInstanceBinding` resource necesssary?

The `ServiceInstanceBinding` object is a byproduct of the successful execution
of a `ServiceInstanceClaim`. Instead of writing a new `ServiceInstanceBinding`
object, can the service catalog controller just add the binding information
directly into the `ServiceInstanceClaim` on success?

## Should we write Kubernetes references for binding information?

Currently, the information in a `ServiceInstanceBinding` that points to created
resources looks like this:

```yaml
secrets:
  - my-zip-code-creds-password
configMaps:
  - my-zip-code-creds-conn-info
```

Instead, should we write pointers to these resources using Kubernetes reference
notation, as below?

```yaml
resources:
  - /api/v1/namespaces/${NAMESPACE}/secrets/my-zip-code-creds-password
  - /api/v1/namespaces/${NAMESPACE}/configmaps/my-zip-code-creds-conn-info
```

## 2. Should `ServiceInstance`s be namespaced?

If `ServiceInstance`s were namespaced, then operators could enforce their
security with [RBAC](http://kubernetes.io/docs/admin/authorization/#rbac-mode).

If we do namespace `ServiceInstance`s, however, a few questions arise:

1. It would become messy for pods to share a `ServiceInstance` inter-namespace
1. A consumer may still submit a `ServiceInstanceClaim` that refers to a
  `ServiceInstance` that is in a RBAC-forbidden namespace.
  1. Possible solution: the controller enforces security, as is the case with
     the current design doc

## 3. Should we keep `ProvisioningPolicy` and `BindingPolicy`?

As mentioned in (2), we may use RBAC to replace the functionality that these
two policies offered. However, if `ServiceClass` and/or `ServiceInstance`
remain cluster-global, we'll likely need some way to restrict certain
claims (likely all claims in a given namespace) from successfully executing
actions against one or more `ServiceClass`es / `ServiceInstance`s

## 4. Can we reduce the security logic each controller must implement?

As the document stands, service controllers will need to implement logic to
enforce at least `ProvisioningPolicy` and `BindingPolicy` objects. This logic
will be non-trivial, redundant if every controller has to implement it, and
error-prone.

Regardless of whether we keep these policy objects (see (3) above), can we add
some functionality, a library, or make some other design changes to ensure this
implementation is easier and less error-prone?

## 5. How will quotas work?

The section entitled "Quotas" is TBD. It needs more thought and detail.

## 6. Should we have a `ServiceController` resource?

Currently, we don't surface the existence or function of service catalog
controllers in the cluster. They've been abstracted away from consumers, and
they must maintain a list internally of `ServiceClass`es that they support.

They must do so to determine which `ServiceInstanceClaim`s they should take
action on.

Should we add a resource, called `ServiceController` or similar, to do the
below?

1. Surface the existence of service controllers in the cluster
  (we expect only operators to use this information)
1. Provide an in-cluster record of the existence of each service controller
1. Provide a resource for each `ServiceClass` to reference, indicating that
   all operations on it should be done by the referenced controller

## 7. Should we split `ServiceInstanceClaim`s into two distinct types?

Currently, `ServiceInstanceClaim`s can refer to either a `ServiceClass` or
`ServiceInstance`. In the former case, a controller will do a provision and a
bind, ensuring that the provisioned resource is dedicated to the consumer.
In the latter case, a controller will just do a bind, allowing the
already-provisioned resource to be shared among many consumers.

These very-different behaviors are toggled by a simple change to the same
resource kind. Should we split this resource into separate ones to make clear
the difference between referring to a `ServiceClass` and a `ServiceInstance`.

The resulting two resources may be called `ServiceClassClaim` and
`ServiceInstanceClaim`.

## 7. Can we provide a reference implementation of a controller?

This is a question I have an answer to! We should provide an implementation for
backing CloudFoundry service brokers. Not only should it be a reference, it
should be the preferred controller implementation for all CloudFoundry service
broker use-cases.

Finally, we should expect it to be used in a wide variety of
situations. It should be highly configurable and able to handle the widest
variety of use-cases possible.
