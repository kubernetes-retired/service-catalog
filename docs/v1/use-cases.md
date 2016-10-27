# Service Catalog Use Cases

TODO: add glossary
TODO: scrub doc post glossary
TODO: add CF SB API as the lingua franca for blackbox services

## Catalog Publishing/Curation/Discovery

* As a user, I want to be able to register a broker with the Kubernetes service
  catalog, so that the catalog is aware of the services that broker offers
* As a user, I want to be able to update a registered broker so that the
  catalog can maintain the most recent versions of services that broker offers
* As a user, I want to be able to delete a broker from the catalog, so that I
  can keep the catalog clean of brokers I no longer want to support

* Can I export a list of service brokers and types from my service controller?
* Is there an auth story for adding service brokers?
* As a developer, working outside of the normal production cluster, I would like
  to be able to use the services available to me from the production cluster
  from my local environment without needing to establish a formal business
  relationship with each service provider.


* How are services identified: name, service name/id, plan name/id?
* Who can see which services? (TODO: Include scope? Global/Namespaced)
* Who can see which service instances? (TODO: Include scope? Global/Namespaced)
* The Service Catalog should contain Services and not Service Instances

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

There must be a way to delete brokers from the catalog. We should evaluate
whether deleting a broker should:

1. Cascade down to the service instances for the broker
2. Leave orphaned service instances in the catalog
3. Fail if service instances still exist for the broker

## Requesting Services (Consumer)

* As a User, how do I cause a new Service Instance to be created from the
  Service Catalog?
* As a User, how do I bind an application to an existing Service Instance?
* How does the catalog support multiple consumers in different Kubernetes
  namespaces of the same Service Instance?
* As a User, who has requested a Service Instance, know that a request for a
  service instance has been fulfilled?
* As a User, I should be able to pass parameters to be used by the Service
  Instance or Binding when causing a new Service Instance to be created, so that
  I may change the attributes of the Service Instance or Binding.

## Provisioning a Service Instance

* As a Broker operator, I want to control the number of instances of my Service,
  so that I can control the resource footprint of my Service.

## Binding to a Service Instance

* As a Broker operator, I want to control the number of bindings to a Service
  Instance so that I may provide limits for services (e.g. free plan with 3
  bindings). (TODO: Do we care?)
* As a user of a service instance, I want a predictable set of Kubernetes
  resources (Secrets, ConfigMap) created after binding, so that I know how to
  configure my application to use the Service Instance.
* As a catalog operator, I want to be able to discover what applications are
  bound to services I am responsible for, so that I may operate the service
  properly.
* As a User I should be able to see what service instances my applications are
  bound to.
* As a User I should be able to pass paramters when binding to a service
  instance so that I may indicate what type of binding my application needs.
  (e.g. credential type, admin binding, ro binding, rw binding)

As a User, I should be able to accomplish the following sets of bindings:

* One application may binding to many Service Instances
* Many different applications may bind to a single Service Instance
  * ...with unique credentials
  * ...with identical credentials
* One application, binding multiple times to the same Service Instance

## Using/Consuming a Service Instance

* As a User consuming a Service Instance, I need to be able to understand the structure
  of the Kubernetes resources that are created when a new binding to a service
  instance is created, so that I can configure my application appropriately.
* As a User, I want to be able to understand the relationship between a Secret
  and Service Instance, so that I can properly configure my application (e.g.
  app connecting to sharded database).
* The consuming application may safely assume that network connectivity to the
  Service Instance is available

Consuming applications that need specific handling of credentials or
configuration should be able to use additional Kubernetes facilities to
adapt/transform the contents of the the credentials/configuration. This
includes, but is not limited to, side-car and init containers.

If the user were willing to change the application, then we could
drop the credentials in some "standard" place by convention. This would be
similar to how K8s service accounts work (service-account secrets are mounted at
`/var/run/secrets/kubernetes.io/serviceaccount`), as well as `VCAP_SERVICES` in
CF.

If the user were willing to change the configuration instead, they could specify
how the credentials were surfaced to the application -- which environment
variables, volumes, etc.

## Lifecycle of Service Instances

* As a Service Provider, I should be able to indicate that a Service Instance
  _may_ be upgraded (plan updateable), so that I can communicate Service
  capabilities to end users.
* As a User of a Service Instance, I want to be able to change the Service
  Instance plan so that I may size it appropriately for my needs.
* What is the update story for bindings to a service instance?
* What is the versioning and update story for a service instance: what happens
  when a broker changes the services it provides?

## Unbinding from a Service Instance

* TODO

## Deprovisioning a Service Instance

* TODO
