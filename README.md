## `service-catalog`

### Terminology

**Application**: code that will access a Service. While in Kubernetes the code
that is deployed is often called a "Service", to avoid confusion, this
document will refer to the code that accesses a service as an "application".

**Service**: running code that is made available for use by an application.
Traditionally, services are available via HTTP REST endpoints, but this is not
a requirement.

**Service Broker**: An endpoint that manages a set of services.

**Service Catalog**: An endpoint that manages 1) a set of registered Service
Brokers and 2) the list of services that are available for instantiation from
those Service Brokers.

**Service Instance**: Each request for a unique use of a Service will result in
the Service Catalog requesting a new Service Instance from the owning Service
Broker.

### Purpose and Scope

The exact purpose of the SIG will grow over time, and as such, please see
the various "version" folders under the [docs](./docs) directory for the
exact list of use-cases, features and design decisions that have been agreed to.

However, this SIG is (initially) focused on an implementation of the
Cloud Foundry Service Broker APIs. This would include the ability for Service
Brokers to register themselves with a Kubernetes environment the same way
they can with a Cloud Foundry environment. Additionally, we will be exploring
what can be done to help Service developers when they want to expose their
code via a Service Broker.

We are currently scoping our **v1** milestone [here](./docs/v1). Interested
in contributing?  Check out the [documentation](./CONTRIBUTING.md)

## Contributing

We have a google mailing list: https://groups.google.com/forum/#!forum/kubernetes-sig-service-catalog

We have a SIG Slack Channel: https://kubernetes.slack.com/archives/sig-service-catalog

We have a weekly call: https://plus.google.com/hangouts/_/google.com/k8s-sig-service-catalog   at 1pm PT on Mondays
- And an agenda doc: https://docs.google.com/document/d/10VsJjstYfnqeQKCgXGgI43kQWnWFSx8JTH7wFh8CmPA/edit
