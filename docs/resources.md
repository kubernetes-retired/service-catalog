---
title: Resources
layout: docwithnav
---

# Service Catalog Resources

The Service Catalog resource model specifies all the behaviors that 
Service Catalog supports. This document details each resource.

All of these resources are also defined in Go code at
[`pkg/apis/servicecatalog/v1beta1/types.go`](https://github.com/kubernetes-incubator/service-catalog/blob/master/pkg/apis/servicecatalog/v1beta1/types.go).


## Service Brokers

Before a Service can be used by an Application it must first be registered
with the Kubernetes platform. Since Services are managed by Service Brokers
we must first register the Service Broker by creating an instance of a
`ClusterServiceBroker` or a `ServiceBroker`. These resources are similar, 
however one is cluster-scoped and one is namespace-scoped.

### ClusterServiceBroker

If you would like to make a service broker available cluster wide, you register 
the broker using a `ClusterServiceBroker` resource. This will result in Service Class 
and Service Plan objects being created with a cluster-scope as well. 

```console
kubectl create -f broker.yaml
```

The `broker.yaml` looks similar to this:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
  metadata:
    name: broker-name
  spec:
    url: http://broker-url.com
```

### ServiceBroker

If you would like to make a service broker available to only a single namespace, you register 
the broker using a `ServiceBroker` resource. This will result in Service Class 
and Service Plan objects being created with a namespace-scope as well.

```console
kubectl create -f broker.yaml
```

The `broker.yaml` looks similar to this:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBroker
  metadata:
    name: broker-name
    namespace: default
  spec:
    url: http://broker-url.com
```

## Service Classes

After a Service Broker has been registered by creating either a `ClusterServiceBroker` or 
`ServiceBroker`, the Service Catalog  will query the Service Broker (at the `url` specified) for the list
of available Services (the catalog). Each Service will then have a corresponding
`ClusterServiceClass` or `ServiceClass` resource created.

### ClusterServiceClass

After a `ClusterServiceBroker` resource is created, each service provided by the broker will then have a corresponding
`ClusterServiceClass` resource created. These resources will also be cluster-scoped. A `ClusterServiceClass` looks similar to this example:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceClass
metadata:
  name: 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
spec:
  bindable: true
  clusterServiceBrokerName: ups-broker
  description: A user provided service
  externalID: 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
  externalName: user-provided-service
  planUpdatable: false
```

## ServiceClass

After a `ServiceBroker` resource is created, each service provided by the broker will then have a corresponding
`ServiceClass` resource created. These resources will also be namespaced-scoped. A `ServiceClass` looks similar to this example:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceClass
metadata:
  name: 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
  namespace: default
spec:
  bindable: true
  serviceBrokerName: ups-broker
  description: A user provided service
  externalID: 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
  externalName: user-provided-service
  planUpdatable: false
```

## Service Plans

Each Service Class has one or more Plans associated with it. Each
`{ClusterServiceClass, ClusterServicePlan}` or `{ServiceClass, ServicePlan}` pair is the broker's 
service that we can provision. Plans generally indicate details like cost, performance, or 
quality-of-service.

### ClusterServicePlan

For each plan of each `ClusterServiceClass`, a `ClusterServicePlan` will be created.

### ServicePlan

For each plan of each `ServiceClass`, a `ServicePlan` will be created.

## ServiceInstance

Use a `ServiceInstance` to tell the broker to provision a new service. The 
`ServiceInstance` can use either cluster-scoped or namespace-scoped Service Class 
and Service Plan resources. When using `ServiceClass` and `ServicePlan`, the `ServiceInstance`
must be in the same namespace.

Create the `ServiceInstance`:

```console
kubectl create -f instance.yaml
```

where `instance.yaml` might look like:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  namespace: example-ns
  name: test-database
spec:
  clusterServiceClassExternalName: small-db
  clusterServicePlanExternalName: free
 ```

or

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  namespace: default
  name: test-database
spec:
  serviceClassExternalName: small-db
  servicePlanExternalName: free
 ```

### Service Instance Parameters

Each `ServiceInstance` has a `paramters` field that you can add 
metadata to. Service Catalog passes this metadata directly through to the
service broker.

You can pass this metadata in two different ways (you can pass both at the
same time): 

- Including raw JSON (inline)
- Referencing a Kubernetes `Secret`

If you reference a `Secret`, you must provide the secret name and a key.
The key in the named secret must contain the JSON to pass to the broker.

This JSON is merged with the inline JSON, but it is an error for two
sets of parameters to include the same top-level JSON property name.

If you reference a `Secret` in your `ServiceInstance`, and then the secret
is updated with new parameters, Service Catalog will not update the broker with
the new parameters. 

If you want to manually trigger an update after you've updated a `Secret`,
you have to manually increment the `UpdateRequests` field in the
`ServiceInstance`.

For more information, see the documentation on [parameters](parameters.md).

## ServiceBinding

`ServiceBinding` is the final resource that will be created in most
workflows. This resource indicates that an application wants to connect
to the service that was provisioned by a `ServiceInstance`.

Create a `ServiceBinding`:

```console
kubectl create -f binding.yaml
```

where `binding.yaml` might look like:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  namespace: example-ns
  name: test-database-binding
spec:
  instanceRef:
    name: test-database
  secretName: db-secret
```

After you create the `ServiceBinding`, Service Catalog will issue a bind
request to the appropriate broker. 

When the broker responds, Service Catalog will write the credentials that it
responds with into the secret you specified in `spec.secretName`. This
secret will be in the same namespace as the `ServiceBinding`. If you leave
`spec.SecretName` blank, the secret will be the same name as `metadata.name`.

Most secrets will have credentials (username, password, etc...) and a
hostname that your application can use to connect to the provisioned
service.

After Service Catalog creates the secret, just bind your application
pods to it and start using the service.

## What's in the Secrets?

The OSB API specification does not mandate what properties might appear
in the Credentials, so make sure your application knows what to expect
in the secret. Typically, the documentation for the broker will detail
what it returns.
