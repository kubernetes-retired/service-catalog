## `service-catalog`

[![Build Status](https://travis-ci.org/kubernetes-incubator/service-catalog.svg?branch=master)](https://travis-ci.org/kubernetes-incubator/service-catalog)

### Introduction

The service-catalog project is in incubation to bring integration with service
brokers to the Kubernetes ecosystem via the
[Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker).
A service broker is an endpoint that manages a set of services.  The end-goal of
the service-catalog project is to provide a way for Kubernetes users to consume
services from brokers and easily configure their applications to use those
services.

---

We are currently scoping our **v1** milestone [here](./docs/v1).

### Terminology

This project's problem domain contains a few inconvenient but unavoidable
overloads with other Kubernetes terms.  Check out our [terminology
page](./terminology.md) for definitions of terms as they are used in this
project.

### Contributing

Interested in contributing?  Check out the [documentation](./CONTRIBUTING.md).

### Kubernetes Incubator

This is a [Kubernetes Incubator project](https://github.com/kubernetes/community/blob/master/incubator.md).
The project was established 2016-Sept-12.  The incubator team for the project is:

- Sponsor: Brian Grant ([@bgrant0607](https://github.com/bgrant0607))
- Champion: Paul Morie ([@pmorie](https://github.com/pmorie))
- SIG: [sig-service-catalog](https://github.com/kubernetes/community/tree/master/sig-service-catalog)

For more information about sig-service-catalog such as meeting times and agenda,
check out the [community site](https://github.com/kubernetes/community/tree/master/sig-service-catalog).

### Code of Conduct

Participation in the Kubernetes community is governed by the
[Kubernetes Code of Conduct](./code-of-conduct.md).

