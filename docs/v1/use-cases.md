## V1 Use Cases

This document contains the complete list of accepted use-cases for the v1
version of `service-catalog`.

## Use cases

1.  Sharing services:
    1.  (Blackbox services) As a SaaS provider that already runs a Service
        Broker, I want users of Kubernetes to be able to use my service
        via the Service Broker API, so that I can grow my user base to
        include users of Kubernetes

### Blackbox services

There are numerous SaaS providers that already operate Service Brokers today.
It should be possible for the operator of an existing Service Broker to
publish their services into the catalog and have them consumed by users of
Kubernetes.  This offers a new set of users to the service operator and offers
a wide variety of SaaS products to users of Kubernetes.
