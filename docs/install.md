# Installing Service Catalog

Kubernetes 1.7 or higher clusters run the
[API Aggregator](https://kubernetes.io/docs/concepts/api-extension/apiserver-aggregation/),
which is a specialized proxy server that sits in front of the core API Server.

Service Catalog provides an API server that sits behind the API aggregator, 
so you'll be using `kubectl` as normal to interact with Service Catalog.

To learn more about API aggregation, please see the 
[Kubernetes documentation](https://kubernetes.io/docs/concepts/api-extension/apiserver-aggregation/).

The rest of this document details how to:

- Set up Service Catalog on your cluster
- Interact with the Service Catalog API

# Step 1 - Prerequisites

## Starting Kubernetes with DNS

You *must* have a Kubernetes cluster with cluster DNS enabled. We can't list
instructions here for enabling cluster DNS for all Kubernetes distributions, but
here are a few notes:

* If you are using a cloud-based Kubernetes cluster or minikube, you likely have
cluster DNS enabled already.
* If you are using `hack/local-up-cluster.sh`, ensure the
`KUBE_ENABLE_CLUSTER_DNS` environment variable is set, then run the install
script.

## Helm

You *must* use [Helm](http://helm.sh/) v2.7.0 or newer in the installation
steps below.

### If You Don't Have Helm Installed

If you don't have Helm installed already, 
[download the `helm` CLI](https://github.com/kubernetes/helm#install) and
then run `helm init` (this installs Tiller, the server-side component of
Helm, into your Kubernetes cluster).

### If You Already Have Helm Installed

If you already have Helm installed, run `helm version` and ensure that both
the client and server versions are `v2.7.0` or above.

If they aren't, 
[install a newer version of the `helm` CLI](https://github.com/kubernetes/helm#install)
and run `helm init --upgrade`. 

For more details on installation, see the
[Helm installation instructions](https://github.com/kubernetes/helm/blob/master/docs/install.md).

### Helm Charts

Service Catalog is easily installed via a 
[Helm chart](https://github.com/kubernetes/helm/blob/master/docs/charts.md).

Before installation, add the service-catalog Helm repository to your local machine:

```console
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
```

Then, ensure that the repository was successfully added:

```console
helm search service-catalog
```

You should see the following output:

```console
NAME           	VERSION	DESCRIPTION
svc-cat/catalog	x,y.z  	service-catalog API server and controller-manag...
```

If you see it, your repository is properly added.

## RBAC

Your Kubernetes cluster must have 
[RBAC](https://kubernetes.io/docs/admin/authorization/rbac/) enabled to use
Service Catalog.

If you are using Minikube, make sure to run your `minikube start` command with
this flag:

```console
minikube start --extra-config=apiserver.Authorization.Mode=RBAC
```

If you are using `hack/local-up-cluster.sh`, ensure the
`AUTHORIZATION_MODE` environment variable is set as follows:

```console
AUTHORIZATION_MODE=Node,RBAC hack/local-up-cluster.sh -O
```

### Tiller Permissions

Tiller is the in-cluster server component of Helm. By default, 
`helm init` installs the Tiller pod into the `kube-system` namespace,
and configures Tiller to use the `default` service account.

Tiller will need to be configured with `cluster-admin` access to properly install
Service Catalog:

```console
kubectl create clusterrolebinding tiller-cluster-admin \
    --clusterrole=cluster-admin \
    --serviceaccount=kube-system:default
```

## A Recent kubectl

As with Kubernetes itself, interaction with the service catalog system is
achieved through the `kubectl` command line interface. Service Catalog 
requires `kubectl` version 1.7 or newer.

To check your version of `kubectl`, run:

```console
kubectl version
```

Recall that the server version must be `1.7` or above. If the client version
is below 1.7, follow the 
[installation instructions](https://kubernetes.io/docs/tasks/kubectl/install/) 
to get a new `kubectl` binary.

For example, run the following command to get an up-to-date binary on Mac OS:

```console
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/darwin/amd64/kubectl
chmod +x ./kubectl
```

# Step 2 - Install Service Catalog

Now that your cluster and Helm are configured properly, installing 
Service Catalog is simple:

```console
helm install svc-cat/catalog \
    --name catalog --namespace catalog
```
