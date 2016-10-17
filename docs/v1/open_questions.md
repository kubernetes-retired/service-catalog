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
some functionality, a library, or something else to make this implementation
easier and less error-prone?

## 5. How will quotas work?

The section entitled "Quotas" is TBD. It needs far more thought and detail.

## 6. Can we provide a reference implementation of a controller?

This is a question I have an answer to! We should provide an implementation for
backing CloudFoundry service brokers. Not only should it be a reference, it
should be the preferred controller implementation for all CloudFoundry service
broker use-cases.

Finally, we should expect it to be used in a wide variety of
situations. It should be highly configurable and able to handle the widest
variety of use-cases possible.
