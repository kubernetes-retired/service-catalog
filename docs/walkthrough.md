---
title: Walkthrough
layout: docwithnav
---

This document assumes that you've installed Service Catalog onto your cluster.
If you haven't, please see the [installation instructions](./install.md). Optionally you may install
the Service Catalog CLI, svcat. Examples for both svcat and kubectl are provided
so that you may follow this walkthrough using svcat or using only kubectl.

All commands in this document assume that you're operating out of the root
of this repository.

<a id="install" />

# Step 1 - Installing the UPS Broker Server

Since the Service Catalog provides a Kubernetes-native interface to an
[Open Service Broker API](https://www.openservicebrokerapi.org/) compatible broker
server, we'll need to install one in order to proceed with a demo.

In this repository, there's a simple, "dummy" server called the User Provided
Service (UPS) broker. The codebase for that broker is
[here](https://github.com/kubernetes-incubator/service-catalog/tree/master/contrib/pkg/broker/user_provided/controller).

We're going to deploy the UPS broker to our Kubernetes cluster before
proceeding, and we'll do so with the UPS helm chart. You can find details about
that chart in the chart's
[README](https://github.com/kubernetes-incubator/service-catalog/blob/master/charts/ups-broker/README.md).

Otherwise, to install with sensible defaults, run the following command:

```console
helm install ./charts/ups-broker --name ups-broker --namespace ups-broker
```
**NOTE:** The walkthrough installs a [cluster-wide UPS Broker](https://github.com/kubernetes-incubator/service-catalog/tree/master/contrib/examples/walkthrough/ups-clusterservicebroker.yaml). For a namespace-scoped service broker, see [this](https://github.com/kubernetes-incubator/service-catalog/tree/master/contrib/examples/walkthrough/ups-servicebroker.yaml) file.

# Step 2 - Creating a ClusterServiceBroker Resource

Because we haven't created any resources in the service-catalog API server yet,
querying service catalog returns an empty list of resources:

```console
$ svcat get brokers
  NAME   URL   STATUS
+------+-----+--------+

$ kubectl get clusterservicebrokers,clusterserviceclasses,serviceinstances,servicebindings
No resources found.
```

We'll register a broker server with the catalog by creating a new
[`ClusterServiceBroker`](../contrib/examples/walkthrough/ups-clusterservicebroker.yaml) resource:

```console
$ kubectl create -f contrib/examples/walkthrough/ups-clusterservicebroker.yaml
clusterservicebroker.servicecatalog.k8s.io/ups-broker created
```

When we create this `ClusterServiceBroker` resource, the service catalog controller responds
by querying the broker server to see what services it offers and creates a
`ClusterServiceClass` for each.

We can check the status of the broker:

```console
$ svcat describe broker ups-broker
  Name:     ups-broker
  URL:      http://ups-broker-ups-broker.ups-broker.svc.cluster.local
  Status:   Ready - Successfully fetched catalog entries from broker @ 2018-10-09 08:25:25 +0000 UTC

$ kubectl get clusterservicebrokers ups-broker -o yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  creationTimestamp: 2018-10-09T08:25:25Z
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  name: ups-broker
  resourceVersion: "10"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers/ups-broker
  uid: deefbd1e-cb9c-11e8-8372-fade7e9a18e5
spec:
  relistBehavior: Duration
  relistRequests: 0
  url: http://ups-broker-ups-broker.ups-broker.svc.cluster.local
status:
  conditions:
  - lastTransitionTime: 2018-10-09T08:25:25Z
    message: Successfully fetched catalog entries from broker.
    reason: FetchedCatalog
    status: "True"
    type: Ready
  lastCatalogRetrievalTime: 2018-10-09T08:25:25Z
  reconciledGeneration: 1
```

Notice that the status reflects that the broker's
catalog of service offerings has been successfully added to our cluster's
service catalog.

# Step 3 - Viewing ClusterServiceClasses and ClusterServicePlans

The controller created a `ClusterServiceClass` for each service that the UPS broker
provides. We can view the `ClusterServiceClass` resources available:

```console
$ svcat get classes
                 NAME                  NAMESPACE         DESCRIPTION
+------------------------------------+-----------+-------------------------+
  user-provided-service                            A user provided service
  user-provided-service-single-plan                A user provided service
  user-provided-service-with-schemas               A user provided service

$ kubectl get clusterserviceclasses
NAME                                   EXTERNAL NAME
4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468   user-provided-service
5f6e6cf6-ffdd-425f-a2c7-3c9258ad2468   user-provided-service-single-plan
8a6229d4-239e-4790-ba1f-8367004d0473   user-provided-service-with-schemas
```

**NOTE:** The above kubectl command uses a custom set of columns.  The `NAME` field is
the Kubernetes name of the `ClusterServiceClass` and the `EXTERNAL NAME` field is the
human-readable name for the service that the broker returns.

The UPS broker provides a service with the external name
`user-provided-service`. View the details of this offering:

```console
$ svcat describe class user-provided-service
  Name:          user-provided-service
  Description:   A user provided service
  UUID:          4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
  Status:        Active
  Tags:
  Broker:        ups-broker

Plans:
   NAME           DESCRIPTION
+---------+-------------------------+
  default   Sample plan description
  premium   Premium plan

$ kubectl get clusterserviceclasses 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468 -o yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceClass
metadata:
  creationTimestamp: 2018-10-09T08:25:25Z
  name: 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
  ownerReferences:
  - apiVersion: servicecatalog.k8s.io/v1beta1
    blockOwnerDeletion: false
    controller: true
    kind: ClusterServiceBroker
    name: ups-broker
    uid: deefbd1e-cb9c-11e8-8372-fade7e9a18e5
  resourceVersion: "3"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterserviceclasses/4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
  uid: deff758e-cb9c-11e8-8372-fade7e9a18e5
spec:
  bindable: true
  bindingRetrievable: false
  clusterServiceBrokerName: ups-broker
  description: A user provided service
  externalID: 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
  externalName: user-provided-service
  planUpdatable: true
status:
  removedFromBrokerCatalog: false
```

Additionally, the controller created a `ClusterServicePlan` for each of the
plans for the broker's services. We can view the `ClusterServicePlan`
resources available in the cluster:

```console
$ svcat get plans
   NAME     NAMESPACE                 CLASS                           DESCRIPTION
+---------+-----------+------------------------------------+--------------------------------+
  default               user-provided-service                Sample plan description
  premium               user-provided-service                Premium plan
  default               user-provided-service-single-plan    Sample plan description
  default               user-provided-service-with-schemas   Plan with parameter and
                                                             response schemas

$ kubectl get clusterserviceplans
NAME                                   EXTERNAL NAME
4dbcd97c-c9d2-4c6b-9503-4401a789b558   default
86064792-7ea2-467b-af93-ac9694d96d52   default
96064792-7ea2-467b-af93-ac9694d96d52   default
cc0d7529-18e8-416d-8946-6f7456acd589   premium
```

You can view the details of a `ClusterServicePlan` with this command:

```console
$ svcat describe plan user-provided-service/default
  Name:          default
  Description:   Sample plan description
  UUID:          86064792-7ea2-467b-af93-ac9694d96d52
  Status:        Active
  Free:          true
  Class:         user-provided-service

Instances:
No instances defined

$ kubectl get clusterserviceplans 86064792-7ea2-467b-af93-ac9694d96d52 -o yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServicePlan
metadata:
  creationTimestamp: 2018-10-09T08:25:25Z
  name: 86064792-7ea2-467b-af93-ac9694d96d52
  ownerReferences:
  - apiVersion: servicecatalog.k8s.io/v1beta1
    blockOwnerDeletion: false
    controller: true
    kind: ClusterServiceBroker
    name: ups-broker
    uid: deefbd1e-cb9c-11e8-8372-fade7e9a18e5
  resourceVersion: "6"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterserviceplans/86064792-7ea2-467b-af93-ac9694d96d52
  uid: df0b1a4c-cb9c-11e8-8372-fade7e9a18e5
spec:
  clusterServiceBrokerName: ups-broker
  clusterServiceClassRef:
    name: 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
  description: Sample plan description
  externalID: 86064792-7ea2-467b-af93-ac9694d96d52
  externalName: default
  free: true
status:
  removedFromBrokerCatalog: false
```

# Step 4 - Creating a New ServiceInstance

Now that a `ClusterServiceClass` named `user-provided-service` exists within our
cluster's service catalog, we can create a `ServiceInstance` that points to
it.

Unlike `ClusterServiceBroker` and `ClusterServiceClass` resources, `ServiceInstance`
resources must be namespaced. Create a namespace with the following command:

```console
$ kubectl create namespace test-ns
namespace "test-ns" created
```

Then, create the `ServiceInstance`:

```console
$ kubectl create -f contrib/examples/walkthrough/ups-instance.yaml
serviceinstance.servicecatalog.k8s.io/ups-instance created
```

After the `ServiceInstance` is created, the service catalog controller will
communicate with the appropriate broker server to initiate provisioning.
Check the status of that process:

```console
$ svcat describe instance -n test-ns ups-instance
  Name:        ups-instance
  Namespace:   test-ns
  Status:      Ready - The instance was provisioned successfully @ 2018-10-09 08:30:39 +0000 UTC
  Class:       user-provided-service
  Plan:        default

Parameters:
  param-1: value-1
  param-2: value-2

Bindings:
No bindings defined

$ kubectl get serviceinstances -n test-ns ups-instance -o yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  creationTimestamp: 2018-10-09T08:30:39Z
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  name: ups-instance
  namespace: test-ns
  resourceVersion: "15"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/test-ns/serviceinstances/ups-instance
  uid: 9a40ca6f-cb9d-11e8-8372-fade7e9a18e5
spec:
  clusterServiceClassExternalName: user-provided-service
  clusterServiceClassRef:
    name: 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
  clusterServicePlanExternalName: default
  clusterServicePlanRef:
    name: 86064792-7ea2-467b-af93-ac9694d96d52
  externalID: 9a40c9fb-cb9d-11e8-8372-fade7e9a18e5
  parameters:
    param-1: value-1
    param-2: value-2
  updateRequests: 0
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: docker-for-desktop
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: 2018-10-09T08:30:39Z
    message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
  deprovisionStatus: Required
  externalProperties:
    clusterServicePlanExternalID: 86064792-7ea2-467b-af93-ac9694d96d52
    clusterServicePlanExternalName: default
    parameterChecksum: 4fa544b50ca7a33fe5e8bc0780f1f36aa0c2c7098242db27bc8a3e21f4b4ab55
    parameters:
      param-1: value-1
      param-2: value-2
    userInfo:
      groups:
      - system:masters
      - system:authenticated
      uid: ""
      username: docker-for-desktop
  observedGeneration: 1
  orphanMitigationInProgress: false
  provisionStatus: Provisioned
  reconciledGeneration: 1
```

# Step 5 - Requesting a ServiceBinding to use the ServiceInstance

Now that our `ServiceInstance` has been created, we can bind to it.
Create a `ServiceBinding` resource:

```console
$ kubectl create -f contrib/examples/walkthrough/ups-binding.yaml
servicebinding.servicecatalog.k8s.io/ups-binding created
```

After the `ServiceBinding` resource is created, the service catalog controller will
communicate with the appropriate broker server to initiate binding. Generally,
this will cause the broker server to create and issue credentials that the
service catalog controller will insert into a Kubernetes `Secret`. We can check
the status of this process like so:

```console
$ svcat describe binding -n test-ns ups-binding
  Name:        ups-binding
  Namespace:   test-ns
  Status:      Ready - Injected bind result @ 2018-10-09 08:31:38 +0000 UTC
  Secret:      ups-binding
  Instance:    ups-instance

Parameters:
  No parameters defined

Secret Data:
  special-key-1   15 bytes
  special-key-2   15 bytes

$ kubectl get servicebindings -n test-ns ups-binding -o yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  creationTimestamp: 2018-10-09T08:31:38Z
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  name: ups-binding
  namespace: test-ns
  resourceVersion: "18"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/test-ns/servicebindings/ups-binding
  uid: bd7f81ed-cb9d-11e8-8372-fade7e9a18e5
spec:
  externalID: bd7f81b9-cb9d-11e8-8372-fade7e9a18e5
  instanceRef:
    name: ups-instance
  secretName: ups-binding
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: docker-for-desktop
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: 2018-10-09T08:31:38Z
    message: Injected bind result
    reason: InjectedBindResult
    status: "True"
    type: Ready
  externalProperties:
    userInfo:
      groups:
      - system:masters
      - system:authenticated
      uid: ""
      username: docker-for-desktop
  orphanMitigationInProgress: false
  reconciledGeneration: 1
  unbindStatus: Required
```

Notice that the status has a `Ready` condition set.  This means our binding is
ready to use!  If we look at the `Secret`s in our `test-ns` namespace, we should
see a new one:

```console
$ kubectl get secrets -n test-ns
NAME                  TYPE                                  DATA   AGE
default-token-v24x9   kubernetes.io/service-account-token   3      1m
ups-binding           Opaque                                2      37s
```

Notice that a new `Secret` named `ups-binding` has been created.

# Step 6 - Deleting the ServiceBinding

Now, let's unbind the instance:

```console
$ svcat unbind -n test-ns ups-instance
deleted ups-binding
```

After the deletion is complete, we should see that the `Secret` is gone:

```console
$ kubectl get secrets -n test-ns
NAME                  TYPE                                  DATA   AGE
default-token-v24x9   kubernetes.io/service-account-token   3      2m
```

# Step 7 - Deleting the ServiceInstance

Now, we can deprovision the instance:

```console
$ svcat deprovision -n test-ns ups-instance
deleted ups-instance
```

# Step 8 - Deleting the ClusterServiceBroker

Next, we should remove the `ClusterServiceBroker` resource. This tells the service
catalog to remove the broker's services from the catalog. Do so with this
command:

```console
$ kubectl delete clusterservicebrokers ups-broker
clusterservicebroker.servicecatalog.k8s.io "ups-broker" deleted
```

We should then see that all the `ClusterServiceClass` resources that came from that
broker have also been deleted:

```console
$ svcat get classes
  NAME   NAMESPACE   DESCRIPTION
+------+-----------+-------------+

$ kubectl get clusterserviceclasses
No resources found.
```

# Step 9 - Final Cleanup

## Cleaning up the User Provided Service Broker

To clean up, delete the helm deployment:

```console
helm delete --purge ups-broker
```

Then, delete all the namespaces we created:

```console
kubectl delete ns test-ns ups-broker
```

## Cleaning up the Service Catalog

Delete the helm deployment and the namespace:

```console
helm delete --purge catalog
kubectl delete ns catalog
```

# Troubleshooting

## Firewall rules

If you are using Google Cloud Platform, you may need to run the following
command to setup proper firewall rules to allow your traffic get in.

```console
gcloud compute firewall-rules create allow-service-catalog-secure --allow tcp:30443 --description "Allow incoming traffic on 30443 port."
```
