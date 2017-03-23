# Service Catalog Demonstration Walk-through

This document outlines the basic features of the service catalog by walking
through a basic use case.

## Step 0 - Prerequisites

### Starting Kubernetes with DNS

You *must* have a Kubernetes cluster with cluster DNS enabled. We can't list instructions here
for enabling cluster DNS for all Kubernetes cluster installations, but here are a few notes:

* If you are using Google Container Engine or minikube, you likely already have cluster DNS
enabled with no additional steps
* If you are using hack/local-up-cluster.sh, ensure to set `KUBE_ENABLE_CLUSTER_DNS` as follows:

```console
KUBE_ENABLE_CLUSTER_DNS=true hack/local-up-cluster.sh -O
```

### Getting Helm and installing Tiller

You must use Helm v2 or newer in the installation steps below. If you already have a `helm` v2 CLI,
execute `helm init` (if you haven't already) to set up Tiller (the server-side component of Helm), 
and you should be done with helm setup.

If you don't already have Helm v2, see [the installation instructions](https://github.com/kubernetes/helm/blob/master/docs/install.md). The instructions below will not
work with previous versions of Helm.

## Step 1 - Installing the Service Catalog System

The service catalog is conveniently packaged as a [helm](http://helm.sh/) 
chart for installation.

The chart is located in the [charts/catalog](../charts/catalog) directory in this
repository, and supports a wide variety of customizations which are laid out in
the 
[README.md](https://github.com/kubernetes-incubator/service-catalog/blob/master/charts/catalog/README.md) 
in that directory. To install the service-catalog with sensible defaults, execute this command from 
the Kubernetes context:

```console
helm install charts/catalog --name catalog --namespace catalog
```

Note: in the event you need to start the walk through over, the easiest way is
to execute `helm delete --purge <name>` for each helm install.

## Step 2 - Understand Service Catalog Components

Now that the system has been deployed to our Kubernetes cluster, we can use
`kubectl` to talk to the service catalog API server.  The service catalog API
has five main concepts:

- Broker Server: A server that acts as a service broker and conforms to the 
    [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md)
- `Broker` Resource: a representation of the broker server in the service-catalog system
    that indicates where a broker server is running
- `ServiceClass`: a service offered by a particular service broker. `ServiceClass`es are created
    in response to a `Broker` resource being submitted to the system
- `Instance`: an instance of a `ServiceClass` provisioned by the `Broker` for that `ServiceClass`
- `Binding`: a binding to an `Instance` which is manifested into a Kubernetes namespace

These components are building blocks of the service catalog in Kubernetes from an
API standpoint.

## Step 3 - Installing a UPS broker

Service Catalog requires broker servers to operate properly. There is a User 
Provided Service broker (UPS from now on), which allows consumption of existing
services through the Service Catalog model. Just like any other broker, the
UPS broker needs to be running somewhere before it can be added to the
catalog. We need to deploy it first by using the
[`ups-broker` Helm chart](../charts/ups-broker) into your cluster, just like
you installed the catalog chart above. To install the broker with sensible
defaults, execute this command from within the Kubernetes context:

```console
helm install charts/ups-broker --name ups-broker --namespace ups-broker
```

## Step 4 - Install and Configure `kubectl` 1.6

Before we begin doing operations on the service catalog API server, we'll
need to do two setup steps:

- Download and install `kubectl` version 1.6
- Add entries to our `kubeconfig` file that tell `kubectl` how to talk to the service catalog API
server

To install `kubectl` 1.6, simply execute the following:

```console
curl -o kubectl https://storage.googleapis.com/kubernetes-release/release/v1.6.0-beta.3/bin/darwin/amd64/kubectl
chmod +x ./kubectl
```

We'll assume that, after this step, all `kubectl` commands are using this newly-downloaded 
executable. Now, we'll add configuration to allow `kubectl` to talk to our service catalog API
server:

```console
kubectl config set-cluster service-catalog --server=http://$SVC_CAT_API_SERVER_IP:80
kubectl config set-context service-catalog --cluster=service-catalog
```

Note that you'll need to specify the service IP of your service catalog API server, and 
substitute `$SVC_CAT_API_SERVER_IP` for that IP.

## Step 5 - Creating a `Broker` Resource

Next, we'll register a service broker with the catalog.  To do this, we'll
create a new [`Broker`](../contrib/examples/walkthrough/ups-broker.yaml)
resource against our API server.

Because we haven't created any resources in the service-catalog API server yet,
`kubectl get` will return an empty list of resources.

```console
kubectl --context=service-catalog get brokers,serviceclasses,instances,bindings
```


Then, create the new `Broker` resource with the commands below:

```console
./kubectl --context=service-catalog create -f contrib/examples/walkthrough/ups-broker.yaml
```

The output of that command should be the following:

```console
broker "ups-broker" created
```

Kubernetes APIs are intention based; creating this resource indicates that the
want for the service broker it represents to be consumed in the catalog.  When
we create the resource, the controller handles loading that broker into the
catalog by seeing what services it provides and adding them to the catalog.

We can check the status of the broker using `kubectl get`:

```console
kubectl --context=service-catalog get brokers ups-broker -o yaml
```

We should see something like:

```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: Broker
metadata:
  creationTimestamp: 2017-03-03T04:11:17Z
  finalizers:
  - kubernetes
  name: ups-broker
  resourceVersion: "6"
  selfLink: /apis/servicecatalog.k8s.io/v1alpha1/brokers/ups-broker
  uid: 72fa629b-ffc7-11e6-b111-0242ac110005
spec:
  url: http://ups-broker.ups-broker.svc.cluster.local:8000
status:
  conditions:
  - message: Successfully fetched catalog from broker
    reason: FetchedCatalog
    status: "True"
    type: Ready
```

Notice that the controller has set this brokers `status` field to reflect that
it's catalog has been added to our cluster's catalog.

## Step 6 - Viewing ServiceClasses

The controller created a `ServiceClass` for each service that the broker we
added provides. We can view the `ServiceClass` resources available in the
cluster by doing:

```console
kubectl get serviceclasses
NAME                    KIND
user-provided-service   ServiceClass.v1alpha1.servicecatalog.k8s.io
```

The `Broker` resource we created points to our UPS broker, which provides a service
called the `user-provided-service`.  Run the following command to get detail on this service:

```console
kubectl --context=service-catalog get serviceclasses user-provided-service -o yaml
```

We should see something like:

```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: ServiceClass
metadata:
  creationTimestamp: 2017-03-03T04:11:17Z
  name: user-provided-service
  resourceVersion: "7"
  selfLink: /apis/servicecatalog.k8s.io/v1alpha1/serviceclassesuser-provided-service
  uid: 72fef5ce-ffc7-11e6-b111-0242ac110005
brokerName: ups-broker
osbGuid: 4F6E6CF6-FFDD-425F-A2C7-3C9258AD2468
bindable: false
planUpdatable: false
plans:
- name: default
  osbFree: true
  osbGuid: 86064792-7ea2-467b-af93-ac9694d96d52
```

## Step 7 - Provisioning a new Instance

Now that one or more `ServiceClass` resources are in the catalog, we can provision a new
instance of the `user-provided-service`. We do this by creating a new 
[`Instance`](../contrib/examples/walkthrough/ups-instance.yaml) resource for each provision:

```console
./kubectl --context=service-catalog create -f contrib/examples/walkthrough/ups-instance.yaml
```

That operation should output:

```console
instance "ups-instance" created
```

Now that the new `Instance` is created, we can check the status of it with:

```console
kubectl --context=service-catalog get instances -n test-ns ups-instance -o yaml
```

We should see something like:

```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: Instance
metadata:
  creationTimestamp: 2017-03-03T04:26:08Z
  name: ups-instance
  namespace: test-ns
  resourceVersion: "9"
  selfLink: /apis/servicecatalog.k8s.io/v1alpha1/namespaces/test-ns/instances/ups-instance
  uid: 8654e626-ffc9-11e6-b111-0242ac110005
spec:
  osbGuid: 34c984e1-4626-4574-8a95-9e500d0d48d3
  planName: default
  serviceClassName: user-provided-service
status:
  conditions:
  - message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
```

## Step 8 - Bind to the Instance

Now that our `Instance` has been created, we can bind to it. After the bind
operation is complete on the UPS broker server, the service catalog will write
the resulting credentials to a Kubernetes secret. 

After a [`Binding`](../contrib/examples/walkthrough/ups-binding.yaml) resource is created, 
the service catalog will execute a bind operation on  the broker server and write the 
results to a `Secret` in the same namespace as the `Binding` itself.

In order for that operation to succeed, we'll need to ensure that namespace exists in 
Kubernetes. Do so by executing this command:

```console
kubectl --context=service-catalog create namespace test-ns
```

Then, create the `Binding`:

```console
kubectl --context=service-catalog create -f contrib/examples/walkthrough/ups-binding.yaml
```

That command should output:

```console
binding "ups-binding" created
```

We can check the status of the `Binding` using `kubectl get`:

```console
kubectl --context=service-catalog get bindings -n test-ns ups-binding -o yaml
```

We should see something like:

```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: Binding
metadata:
  creationTimestamp: 2017-03-07T01:44:36Z
  finalizers:
  - kubernetes
  name: ups-binding
  namespace: test-ns
  resourceVersion: "29"
  selfLink: /apis/servicecatalog.k8s.io/v1alpha1/namespaces/test-ns/bindings/ups-binding
  uid: 9eb2cdce-02d7-11e7-8edb-0242ac110005
spec:
  instanceRef:
    name: ups-instance
  osbGuid: b041db94-a5a0-41a2-87ae-1025ba760918
  secretName: my-secret
status:
  conditions:
  - message: Injected bind result
    reason: InjectedBindResult
    status: "True"
    type: Ready
```

Notice that the status has a ready condition set.  This means our binding is
ready to use.  If we look at the secrets in our `test-ns` namespace in
Kubernetes, we should see:

```console
kubectl get secrets -n test-ns
NAME                  TYPE                                  DATA      AGE
default-token-3k61z   kubernetes.io/service-account-token   3         29m
my-secret             Opaque                                2         1m
```

Notice that a secret named `my-secret` has been created in our namespace.

## Step 9 - Unbind from the Instance

Now, let's unbind from the Instance.  To do this, we just delete the `Binding`
that we created:

```console
kubectl --context=service-catalog delete -n test-ns bindings ups-binding
```

If we check the secrets in the `test-ns` namespace, we should see that the
secret we were injected with has been deleted:

```console
kubectl get secrets -n test-ns
NAME                  TYPE                                  DATA      AGE
default-token-3k61z   kubernetes.io/service-account-token   3         30m
```

## Step 10 - Deprovision the Instance

Now, we can deprovision the instance.  To do this, we just delete the `Instance`
that we created:

```console
kubectl --context=service-catalog delete -n test-ns instances ups-instance
```

### Delete the broker

When an administrator wants to remove a broker and the services it offers from
the catalog, they can just delete the broker:

```console
kubectl --context=service-catalog delete brokers ups-broker
```

And we should see that all the `ServiceClass` resources that came from that
broker were cleaned up:

```console
kubectl get serviceclasses
No resources found
```

## Step 11 - Final clean up

Delete the test-ns namespace:

```console
kubectl delete namespace test-ns
```
