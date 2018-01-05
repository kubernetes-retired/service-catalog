# Service Catalog Demonstration Walkthrough

This document assumes that you've installed Service Catalog and the Service Catalog CLI, svcat, onto your cluster.
If you haven't, please see [install.md](./install.md).

All commands in this document assume that you're operating out of the root
of this repository.

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

```
helm install charts/ups-broker --name ups-broker --namespace ups-broker
```

# Step 2 - Creating a `ClusterServiceBroker` Resource

Because we haven't created any resources in the service-catalog API server yet,
querying service catalog returns an empty list of resources:

```console
$ svcat get brokers
  NAME   URL   STATUS
+------+-----+--------+
```

We'll register a broker server with the catalog by creating a new
[`ClusterServiceBroker`](../contrib/examples/walkthrough/ups-broker.yaml) resource.
Do so with the following command:

```console
$ kubectl create -f contrib/examples/walkthrough/ups-broker.yaml
servicebroker "ups-broker" created
```

When we create this `ClusterServiceBroker` resource, the service catalog controller responds
by querying the broker server to see what services it offers and creates a
`ClusterServiceClass` for each.

We can check the status of the broker using `svcat describe`:

```console
$ svcat describe broker ups-broker
  Name:     ups-broker
  URL:      http://ups-broker-ups-broker.ups-broker.svc.cluster.local
  Status:   Ready - Successfully fetched catalog entries from broker @ 2018-03-02 16:03:52 +0000 UTC
```

Notice that the status reflects that the broker's
catalog of service offerings has been successfully added to our cluster's
service catalog.

# Step 3 - Viewing `ClusterServiceClass`es and `ClusterServicePlan`s

The controller created a `ClusterServiceClass` for each service that the UPS broker
provides. We can view the `ClusterServiceClass` resources available by executing:

```console
$ svcat get classes
                NAME                        DESCRIPTION                         UUID
+-----------------------------------+-------------------------+--------------------------------------+
  user-provided-service               A user provided service   4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
  user-provided-service-single-plan   A user provided service   5f6e6cf6-ffdd-425f-a2c7-3c9258ad2468
```

The UPS broker provides a service with the external name
`user-provided-service`. Run the following command to see the details of this
offering:

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
```

Additionally, the controller created a `ClusterServicePlan` for each of the
plans for the broker's services. We can view the `ClusterServicePlan`
resources available in the cluster by executing:

```console
$ svcat get plans
   NAME                   CLASS                       DESCRIPTION                         UUID
+---------+-----------------------------------+-------------------------+--------------------------------------+
  default   user-provided-service               Sample plan description   86064792-7ea2-467b-af93-ac9694d96d52
  premium   user-provided-service               Premium plan              cc0d7529-18e8-416d-8946-6f7456acd589
  default   user-provided-service-single-plan   Sample plan description   96064792-7ea2-467b-af93-ac9694d96d52
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
```

# Step 4 - Creating a New `ServiceInstance`

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
serviceinstance "ups-instance" created
```

After the `ServiceInstance` is created, the service catalog controller will 
communicate with the appropriate broker server to initiate provisioning. 
Check the status of that process with this command:

```console
$ svcat describe instance -n test-ns ups-instance
  Name:        ups-instance
  Namespace:   test-ns
  Status:      Ready - The instance was provisioned successfully @ 2018-03-02 16:07:09 +0000 UTC
  Class:       user-provided-service
  Plan:        default

Bindings:
No bindings defined
```

# Step 5 - Requesting a `ServiceBinding` to use the `ServiceInstance`

Now that our `ServiceInstance` has been created, we can bind to it.
Create a `ServiceBinding` resource with the following command:

```console
$ kubectl create -f contrib/examples/walkthrough/ups-binding.yaml
servicebinding "ups-binding" created
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
  Status:      Ready - Injected bind result @ 2018-03-02 16:11:25 +0000 UTC
  Instance:    ups-instance
```

Notice that the status has a `Ready` condition set.  This means our binding is
ready to use!  If we look at the `Secret`s in our `test-ns` namespace, we should
see a new one:

```console
$ kubectl get secrets -n test-ns
NAME                              TYPE                                  DATA      AGE
default-token-3k61z               kubernetes.io/service-account-token   3         29m
ups-binding                       Opaque                                2         1m
```

Notice that a new `Secret` named `ups-binding` has been created.

# Step 6 - Deleting the `ServiceBinding`

Now, let's unbind the instance:

```
svcat unbind -n test-ns ups-instance
```

After the deletion is complete, we should see that the `Secret` is gone:

```console
$ kubectl get secrets -n test-ns
NAME                  TYPE                                  DATA      AGE
default-token-3k61z   kubernetes.io/service-account-token   3         30m
```

# Step 7 - Deleting the `ServiceInstance`

Now, we can deprovision the instance:

```
svcat deprovision -n test-ns ups-instance
```

# Step 8 - Deleting the `ClusterServiceBroker`

Next, we should remove the `ClusterServiceBroker` resource. This tells the service
catalog to remove the broker's services from the catalog. Do so with this
command:

```
kubectl delete clusterservicebrokers ups-broker
```

We should then see that all the `ClusterServiceClass` resources that came from that
broker have also been deleted by running the following command:

```console
$ svcat get classes
  NAME   DESCRIPTION   UUID
+------+-------------+------+
```

# Step 9 - Final Cleanup

## Cleaning up the UPS Service Broker Server

To clean up, delete the helm deployment:

```
helm delete --purge ups-broker
```

Then, delete all the namespaces we created:

```
kubectl delete ns test-ns ups-broker
```
## Cleaning up the Service Catalog

Delete the helm deployment and the namespace:

```
helm delete --purge catalog
kubectl delete ns catalog
```

# Troubleshooting

## Firewall rules

If you are using Google Cloud Platform, you may need to run the following
command to setup proper firewall rules to allow your traffic get in.

```
gcloud compute firewall-rules create allow-service-catalog-secure --allow tcp:30443 --description "Allow incoming traffic on 30443 port."
```
