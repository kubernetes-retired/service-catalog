---
title: Using Namespaced Broker Resources
layout: docwithnav
---

# Cluster-Scoped vs Namespace-Scoped Broker Resources

Service Catalog enables service brokers to be registered in two manners: as a 
cluster-scoped resource or as a namespace-scoped resource. As a user of service
 catalog, you might use these approaches to accomplish different goals such as 
 providing a common set of service broker resources to all users or utilizing 
 role based access (RBAC) policies to control service provisioning. This 
 document will explain some use cases for namespace-scoped resources and how to
  accomplish them using Service Catalog.

## Possible Use Cases

When using `ClusterServiceBroker` and associated `ClusterServiceClass` and 
`ClusterServicePlan` resources, service broker resources, such as classes and 
plans, are created as cluster-scoped resources. This means that you are limited
 in how you can apply RBAC and you can only have a single instance of that 
 resource for a given identifier. As an example, if the service broker you are 
 registering has fixed class and plan identifiers, you will be limited to one 
 instance of the broker. With namespace-scoped brokers, however, the 
 `ServiceBroker`, along with the `ServiceClass` and `ServicePlan` resources are
  scoped to a particular namespace. This allows for some more advanced use 
  cases that were not possible with the cluster-scoped broker resources. 

### Registering Brokers Per Namespace

A service broker that provisions services in a cloud provider usually needs 
credentials in order to complete the request on behalf of Service Catalog. 
Some organizations may provide different access credentials to different teams 
in order to separate billing usage or to isolate control of resources. In these 
cases, the cluster operator might want to register two copies of the broker 
using different credentials for each team. When using `ClusterServiceBroker` 
and the associated `ClusterServiceClass` and `ClusterServicePlan` resources, it
 was not possible to register two instances of a service broker unless each 
 registration could provide unique identifiers for service c lasses and service
  plans.

Using namespace-scoped brokers, however, enables the broker to be installed in
 each namespace without conflicting at the class and plan level. When creating
  a service instance, you specify either the external class or plan name, or 
  provide the class or plan identifier. Service Catalog then resolves these in 
  order to determine which broker it should issue the provision command to. 
  When using namespace-scoped brokers and their associated resources, this 
  resolution occurs within the namespace. That means that users in namespace 
  `backend-team` and namespace `frontend-team` can have the their own broker 
  registrations and provision requests will be issued to the correct broker.

### Limiting Access to Plans

There are often situations when not all services and plans should be available 
to all users. A cluster administrator may wish to only provide free plans to 
certain users or restrict the ability to provision very expensive services. 
Additionally, when developers are creating new services that are exposed by 
brokers, they want to be able to iterate on those services without exposing 
them to all users in a cluster.

Service Catalog's cluster-scoped resources for brokers, services, and plans are
 not sufficient to implement access control to ensure that users have access 
 only to the service and plans that they should. For these resources, 
 application of RBAC is really centered around what is visible to them, but is 
 not enforced when a provision request is issued. For example, a `ClusterRole` 
 could be created to prohibit a given user or group to view classes and plans, 
 but cannot at provisioning time to ensure that the user cannot create the 
 resource. Namespace-scoped brokers, services and plans, however, can be 
 effectively combined with Kubernetes role based access control and Service 
 Catalog Catalog Restrictions in order to provide more granular control over 
 service instance provisioning.

## Enabling Namespace Scoped Broker Resources

Currently, namespace-scoped broker resources are an alpha-feature of Service 
Catalog behind a feature flag. To start using these resources, you will need 
to pass an argument to the API Server when you install Service Catalog:
 `--feature-gates NamespacedServiceBroker=true`.

If you are using Helm, you can use the `namespacedServiceBrokerEnabled` setting
 to control that flag:

```console
helm install svc-cat/catalog \
   --name catalog \
   --namespace catalog \
   --set namespacedServiceBrokerEnabled=true
```

## Using Namespace Scoped Broker Resources

Once Service Catalog has been installed with this feature gate enabled, you 
should see three new resource types: `ServiceBroker`, `ServiceClass`, 
and `ServicePlan`.

In order to register a `ServiceBroker` resource, you create a YAML definition 
that looks similar to a `ClusterServiceBroker`. This resource will use resource
 kind `ServiceBroker` and requires a namespace. An example might look like:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBroker
metadata:
  name: example-ns-broker
  namespace: ns-broker
spec:
  authInfo:
    basic:
      secretRef:
        name: my-service-broker-auth
        namespace: broker
  url: http://my-service-broker.broker.svc.cluster.local
```

Once this resource is created, Service Catalog will query the Service Broker 
for the list of available Services and create corresponding `ServiceClass` 
and `ServicePlan` resources. These resources might look like this:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceClass
metadata:
  creationTimestamp: 2018-07-12T13:30:01Z
  name: 25434f16-d762-41c7-bbdd-8045d7f74ca6
  namespace: ns-broker
  resourceVersion: "13"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/ns-broker/serviceclasses/25434f16-d762-41c7-bbdd-8045d7f74ca6
  uid: adfa2d9a-85d7-11e8-a4f3-2ae408f4a9e4
spec:
  bindable: true
  bindingRetrievable: false
  description: MySQL
  externalID: 25434f16-d762-41c7-bbdd-8045d7f74ca6
  externalName: mysql-5-7
  planUpdatable: false
  serviceBrokerName: example-ns-broker
```

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServicePlan
metadata:
  creationTimestamp: 2018-07-12T13:30:02Z
  name: 4c6932e8-30ec-4af9-83d2-6e27286dbab3
  namespace: ns-broker
  resourceVersion: "24"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/ns-broker/serviceplans/4c6932e8-30ec-4af9-83d2-6e27286dbab3
  uid: ae8e23ac-85d7-11e8-a4f3-2ae408f4a9e4
spec:
  description: basic plan
  externalID: 4c6932e8-30ec-4af9-83d2-6e27286dbab3
serviceBrokerName: example-ns-broker
  serviceClassRef:
    name: 25434f16-d762-41c7-bbdd-8045d7f74ca6e
```

The `ServiceInstance` resource has also been updated to allow you to use these 
resources just as you would the existing `ClusterServiceBroker`, 
`ClusterServiceClass` and `ClusterServicePlan` resources, except you will use 
them in the context of a namespace. For example, a `ServiceInstance` YAML that
 references a `ClusterServiceClass` and a `ClusterServicePlan` resource might 
 look like this:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: example-mysql-instance
  namespace: default
spec:
  clusterServiceClassExternalName: mysql-5-7
  clusterServicePlanExternalName: basic
```

If you instead want to use the `ServiceClass` and `ServicePlan`
 namespace-scoped resources, the yaml might look like this:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: example-mysql-instance
  namespace: default
spec:
  serviceClassExternalName: mysql-5-7
  servicePlanExternalName: basic
```

For comparison, using the cluster-scoped `ClusterServiceClass` or 
`ClusterServicePlan`, the yaml would look like:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: example-mysql-instance
  namespace: default
spec:
  clusterServiceClassExternalName: mysql-5-7
  clusterServicePlanExternalName: basic
```

Instances can reference cluster-scoped `ClusterServiceClass` or 
`ClusterServicePlan` resources or to the namespace scope `ServiceClass` and 
`ServicePlan` resources in the same namespace. They cannot reference 
`ServiceClass` and `ServicePlan` resources in another namespace.

## Further Restricting Plan Access

The use of namespace-scoped resources enables you to register brokers within a
 given namespace and leverage RBAC in order to control who can 
provision services in that namespace. By default, all service classes and plans 
from that broker will be available to users of the namespace. When registering 
a broker, catalog restrictions can be specified in order to restrict what plans
 are available within a given namespace. This allows you to specify that in 
 the `developer` namespace, only plans named `basic` can be created. The YAML 
 to accomplish this might look like:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBroker
metadata:
  name: example-ns-broker
  namespace: ns-broker
spec:
  authInfo:
    basic:
      secretRef:
        name: my-service-broker-auth
        namespace: broker
  url: http://my-service-broker.broker.svc.cluster.local
  catalogRestrictions:
    servicePlan:
    - "spec.externalName==basic"
```

When you combine the two capabilities, you can effectively restrict 
provisioning of service classes or plans to very specific namespaces. 
Production grade instances, for example, could be heavily restricted to a small
 subset of users. Other namespaces could be given access to other plans.
