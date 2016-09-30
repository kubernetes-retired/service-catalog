## V1 Use Cases

This document contains the complete list of accepted use-cases for the v1
version of `service-catalog`.

## High-Level Use Cases

These are the high-level, user-facing use cases the v1 service catalog will
implement.

1.  Sharing services:
    1.  (Blackbox services) As a SaaS provider that already runs a service
        broker, I want users of Kubernetes to be able to use my service
        via the service broker API, so that I can grow my user base to
        include users of Kubernetes
    2.  As the operator of an existing service running in Kubernetes, I want to
        be able to publish my services using a service broker, so that users
        external to my Kubernetes cluster can use my service

### Blackbox services

There are numerous SaaS providers that already operate service brokers today.
It should be possible for the operator of an existing service broker to
publish their services into the catalog and have them consumed by users of
Kubernetes.  This offers a new set of users to the service operator and offers
a wide variety of SaaS products to users of Kubernetes.

### Exposing existing services outside the cluster

It should be possible for service operators whose services are deployed in a
Kubernetes cluster to publish their services using a service broker.  This
would allow these operators to participate in the existing service broker
ecosystem and grow their user base accordingly.

## Low-Level Use Cases

These are lower-level use cases the service catalog will implement in service
to the high-level use cases.

1.  As a service broker operator, I want to be able to register my broker with
    the Kubernetes service catalog, so that the catalog is aware of the services
    my broker offers

### Registering a service broker with the catalog

Each service broker must register with the service catalog to advertise the
services it offers to the catalog.  After the broker has registered with the
catalog, the catalog makes a call to the service broker's `/v2/catalog`
endpoint.  The broker's response to this call is a list of services offered by
that broker.

For more information, see the
[Cloud Foundry documentation on registering service brokers](https://docs.cloudfoundry.org/services/managing-service-brokers.html#register-broker).
