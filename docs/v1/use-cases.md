## V1 Use Cases

This document contains the complete list of accepted use-cases for the v1
version of the service catalog.

## SIG Meeting

### Use Cases

1.  Adding service brokers into catalog
2.  Catalog curation: which broker-provided services are globally visible?
3.  Catalog curation: which namespaces can see which catalog services?
4.  As a consumer of a service, how do I indicate that I want to consume a particular service?
5.  As a consumer of a service, how do I indicate that I want to bind to an instance of a service?
6.  What is the unit of consumption of a service?  Namespace?  Pod? Something else? (brian to comment)
7.  What is the versioning and update story for a service: example: what happens when a broker changes the services it provides?
8.  When a user binds to an instance of a service, what does that look like?  Which resources are created in their namespace?
9.  What is the update story for bindings to a service instance?
10. How does the catalog support multiple consumers in different Kubernetes namespaces of the same instance of a service? 
11. How do I control the number and identity of consumers allowed to bind to my instance?
11. Curation: how do I control who can see a service instance that I have provisioned?
12. Curation: how do I control the number of consumers of my service instance?

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

### Sharing blackbox services

There are numerous SaaS providers that already operate service brokers today.
It should be possible for the operator of an existing service broker to
publish their services into the catalog and have them consumed by users of
Kubernetes.  This offers a new set of users to the service operator and offers
a wide variety of SaaS products to users of Kubernetes.

### Exposing Kubernetes services outside the cluster

It should be possible for service operators whose services are deployed in a
Kubernetes cluster to publish their services using a service broker.  This
would allow these operators to participate in the existing service broker
ecosystem and grow their user base accordingly.

## CF Service Broker `v2` API Use Cases

Initially, the catalog should support the current [CF Service Broker
API](https://docs.cloudfoundry.org/services/api.html) These are the use cases
that the service catalog has to implement in order to use that API.

### Managing service brokers

1.  As user, I want to be able to register a broker with the Kubernetes service
    catalog, so that the catalog is aware of the services that broker offers
2.  As a user, I want to be able to update a registered broker so that the
    catalog can maintain the most recent versions of services that broker offers
3.  As a user, I want to be able to delete a broker from the catalog, so that I
    can keep the catalog clean of brokers I no longer want to support

#### Registering a service broker with the catalog

An user must register each service broker with the service catalog to
advertise the services it offers to the catalog.  After the broker has been
registered with the catalog, the catalog makes a call to the service broker's
`/v2/catalog` endpoint.  The broker's returns a list of services offered by
that broker.  Each Service has a set of plans that differentiate the tiers of
that service.

#### Updating a service broker

Broker authors make changes to the services their brokers offer.  To refresh the
services a broker offers, the catalog should re-list the `/v2/catalog` endpoint.
The catalog should apply the result of re-listing the broker to its internal
representation of that broker's services:

1.  New service present in the re-list results are added
2.  Existing services are updated if a diff is present
3.  Existing services missing from the re-list are deleted

TODO: spell out various update scenarios and how they affect end-users

#### Delete a service broker

There must be a way to delete brokers from the catalog.  In Cloud Foundry, it is
possible to delete a broker and leave orphaned service instances.  We should
evaluate where all broker deletes should:

1.  Cascade down to the service instances for the broker
2.  Leave orphaned service instances in the catalog
3.  Fail if service instances still exist for the broker

## Supporting multiple backend APIs

The CF service broker API is under active development, leading to two
possibilities that may both occur:

1.  The `v2` API undergoes backward-compatible changes
2.  There is a new `v3` API that is not backward-compatible

The service catalog should be able to support new backward-compatible fields or
a new backend API without a major rewrite.  This should be kept in mind when
designing the architecture of the catalog.


For more information, see the
[Cloud Foundry documentation on registering service brokers](https://docs.cloudfoundry.org/services/managing-service-brokers.html#register-broker).
