## `service-catalog`

### Terminology

Application: code that will access a Service. While in Kubernetes the code that
is deployed is often called a "Service", to avoid confusion, this document will
refer to the code that accesses a service as an "application".

Recipe: the definition of an application or service in a form that can then be
deployed into a cloud provider. For example, in Kubernetes terms, a recipe
might be a [Deployment](http://kubernetes.io/docs/user-guide/deployments/)
yaml file.

Service: running code that is made available for use by an application.
Traditionally, services are available via HTTP REST endpoints, but this is not
a requirement.

Service Broker: An endpoint that manages a set of services.

Service Catalog: An endpoint that manages 1) a set of registered Service
Brokers and 2) the list of services that are available for instantiation from
those Service Brokers.

Service Instance: Each request for a unique use of a Service will result in the
Service Catalog requesting a new Service Instance from the owning Service
Broker.

### Purpose and Scope

The service-catalog SIG will focus on a few key tasks:

#### Service Catalog:
1. Support a Service Catalog similar to what Cloud Foundry supports via
   their Cloud Controller APIs. This Service Catalog will support the
   notion of managing registered Service Brokers, which in turn dictates the
   list of available services that can be used by applications running within
   Kubernetes. The Service Catalog may be cluster-wide or scoped to something smaller, such
   as a Namespace.
2. Provide a mechanism by which applications deployed into Kubernetes can
   signal their desire to connect to an instance of a Service Catalog managed
   service. And, to have the information necessary for the application to
   use the service instance provided to the application.

#### Service Brokers
1. Provide a framework to aide in the development of a Kubernetes-hosted 
   Service Broker that can then be registered into a Service Catalog - 
   either a Kubernetes Service Catalog or a Cloud Foundry one.
2. Provide a mechanism by which a Kubernetes managed Service Broker can be
   told which Namespace into which a new service instance should be created.
   This is optional because not all services necessarily deploy new running
   code for each service instance. For example, a DB service may simply
   create a new set of DB tables per instance.

#### Service Recipes
1. Provide a Deployment Catalog that includes a list of pre-built deployable
   resources that can then be instantiated into a Kubernetes environment.
   Note that these may or may not be "Services" as defined above.
2. These deployments, if they support the Service Broker APIs, can then be
   registered into the Service Broker for use by applications on the
   platform.

This is the Kubernetes implementation of the service broker concept, which is
joint effort amongst the different member organizations of the
[CNCF](https://cncf.io/).

Interested in contributing?  Check out the [documentation](./CONTRIBUTING.md)

### Use Cases

A very basic set of use cases to describe the problem space is:

1.  Advertising and discovering of recipes:
    1.  As a recipte provider, I want to be able to publish recipes
        so that users can search and deploy them into their own environments.
    2.  As someone who wants to run a particular software system in Kubernetes,
        I want to be able to search for and use recipes that others may have
        already created, so I can avoid spending time getting it to run myself.
2.  Sharing resources for a service:
    1.  As an operator of a software system, I want to share the resources that
        are required to use the system so that my users can easily consume
        them in their own namespaces
    2.  As a user of a software system running in Kubernetes, I want to consume
        the shared resources associated with that system in my own namespace so
        that I can use the system in my application

#### Advertising services and recipes

Within and outside a Kubernetes cluster, there are services that users wish to
highlight and make available to other users.  Users might also wish to publish
recipes that allow other users to run their own services.  Some examples:

1.  A user's namespace contains `etcd`, `etcd-discovery`, and `postgresql`
    services, and the only one the user wants to share with others is the
    `postgresql` service
2.  A SaaS product like a externally hosted database for which a Kubernetes
    Service exists to provide a stable network endpoint
3.  A user makes a database run in Kubernetes and wants to share their recipe

In order to share these services, there has to be a central place where they can
be registered and advertised.  This is the service catalog.

#### Sharing recipes

Users also want the ability to share recipes for running services in addition to
sharing access to services that are already running.  As a completely fictitious
example, say the a user creates some kind of recipe that makes it easy to create
everything needed to spin up a new PostgreSQL database (customizable
username/password, `Service`, `Deployment`, etc.). The user wants to share this
recipe in a service catalog so others can find it and use it.

#### Consuming recipes

When a user consumes a recipe, the pieces of the recipe are fully realized in
that user's namespace.  For example, if the recipe is to run an instance of
PostgreSQL, the user's namespace would probably have several new resources
created in it:

1.  A `Deployment` for the actual PostgreSQL containers
2.  A `Service` to provide a stable network endpoint
3.  A `Secret` with credentials to use the database

#### Sharing a single set of resources for a service

The simplest way to share resources for an existing service is to share the same
resources for each consumer.  As an example: a development team is working on an
application that uses a database. The IT department manages the database (i.e.,
it lives off-cluster). All developers share the same credentials to access the
database, but these credentials are managed by IT. Rather than having each
developer create his or her own `Service` and `Secret` to connect to the
database, IT creates a "db-app-xyz" `Service` and a "db-app-xyz" `Secret` in the
"info-tech" namespace.

#### Consuming a set of shared resources for a service

Continuing our shared database example from a developer perspective: to use the
shared database service, a developer searches for it in the service catalog and
adds it to their namespace.  When the developer adds the service from the
catalog into their own namespace, they receive a copy of each of the resources
(Secrets, ConfigMaps, etc) that the service publisher has associated with that
service in their namespace.

## Contributing

We have a google mailing list: https://groups.google.com/forum/#!forum/kubernetes-sig-service-catalog

We have a SIG Slack Channel: https://kubernetes.slack.com/archives/sig-service-catalog

We have a weekly call: https://plus.google.com/hangouts/_/google.com/k8s-sig-service-catalog   at 1pm PT on Mondays
- And an agenda doc: https://docs.google.com/document/d/10VsJjstYfnqeQKCgXGgI43kQWnWFSx8JTH7wFh8CmPA/edit
