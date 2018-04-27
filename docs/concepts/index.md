---
title: Concepts
layout: docwithnav
---

## Introduction

The service-catalog project is in incubation to bring integration with service
brokers to the Kubernetes ecosystem via the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker).

A _service broker_ is an endpoint that manages a set of software offerings
called _services_. The end-goal of the service-catalog project is to provide
a way for Kubernetes users to consume services from brokers and easily
configure their applications to use those services, without needing detailed
knowledge about how those services are created or managed.

As an example:

Most applications need a datastore of some kind. The service-catalog allows
Kubernetes applications to consume services like databases that exist
_somewhere_ in a simple way:

1. A user wanting to consume a database in their application browses a list of
    available services in the catalog
2. The user asks for a new instance of that service to be _provisioned_

    _Provisioning_ means that the broker somehow creates a new instance of a
   service. This could mean basically anything that results in a new instance
   of the service becoming available. Possibilities include: creating a new
   set of Kubernetes resources in another namespace in the same Kubernetes
   cluster as the consumer or a different cluster, or even creating a new
   tenant in a multi-tenant SaaS system. The point is that the
   consumer doesn't have to be aware of or care at all about the details.
3. The user requests a _binding_ to use the service instance in their application

    Credentials are delivered to users in normal Kubernetes secrets and
    contain information necessary to connect to and authenticate to the
    service instance.
    
## Overview

The service catalog API has five main concepts:

- Open Service Broker API Server: A server that acts as a service broker and
conforms to the
[Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md)
specification. This software could be hosted within your own Kubernetes cluster
or elsewhere.

The remaining four concepts all map directly to new Kubernetes resource types
that are provided by the service catalog API.

- `ClusterServiceBroker`: An in-cluster representation of a broker server. A
resource of this type encapsulates connection details for that broker server.
These are created and managed by cluster operators who wish to use that broker
server to make new types of managed services available within their cluster.
- `ClusterServiceClass`: A *type* of managed service offered by a particular
broker. Each time a new `ClusterServiceBroker` resource is added to the cluster,
the service catalog controller connects to the corresponding broker server to
obtain a list of service offerings. A new `ClusterServiceClass` resource will
automatically be created for each.
- `ServiceInstance`: A provisioned instance of a `ClusterServiceClass`. These
are created by cluster users who wish to make a new concrete _instance_ of some
_type_ of managed service to make that available for use by one or more
in-cluster applications. When a new `ServiceInstance` resource is created, the
service catalog controller will connect to the appropriate broker server and
instruct it to provision the service instance.
- `ServiceBinding`: Expresses intent to use a `ServiceInstance`. These are
created by cluster users who wish for their applications to make use of a
`ServiceInstance`. Upon creation, the service catalog controller will create a
Kubernetes `Secret` containing connection details and credentials for the
service represented by the `ServiceInstance`. Such `Secret`s can be used like
any other-- mounted into a container's file system or injected into a container
as environment variables.

These concepts and resources are the building blocks of the service catalog.

## Service Resources

See [Resources](../resources.md) for details on each Service Catalog resource (or type).
