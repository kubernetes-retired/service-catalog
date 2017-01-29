## `service-catalog`

[![Build Status](https://travis-ci.org/kubernetes-incubator/service-catalog.svg?branch=master)](https://travis-ci.org/kubernetes-incubator/service-catalog)

### Introduction

The service-catalog project is in incubation to bring integration with service
brokers to the Kubernetes ecosystem via the [Open Service Broker
API](https://github.com/openservicebrokerapi/servicebroker). A service broker
is an endpoint that manages a set of services.  The end-goal of the service-
catalog project is to provide a way for Kubernetes users to consume services
from brokers and easily configure their applications to use those services,
without needing detailed knowledge about how those services are created /
managed.

As an example:

Say that a distributed application needs a message broker that hosts a set of
queues/topics.  The service-catalog allows Kubernetes applications to consume
services like message brokers that exists _somewhere_ in a simple way:

1.  A user wanting to consume a service in their application browses a list of
    available services in the catalog
2.  The user asks for a new instance of that service to be _provisioned_

     _Provisioning_ means that the broker somehow creates a new instance of a
    service.  This could mean basically anything that results in a new instance
    of the service becoming available.  Possibilities include: creating a new
    set of Kubernetes resources in another namespace in the same Kubernetes
    cluster as the consumer or a different cluster, or even creating a new
    tenant in a multi-tenant SaaS system.  The point is that the
    consumer doesn't have to be aware of or care at all about the details.
3.  The user _binds_ that service to their application

    _Binding_ means that the application is injected with the information
    necessary to use the service, such as coordinates, credentials, and
    configuration items.  Applications are injected using the existing
    application configuration primitives in Kubernetes: Services, Secrets, and
    ConfigMaps.

---

We are currently scoping our **v1** milestone [here](./docs/v1).

### Terminology

This project's problem domain contains a few inconvenient but unavoidable
overloads with other Kubernetes terms.  Check out our [terminology
page](./terminology.md) for definitions of terms as they are used in this
project.

### Contributing

Interested in contributing?  Check out the [documentation](./CONTRIBUTING.md).

Also see our [developer's guide](./docs/DEVGUIDE.md) for information on how to
build and test the code.

### Kubernetes Incubator

This is a [Kubernetes Incubator project](https://github.com/kubernetes/community/blob/master/incubator.md).
The project was established 2016-Sept-12.  The incubator team for the project is:

- Sponsor: Brian Grant ([@bgrant0607](https://github.com/bgrant0607))
- Champion: Paul Morie ([@pmorie](https://github.com/pmorie))
- SIG: [sig-service-catalog](https://github.com/kubernetes/community/tree/master/sig-service-catalog)

For more information about sig-service-catalog such as meeting times and agenda,
check out the [community site](https://github.com/kubernetes/community/tree/master/sig-service-catalog).

There is a YouTube channel for our recorded meetings
[here](https://www.youtube.com/playlist?list=PL69nYSiGNLP2k9ZXx9E1MvRSotFDoHUWs&jct=kNbyYC551SM9HbO2eMTEIvDj1wxQfA).

### Code of Conduct

Participation in the Kubernetes community is governed by the
[Kubernetes Code of Conduct](./code-of-conduct.md).
