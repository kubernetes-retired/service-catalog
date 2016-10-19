# V1 Architecture

This document contains the architectural design for the v1 `service-catalog`.

# Resource Types

This section lists descriptions of Kubernetes resource types.

## `ServiceClass`

This resource indicates a kind of backing service that a consumer may request.

## `ServiceInstance`

This resource indicates that a request by a consumer for a usable `ServiceClass`
has been successfully executed. Consumers may reference these resources to
begin using the backing service it represents.

## `ServiceInstanceClaim`

This resource is used by the consumer to get credentials for the backing service
that a pre-existing `ServiceInstance` represents.

## `ServiceInstanceBinding`

This resource is a byproduct of a successfully executed `ServiceInstanceClaim`.
It contains the following information:

1. A record of what `ServiceInstanceClaim` was successfully executed
1. A list of Kubernetes-style reference links for each Kubernetes resource
   that was created to hold binding information (such as authentication data).
   We expect `ServiceInstanceBinding`s to hold links to `ConfigMap`s and
   `Secret`s initially, but the number and types of these resources will be
   implementation-dependent.
