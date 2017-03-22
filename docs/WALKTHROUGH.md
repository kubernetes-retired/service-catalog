# Service Catalog Demonstration Walk-through

This document outlines the basic features of the service catalog by walking
through a basic use case.

## Step 0 - Prerequisites

### Starting Kubernetes with DNS

You *must* have a Kubernetes cluster with cluster DNS enabled.

* If you are using Google Container Engine or minikube, no additional action is
required.
* If you are using hack/local-up-cluster.sh, ensure to set
KUBE_ENABLE_CLUSTER_DNS:

```console
KUBE_ENABLE_CLUSTER_DNS=true hack/local-up-cluster.sh -O
```

### Getting Helm and installing Tiller

If you already have Helm v2 or newer, executing `helm init` should be all that's
required. Detailed helm installation instructions are available
[here](https://github.com/kubernetes/helm/blob/master/docs/install.md). The
charts will not work with Helm Classic.

## Step 1 - Installing the Service Catalog System

The service catalog is conveniently packaged as a [helm](http://helm.sh/) 
chart for installation.

The chart is located in the [charts/catalog](../charts/catalog) directory in this
repository, and supports a wide variety of customizations which are laid out in
the README.md in that directory. To install the service-catalog with sensible
defaults, execute this command from the Kubernetes context:

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

----

#### Note: accessing the service catalog

Unfortunately, `kubectl` doesn't know how to speak to both the service catalog
API server and the main Kubernetes API server without switching contexts or
`kubeconfig` files. One way to access the service catalog API server is via a
dedicated `kubeconfig` file. You can manage the kubeconfig in use within a
directory using the `direnv` tool. Alternatively, the context can be switched
using kubectl commands shown later.

### Creating kubeconfig for use with direnv
Make sure to change to the service catalog directory and create the .kubeconfig
file copying the following into your shell:

```console
cat << EOF > .kubeconfig
apiVersion: v1
clusters:
- cluster:
    server: http://<api server ip>:80
  name: service-catalog-cluster
contexts:
- context:
    cluster: service-catalog-cluster
  name: service-catalog-ctx
current-context: service-catalog-ctx
kind: Config
preferences: {}
EOF
```

The above `.kubeconfig` file used in combination with `direnv` is the simplest
way to switch between the core API server and the service-catalog API server.
An example of an .envrc file in the service catalog directory would look like:

```console
export KUBECONFIG=<path to service-catalog>/.kubeconfig
```

Then kubectl commands executed from the service-catalog directory will
now make requests to the correct API server.

----

Create a test-ns namespace while still executing Kubernetes API commands:

```console
kubectl create namespace test-ns
```

This is necessary because the `Instance` and `Binding` resources are namespaced,
so they require a namespace.

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

## Step 4 - Creating a `Broker` Resource

Next, we'll register a service broker with the catalog.  To do this, we'll
create a new [`Broker`](../contrib/examples/walkthrough/ups-broker.yaml)
resource against our API server.

Before we do so, we'll need to configure our `kubeconfig` file as described
above, and download a 1.6-beta version of `kubectl`. We'll need this version of
`kubectl` for all `create` operations against the service-catalog API server. For
simplicity, all subsequent kubectl commands will be assumed to be using this
newly downloaded version.

Download & install the `kubectl` 1.6-beta version with this command:

```console
curl -o kubectl16 https://storage.googleapis.com/kubernetes-release/release/v1.6.0-beta.3/bin/darwin/amd64/kubectl
chmod +x ./kubectl16
```

Because we haven't created any resources in the service-catalog API server yet,
`kubectl get` will return an empty list of resources. (The first two commands
may be skipped if direnv is setup.)

```console
kubectl config set-cluster service-catalog --server=http://<api server ip>:80
kubectl config set-context service-catalog --cluster=service-catalog
kubectl get brokers,serviceclasses,instances,bindings
```


Then, create the new `Broker` resource with the commands below:

```console
./kubectl16 create -f contrib/examples/walkthrough/ups-broker.yaml
```

The output of that command should be the following:

```console
broker "ups-broker" created
```

Kubernetes APIs are intention based; creating this resource indicates that the
want for the service broker it represents to be consumed in the catalog.  When
we create the resource, the controller handles loading that broker into the
catalog by seeing what services it provides and adding them to the catalog.

We can check the status of the broker using `kubectl get` (notice that we don't have to use 
`./kubectl16` here, because we only need that version to execute `create` operations against our
server):

```console
kubectl get brokers ups-broker -o yaml
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

## Step 5 - Viewing ServiceClasses

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
kubectl get serviceclasses user-provided-service -o yaml
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

## Step 6 - Provisioning a new Instance

Now that one or more `ServiceClass` resources are in the catalog, we can provision a new
instance of the `user-provided-service`. We do this by creating a new 
[`Instance`](../contrib/examples/walkthrough/ups-instance.yaml) resource for each provision:

```console
./kubectl16 create -f contrib/examples/walkthrough/ups-instance.yaml
```

That operation should output:

```console
instance "ups-instance" created
```

Now that the new `Instance` is created, we can check the status of it with:

```console
kubectl get instances -n test-ns ups-instance -o yaml
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

## Step 7 - Bind to the Instance

Now that our `Instance` has been created, we can bind to it. After the bind
operation is complete on the UPS broker server, the service catalog will write
the resulting credentials to a Kubernetes secret. As mentioned in step 2, this is
an action that requires the test-ns namespace to be present. Here's how we create
the new [`Binding`](../contrib/examples/walkthrough/ups-binding.yaml):


```console
kubectl create -f contrib/examples/walkthrough/ups-binding.yaml
```

That command should output:

```console
binding "ups-binding" created
```

We can check the status of the `Binding` using `kubectl get`:

```console
kubectl get bindings -n test-ns ups-binding -o yaml
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

## Step 8 - Unbind from the Instance

Now, let's unbind from the Instance.  To do this, we just delete the `Binding`
that we created:

```console
kubectl delete -n test-ns bindings ups-binding
```

If we check the secrets in the `test-ns` namespace, we should see that the
secret we were injected with has been deleted:

```console
kubectl get secrets -n test-ns
NAME                  TYPE                                  DATA      AGE
default-token-3k61z   kubernetes.io/service-account-token   3         30m
```

## Step 9 - Deprovision the Instance

Now, we can deprovision the instance.  To do this, we just delete the `Instance`
that we created:

```console
kubectl delete -n test-ns instances ups-instance
```

### Delete the broker

When an administrator wants to remove a broker and the services it offers from
the catalog, they can just delete the broker:

```console
kubectl delete brokers ups-broker
```

And we should see that all the `ServiceClass` resources that came from that
broker were cleaned up:

```console
kubectl get serviceclasses
No resources found
```

## Step 10 - Final clean up

Delete the test-ns namespace:

```console
kubectl delete namespace test-ns
```
