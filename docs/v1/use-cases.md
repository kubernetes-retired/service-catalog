# Service Catalog Use Cases

## Catalog Management

1. As a user, I want to be able to register a broker with the Kubernetes service
   catalog, so that the catalog is aware of the services that broker offers
2. As a user, I want to be able to update a registered broker so that the
   catalog can maintain the most recent versions of services that broker offers
3. As a user, I want to be able to delete a broker from the catalog, so that I
   can keep the catalog clean of brokers I no longer want to support

* Importing a list of brokers: If I have a remote cluster with a service
  controller that has n brokers registered, how can I import those into my local
  cluster?
* Can I export a list of service brokers and types from my service controller?
* Is there an auth story for adding service brokers?

### Registering a service broker with the catalog

A user must register each service broker with the service catalog to advertise
the services it offers to the catalog. After the broker has been registered
with the catalog, the catalog makes a call to the service broker's `/v2/catalog`
endpoint. The broker's returns a list of services offered by that broker. Each
Service has a set of plans that differentiate the tiers of that service.

### Updating a service broker

Broker authors make changes to the services their brokers offer. To refresh the
services a broker offers, the catalog should re-list the `/v2/catalog` endpoint.
The catalog should apply the result of re-listing the broker to its internal
representation of that broker's services:

1. New service present in the re-list results are added
2. Existing services are updated if a diff is present
3. Existing services missing from the re-list are deleted

TODO: spell out various update scenarios and how they affect end-users

### Delete a service broker

There must be a way to delete brokers from the catalog. In Cloud Foundry, it is
possible to delete a broker and leave orphaned service instances. We should
evaluate where all broker deletes should:

1. Cascade down to the service instances for the broker
2. Leave orphaned service instances in the catalog
3. Fail if service instances still exist for the broker

## Catalog Publishing/Curation/Discovery

* How are services identified: name, service name/id, plan name/id?
* Who can see which services? (TODO: Include scope? Global/Namespaced)
* Who can see which service instances? (TODO: Include scope? Global/Namespaced)
* Which service instances are globally visible?
* Catalog curation: which broker-provided services are globally visible?
* Catalog curation: which namespaces can see which catalog services?
* Does my catalog list not only my service types but also my instances?
  * E.g., I have a dev cluster running on my local machine and I want to connect
    my application with a database that exists in a test cluster. How do I find
    that database instance?

## Requesting Services (Consumer)

* As a User, how do I ask for a new service from the Catalog?
* As a User, how do I bind an application to an existing Service Instance?
* How does the catalog support multiple consumers in different Kubernetes
  namespaces of the same instance of a service?

## Provisioning a Service Instance

* As a Broker operator, I want to control the number of instances of my Service,
  so that I can control resource footprint of my Service.
* As an implementer of a service provider on k8s, where may I provision
  Kubernetes resources so that I may provide the requested service.
* As an implementer of a service provider on k8s, what credentials should I use
  to provision Kubernetes resources so that I may provide the requested service.

## Binding to a Service Instance

* As a Broker operator, I want to control the number of bindings to a Service
  Instance so that I may provide limits for services (e.g. free plan with 3
  bindings). (TODO: Do we care?)
* As a user of a service instance, I want a predictable set of Kubernetes
  resources (Secrets, ConfigMap) created after binding, so that I know how to
  configure my application to use the Service Instance.
* As a service operator, I want to be able to discover what applications are
  bound to services I am responsible for, so that I may operate the service
  properly.

## Using/Consuming a Service Instance

* What is the unit of consumption of a service? Namespace? Pod? Something else?
  (brian to comment)
* As a User consuming a Service, I need to be able to understand the structure
  of the Kubernetes resources that are created when a new binding to a service
  instance is created, so that I can configure my application appropriately.
* As a User, I want to be able to understand the relationship between a Secret
  and Service Instance, so that I can properly configure my application (e.g.
  app connecting to sharded database).

### Consuming bound services

Consumers of a service provisioned through the Service Catalog should be able
to access credentials for the new Service Instance using standard Kubernetes
mechanisms.

1. A Secret maintains a 1:1 relationship with a Service Instance Binding
1. The Secret should be written into the consuming application's namespace
1. The Secret should contain enough information for the consuming application
   to successfully find, connect, and authenticate to the Service Instance
   (e.g. hostname, port, protocol, username, password, etc.)
1. The consuming application may safely assume that network connectivity to the
   Service Instance is available

Consuming applications that need specific handling of credentials or
configuration should be able to use additional Kubernetes facilities to
adapt/transform the contents of the Secret. This includes, but is not limited
to, side-car and init containers.

## Lifecycle of Service Instance

* As a Service Provider, I should be able to indicate that a Service _may_ be
  upgraded, so that I can communicate Service capabilities to end users.
* As a User of a Service, I want to be able to upgrade or downgrade that Service
  so that I may size it appropriately for my needs.
* (TODO: this should be turned into a thing): What is the update story for
  bindings to a service instance?
* (TODO: this should be turned into a thing): What is the versioning and update
  story for a service: what happens when a broker changes the services it
  provides?

## Unbinding from a Service Instance

* TODO

## Deprovisioning a Service Instance

* TODO

## Removing a Catalog Entry

* TODO

