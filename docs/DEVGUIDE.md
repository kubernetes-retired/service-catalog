# Developer's Guide to Service-Catalog

Table of Contents
- [Overview](#overview)
- [Working on Issues](#working-on-issues)
- [Prerequisites](#prerequisites)
- [Cloning the Repo](#cloning-the-repo)
- [Building](#building)
- [Testing](#testing)
- [Advanced Build Steps](#advanced-build-steps)
- [Deploying to Kubernetes](#deploying-to-kubernetes)
- [Demo walkthrough](#demo-walkthrough)

## Overview

This repository is organized as similarly to Kubernetes itself as the developers
have found possible (or practical). Below is a summary of the repository's
layout:

    .
    ├── .glide                  # Glide cache (untracked)
    ├── bin                     # Destination for binaries compiled for linux/amd64 (untracked)
    ├── build                   # Contains build-related scripts and subdirectories containing Dockerfiles
    ├── cmd                     # Contains "main" Go packages for each service catalog component binary
    │   └── apiserver           # The service catalog API server binary
    │   └── controller-manager  # The service catalog controller manager binary
    ├── contrib                 # Contains all non-essential source
    │   └── hack                # Non-build related scripts
    ├── deploy                  # Helm charts for deployment
    │   └── wip-catalog         # Helm chart for deploying apiserver-based WIP
    ├── docs                    # Documentation
    ├── pkg                     # Contains all non-"main" Go packages
    └── vendor                  # Glide-managed dependencies (untracked)

## Working on Issues

Github does not allow non-maintainers to assign, or be assigned to, issues.
As such non-maintainers can indicate their desire to work on (own) a particular
issue by adding a comment to it of the form:

	#dibs

However, it is a good idea to discuss the issue, and your intent to work on it,
with the other members via the slack channel to make sure there isn't some
other work alread going on with respect to that issue.

When you create a pull request (PR) that completely addresses an open issue
please include a line in the initial comment that looks like:

	Closes: #1234

where `1234` is the issue number. This allows Github to automatically
close the issue when the PR is merged.

## Prerequisites

At a minimum you will need:

* [Docker](https://www.docker.com) installed locally
* GNU Make
* [git](https://git-scm.com)

These will allow you to build and test service catalog components within a
Docker container.

If you want to deploy service catalog components built from source, you will
also need:

* A working Kubernetes cluster and `kubectl` installed in your local `PATH`,
  properly configured to access that cluster. The version of Kubernetes and
  `kubectl` must be >= 1.4
* [Helm](https://helm.sh) (Tiller) installed in your Kubernetes cluster and the
  `helm` binary in your `PATH`
* To be pre-authenticated to a Docker registry (if using a remote cluster)

**Note:** It is not generally useful to run service catalog components outside
a Kubernetes cluster. As such, our build process only supports compilation of
linux/amd64 binaries suitable for execution within a Docker container.

## Cloning the Repo

The Service Catalog github repository can be found
[here](https://github.com/kubernetes-incubator/service-catalog.git).

To clone the repository:

    $ git clone https://github.com/kubernetes-incubator/service-catalog.git

## Building

First `cd` to the root of the cloned repository tree.
To build the service-catalog:

    $ make build

The above will build all executables and place them in the `bin` directory. This
is done within a Docker container-- meaning you do not need to have all of the
necessary tooling installed on your host (such as a golang compiler or glide).
Building outside the container is possible, but not officially supported.

Note, this will do the basic build of the service catalog. There are more
more [advanced build steps](#advanced-build-steps) below as well.

To deploy to Kubernetes, see the
[Deploying to Kubernetes](#deploying-to-kubernetes) section.

### Notes Concerning the Build Process/Makefile

* The Makefile assumes you're running `make` from the root of the repo.
* There are some source files that are generated during the build process.
  These will be prefixed with `zz`.
* A Docker Image called "scbuildimage" will be used. The image isn't pre-built
  and pulled from a public registry. Instead, it is built from source contained
  within the service catalog repository.
* While many people have utilities, such as editor hooks, that auto-format
  their go source files with `gofmt`, there is a Makefile target called
  `format` which can be used to do this task for you.
* `make build` will build binaries for linux/amd64 only.

## Testing

There are two types of tests: unit and integration. The unit testcases
can be run via the `test-unit` Makefile target, e.g.:

    $ make test-unit

These will execute any `*_test.go` files within the source tree.
The integration tests can be run via the `test-integration` Makefile target,
e.g.:

    $ make test-integration

The integration tests require the Kubernetes client (`kubectl`) so there is a
script called `contrib/hack/kubectl` that will run it from within a
Docker container. This avoids the need for you to download, or install it,
youself. You may find it useful to add `contrib/hack` to your `PATH`.

The `test` Makefile target will run both the unit and integration tests, e.g.:

    $ make test

If you want to run just a subset of the unit testcases then you can
specify the source directories of the tests:

    $ TEST_DIRS="path1 path2" make test

or you can specify a regexp expression for the test name:

    $ UNIT_TESTS=TestBar* make test

To see how well these tests cover the source code, you can use:

    $ make coverage

These will execute the tests and perform an analysis of how well they
cover all code paths. The results are put into a file called:
`coverage.html` at the root of the repo.

## Advanced Build Steps

You can build the service catalog executables into Docker images and push
them to a Docker Registry so they can be accessed by your Kubernetes clusters:

    # Registry URL is the portion up to, but excluding the image name and its
    # preceding slash, e.g., "gcr.io/my-repo", "my-docker-id"
    $ export REGISTRY=<registry URL>

    $ make images push

This will build Docker images for the service controller, Kubernetes service
broker, and service classes registry. The images are also pushed to the
registry specified by the `REGISTRY` environment variable, so they
can be accessed by your Kubernetes cluster.

The images are tagged with the current Git commit SHA:

    $ docker images

----

#### Tip: managing environment variables with direnv

The [direnv](https://www.direnv.net)
([github](https://github.com/direnv/direnv)) helps manages values of environment
variables within a directory. This can be very convenient when setting variables
like `KUBECONFIG` or `REGISTRY` within the service catalog directory.

Once you [install direnv](https://github.com/direnv/direnv#install), you can
create an `.envrc` file in your `service-catalog` directory:

```
$ cat .envrc
export REGISTRY="hub.docker.io/yippee"
export KUBECONFIG=/home/yippee/code/service-catalog/.kubeconfig
```
----

## Deploying to Kubernetes

Use the [`wip-catalog` chart](../deploy/wip-catalog) to deploy the service
catalog into your cluster.  The easiest way to get started is to deploy into a
cluster you regularly use and are familiar with.  One of the choices you can
make when deploying the catalog is whether to back the API server with etcd or
third party resources.  Currently, etcd is the best option; TPR support is
experimental and still under development.

## Demo Walkthrough

The rest of this guide is a walkthrough that is essentially the same as a
basic demo of the catalog.

Now that the system has been deployed to our Kubernetes cluster, we can use
`kubectl` to talk to the service catalog API server.  The service catalog API
has four resources:

- `Broker`: a service broker whose services appear in the catalog
- `ServiceClass`: a service offered by a particular service broker
- `Instance`: an instance of a `ServiceClass` provisioned by the `Broker` for
  that `ServiceClass`
- `Binding`: a binding to an `Instance` which is manifested into a Kubernetes
  namespace

These resources are building blocks of the service catalog in Kubernetes from an
API standpoint.

----

#### Note: accessing the service catalog

Unfortunately, `kubectl` doesn't know how to speak to both the service catalog
API server and the main Kubernetes API server without switching contexts or
`kubeconfig` files.  For now, the best way to access the service catalog API
server is via a dedicated `kubeconfig` file.  You can manage the kubeconfig in
use within a directory using the `direnv` tool.

----

Because we haven't created any resources in the service-catalog API server yet,
`kubectl get` will return an empty list of resources:

    $ kubectl get brokers,serviceclasses,instances,bindings

### Registering a Broker

First, we'll register a service broker with the catalog.  To do this, we'll
create a new [`Broker`](../contrib/examples/walkthrough/ups-broker.yaml)
resource:

    $ kubectl create -f contrib/examples/walkthrough/ups-broker.yaml
    broker "ups-broker" created

Kubernetes APIs are intention based; creating this resource indicates that the
want for the service broker it represents to be consumed in the catalog.  When
we create the resource, the controller handles loading that broker into the
catalog by seeing what services it provides and adding them to the catalog.

We can check the status of the broker using `kubectl get`:

    $ kubectl get brokers ups-broker -o yaml

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

### Viewing ServiceClasses

The controller created a `ServiceClass` for each service that the broker we
added provides. We can view the `ServiceClass` resources available in the
cluster by doing:

    $ kubectl get serviceclasses
    NAME                    KIND
    user-provided-service   ServiceClass.v1alpha1.servicecatalog.k8s.io

It looks like the broker we added provides a service called the `user-provided-
service`.  Let's check it out:

    $ kubectl get serviceclasses user-provided-service -o yaml

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

### Provisioning a new Instance

Let's provision a new instance of the `user-provided-service`.  To do this, we
create a new [`Instance`](../contrib/examples/walkthrough/ups-instance.yaml) to
indicate that we want to provision a new instance of that service:

    $ kubectl create -f contrib/examples/walkthrough/ups-instance.yaml
    instance "ups-instance" created

We can check the status of the `Instance` using `kubectl get`:

    $ kubectl get instances -n test-ns ups-instance -o yaml

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

### Bind to the Instance

Now that our `Instance` has been created, let's bind to it.  To do this, we
create a new [`Binding`](../contrib/examples/walkthrough/ups-binding.yaml).

    $ kubectl create -f contrib/examples/walkthrough/ups-binding.yaml
    binding "ups-binding" created

We can check the status of the `Instance` using `kubectl get`:

    $ kubectl get bindings -n test-ns ups-binding -o yaml

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
kubernetes, we should see:

    $ kubectl get secrets -n test-ns
    NAME                  TYPE                                  DATA      AGE
    default-token-3k61z   kubernetes.io/service-account-token   3         29m
    my-secret             Opaque                                2         1m

Notice that a secret named `my-secret` has been created in our namespace.

### Unbind from the Instance

Now, let's unbind from the Instance.  To do this, we just delete the `Binding`
that we created:

    $ kubectl delete -n test-ns bindings ups-binding

If we check the secrets in the `test-ns` namespace, we should see that the
secret we were injected with has been deleted:

    $ kubectl get secrets -n test-ns
    NAME                  TYPE                                  DATA      AGE
    default-token-3k61z   kubernetes.io/service-account-token   3         30m

### Deprovision the Instance

Now, we can deprovision the instance.  To do this, we just delete the `Instance`
that we created:

    $ kubectl delete -n test-ns instances ups-instance

### Delete the broker

When an administrator wants to remove a broker and the services it offers from
the catalog, they can just delete the broker:

    $ kubectl delete brokers ups-broker

And we should see that all the `ServiceClass` resources that came from that
broker were cleaned up:

    $ kubectl get serviceclasses
    No resources found
