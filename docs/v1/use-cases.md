## V1 Use Cases

This document contains the complete list of accepted use-cases for the v1
version of the service catalog.

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
        
2. Searching and Browsing Services:
    1.  As a consumer, I'm able to search my catalog for services by attributes
        such as category.
    2.  As a consumer, I'm able to see metadata about a service prior to 
        creation which allows me to see if this service fits my need.
    3.  As a consumer, I'm able to see all the required and optional parameters
        I'll need to pass to the service in order to create it.
    4.  As a consumer, when listing services I see a union of the catalogs
        from all brokers I have registered. However, if I want to restrict the
        list to a specific broker I can pass that in as a flag.
   

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

### Search and Browsing Services

#### Searching Services

Consumers should be be able to search or filter their catalog by labels. For
example, if I search for all services with 'catalog=database' the catalog
will return the list of services that match that label. This assumes, of
course, that producers are able to label their service offerings.

#### Service Metadata

Each service should have a list of metadata that it exposes in the catalog.
If we're following the Cloud Foundry model you can view the list of metadata
fields [here](https://docs.cloudfoundry.org/services/catalog-metadata.html).

We should consider what metadata needs to be exposed for a strong CLI and UI
experience. Here are some suggestions for metadata fields:

    * name
    * short description
    * long desciption
    * documentation/support urls
    * icon URL
    * image URLs - a list of images that could be displayed in a UI
    * TOS link
    * a list of plans
        * plan name
        * plan description
        * plan cost
    * construction parameters
        * name
        * description
        * default value
    * category label/tags
    * version
    * publisher name
    * publisher contact url
    * publisher website

#### Viewing Service Parameters

Each service offering may have a list of parameters (e.g., configuration) required 
to create that service. For example, if consuming a hosted database, I may need to 
specify the region, size, a link to a startup scripts, or other parameters. 

For each service, I'm able to see the list of required and optional parameters that
I can pass in during service creation. Service producers are able to specify default
values for these parameters.

#### Listing Services

When listing all services available to be created, users will see a union
of the catalog offerings from all brokers registered. However, users have the
option of passing in a flag to limit results to just a specific registered broker.

TODO: How to deal with name conflicts for {broker, service}.


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
