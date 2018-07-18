---
title: Filtering Broker Catalogs
layout: docwithnav
---

# Catalog Restrictions

Services provided by service brokers are represented in Kubernetes by two 
different [resources](resources.md), service classes and service plans. When a 
`ClusterServiceBroker` or `ServiceBroker` resource is created, the Service 
Catalog will query the Service Broker for the list of available Services. 
Service Catalog will then create `ClusterServiceClass` or `ServiceClass` 
resources to represent the service classes and `ClusterServicePlan` or 
`ServicePlan` resources to represent service plans. By default, Service Catalog
will create a `ClusterServiceClass` or `ServiceClass` for each service class 
and a `ClusterServicePlan` or `ServicePlan` for each service plan 
provided by the service broker. When creating a `ClusterServiceBroker` or 
`ServiceBroker` resource, you can change this behavior by specifying one or 
more catalog restrictions. Catalog restrictions act in a manner similar to 
Kubernetes label selectors to enable you to control how service classes and 
service plans should be exposed from the service brokers.

## Using Catalog Restrictions

Catalog restrictions are specified in `ClusterServiceBroker` or `ServiceBroker`
 resources. A sample YAML might look like:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  name: sample-broker
spec:
  authInfo:
    basic:
      secretRef:
        name: sample-broker-auth
        namespace: brokers
  catalogRestrictions:
    servicePlan:
    - "spec.externalName==basic"
  url: http://sample-broker.brokers.svc.cluster.local
```

In this example, a catalog restriction has been defined that specifies that 
only service plans that have an external name of basic should be selected. 
Catalog restrictions are defined as a set of one or more rules that target 
 service classes and/or service plans. These rules have a special format
 similar to Kubernetes label selectors. 

The rule format is expected to be `<property><conditional><requirement>`

* `<property>` is one of the supported properties of a service class or service plan resource, described below
* `<conditional>` is allowed to be one of the following: `==`, `!=`, `in`, `notin`
* `<requirement>` will be a string value if `==` or `!=` are used, otherwise it will be a set of string values if `in` or `notin` are used
* `<requirement>` is case sensitive

Catalog restrictions, while similar to label selectors, only operate on a 
subset of properties on service class and service plan resources. The following
 sections detail what properties can be used to define catalog restrictions for
  each resource type.  

`ClusterServiceClass` allowed property names:

| Property Key    | Description    |
| name |  This key will match the ClusterServiceClass.Name property |
| spec.externalName | This key will match the ClusterServiceClass.Spec.ExternalName property |
| spec.externalID | This key will match the ClusterServiceClass.Spec.ExternalID property |

`ServiceClass` allowed property names:

| Property Key    | Description    |
| name |  This key will match the ServiceClass.Name |
| spec.externalName | This key will match the ServiceClass.Spec.ExternalName property |
| spec.externalID | This key will match the ServiceClass.Spec.ExternalID property |

`ClusterServicePlan` allowed property names:

| Property Key    | Description    |
| name | This key will match the ClusterServicePlan.Name |
| spec.externalName | This key will match the ClusterServicePlan.Spec.ExternalName property |
| spec.externalID | This key will match the ClusterServicePlan.Spec.ExternalID property |
| spec.free | This key will match the ClusterServicePlan.Spec.Free property |
| spec.clusterServiceClass.name | This key will match the ClusterServicePlan.Spec.ClusterServiceClassRef.Name property |

`ServicePlan` allowed property names:

| Property Key    | Description    |
| name | This key will match the ServicePlan.Name property |
| spec.externalName | This key will match the ServicePlan.Spec.ExternalName property |
| spec.externalID | This key will match the ServicePlan.Spec.ExternalID property |
| spec.free | This key will match the ServicePlan.Spec.Free property |
| spec.serviceClass.name | This key will match the ServicePlan.Spec.ServiceClassRef.Name property |

## Examples

The following examples show some possible ways to apply catalog restrictions.

### Allow Only Service Class Resources with Specific External Name

This example creates a Service Class restriction on spec.externalName using the
 `in` operator. In this case, only services that have the externalName 
 `FooService` or `BarService` will have Service Catalog resources created. 
 The YAML for this would look like:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  name: sample-broker
spec:
  authInfo:
    basic:
      secretRef:
        name: sample-broker-auth
        namespace: brokers
  catalogRestrictions:
    serviceClass:
    - "spec.externalName in (FooService, BarService)"
  url: http://sample-broker.brokers.svc.cluster.local
```

### Allow All Service Class Resources Except Those with Specific External Name

 To allow all services, except those named `FooService` or `BarService`, 
 the `notin` operator can be used. The YAML for this would look like:
 above.

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  name: sample-broker
spec:
  authInfo:
    basic:
      secretRef:
        name: sample-broker-auth
        namespace: brokers
  catalogRestrictions:
    serviceClass:
    - "spec.externalName notin (FooService, BarService)"
  url: http://sample-broker.brokers.svc.cluster.local
```

### Using Multiple Predicates

As mentioned above, you can chain rules together. For example,
to restrict service plans to only those free plans with an externalName of 
`Demo`, the YAML would look like:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  name: sample-broker
spec:
  authInfo:
    basic:
      secretRef:
        name: sample-broker-auth
        namespace: brokers
  catalogRestrictions:
    servicePlan:
    - "spec.externalName in (Demo)"
    - "spec.free=true"
  url: http://sample-broker.brokers.svc.cluster.local
```

### Combining Service Class and Service Plan Catalog Restrictions

You can also combine restrictions on classes and plans. An example that 
allow all free plans with the externalName `Demo`, and not a specific service
 named `AABBB-CCDD-EEGG-HIJK`, you would create a YAML like:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  name: sample-broker
spec:
  authInfo:
    basic:
      secretRef:
        name: sample-broker-auth
        namespace: brokers
  catalogRestrictions:
    serviceClass:
    - "name!=AABBB-CCDD-EEGG-HIJK"
    servicePlan:
    - "spec.externalName in (Demo)"
    - "spec.free=true"
  url: http://sample-broker.brokers.svc.cluster.local
```