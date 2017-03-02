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

## Overview

This repository is organized as similarly to Kubernetes itself as the developers
have found possible (or practical). Below is a summary of the repository's
layout:

    .
    ├── .glide             # Glide cache (untracked)
    ├── bin                # Destination for binaries compiled for linux/amd64 (untracked)
    ├── build              # Contains build-related scripts and subdirectories containing Dockerfiles
    ├── cmd                # Contains "main" Go packages for each service catalog component binary
    ├── contrib            # Contains all non-essential source
    │   └── hack           # Non-build related scripts
    ├── deploy             # Helm charts for deployment
    │   └── catalog        # Helm chart for deploying TPR-based prototype
    │   └── wip-catalog    # Helm chart for deploying apiserver-based WIP
    ├── docs               # Documentation
    ├── pkg                # Contains all non-"main" Go packages
    └── vendor             # Glide-managed dependencies (untracked)

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
* To be pre-authenticated to a Docker registry
* A working Kubernetes cluster and `kubectl` installed in your local `PATH`,
  properly configured to access that cluster. The version of Kubernetes and
  `kubectl` must be >= 1.4
* [Helm](https://helm.sh) (Tiller) installed in your Kubernetes cluster and the
  `helm` binary in your `PATH`

**Note:** It is not generally useful to run service catalog components outside
a Kubernetes cluster. As such, our build process only supports compilation of
linux/amd64 binaries suitable for execution within a Docker container.

## Cloning the Repo

The Service Catalog github repository can be found
[here](https://github.com/kubernetes-incubator/service-catalog.git).

To clone the repository:

    git clone https://github.com/kubernetes-incubator/service-catalog.git

## Building

First `cd` to the root of the cloned repository tree.
To build the service-catalog:

    make build

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

    make test-unit

These will execute any `*_test.go` files within the source tree.
The integration tests can be run via the `test-integration` Makefile target,
e.g.:

    make test-integration

The integration tests require the Kubernetes client (`kubectl`) so there is a
script called `contrib/hack/kubectl` that will run it from within a
Docker container. This avoids the need for you to download, or install it,
youself. You may find it useful to add `contrib/hack` to your `PATH`.

The `test` Makefile target will run both the unit and integration tests, e.g.:

    make test

If you want to run just a subset of the unit testcases then you can
specify the source directories of the tests:

    TEST_DIRS="path1 path2" make test

or you can specify a regexp expression for the test name:

    UNIT_TESTS=TestBar* make test

To see how well these tests cover the source code, you can use:

    make coverage

These will execute the tests and perform an analysis of how well they
cover all code paths. The results are put into a file called:
`coverage.html` at the root of the repo.

## Advanced Build Steps

You can build the service catalog executables into Docker images and push
them to a Docker Registry so they can be accessed by your Kubernetes clusters:

    # Registry URL is the portion up to, but excluding the image name and its
    # preceding slash, e.g., "gcr.io/my-repo", "my-docker-id"
    export REGISTRY=<registry URL>

    make images push

This will build Docker images for the service controller, Kubernetes service
broker, and service classes registry. The images are also pushed to the
registry specified by the `REGISTRY` environment variable, so they
can be accessed by your Kubernetes cluster.

The images are tagged with the current Git commit SHA:

    docker images

## Deploying to Kubernetes

**NOTE**: These instructions are for the TPR-based prototype and will change over
to use the API server and controller-manager soon.
**NOTE**: Do not forget to specify a Kubernetes namespace where the system will
be deployed. Here, we will use `catalog`.

Use Helm to create the Kubernetes deployments:

    export VERSION=$(git describe --tags --always --abbrev=7 --dirty)
    helm install \
        --set "registry=${REGISTRY},version=${VERSION}" \
        --namespace catalog \
        ./deploy/catalog

After the deployment, observe the deployments and services:

    kubectl get deployments,services --namespace catalog

### Walkthrough

Now that the system has been deployed to our Kubernetes cluster, multiple
new Kubernetes resources were registered. Service brokers, classes, instances,
and bindings. These resources are building blocks for composing services.

Because we didn't create any services yet, `kubectl get` will return an empty
list:

    kubectl get servicebrokers,serviceclasses,serviceinstances,servicebindings

**NOTE**: If there are any resources left over from an earlier walkthrough, you
can delete them using `contrib/hack/cleanup.sh`.

Now we are ready to use service catalog. First, register service broker with the
catalog:

    cd contrib/examples/walkthrough/

    kubectl create -f broker.yaml

Confirm that service types are now available for instantiation:

    kubectl get serviceclasses

This will output available service types, for example:

    NAME               LABELS    DATA
    booksbe            <none>    {"apiVersion":"...

We can now create instances of these service classes and connect them
using bindings:

    # Create backend (MySQL) instance.
    kubectl create -f backend.yaml

    # Create binding called 'database'.
    kubectl create -f binding.yaml

Creating the binding will create a `Secret` for binding consumption. This
can now be used by a native Kubernetes app. You can inspect the config map
using `kubectl get secret database`:

    NAME       DATA      AGE
    database   4         55s

Now you can deploy the application that consumes the binding:

    kubectl create -f ../user-bookstore-client/bookstore.yaml

This will create a Kubernetes service `user-bookstore-fe`. Wait for deployments
to start and an IP address of the frontend to be assigned. You can monitor the
external IP address creation using `kubectl get services`:

    NAME                    CLUSTER-IP       EXTERNAL-IP       PORT(S)    AGE
    cf-i-3a121d22-booksbe   10.107.254.221   <none>            3306/TCP   2m
    user-bookstore-fe       10.107.254.221   **<pending>**         3306/TCP   2m
    kubernetes              10.107.240.1     <none>            443/TCP    1d

Once the IP address becomes available we can use it to contact the frontend
endpoint. In this example, the IP address is `104.154.153.120`.

Save the IP address in an environment variable:

    IP=104.154.153.120

And interact with the Bookstore API:

    # List shelves
    curl "http://${IP}:8080/shelves"

    # List a specific shelf
    curl "http://${IP}:8080/shelves/1"

    # Create a new shelf
    curl -H 'Content-Type: application/json' \
         -H 'x-api-key: 123' \
          -d '{ "theme": "Travel" }' \
          "http://${IP}:8080/shelves"

    # Create a book on the shelf:
    curl -H 'Content-Type: application/json' \
         -H 'x-api-key: 123' \
         -d '{ "author": "Rick Steves", "title": "Travel as a Political Act" }' \
         "http://${IP}:8080/shelves/3/books"

    # List the books on the travel shelf:
    curl -H 'x-api-key: 123' "http://${IP}:8080/shelves/3/books"

    # Get the book:
    curl -H 'x-api-key: 123' "http://${IP}:8080/shelves/3/books/3"

### Consume a user-provided service

**NOTE**: This demo requires that you have cleaned up the resources created in
the previous demo, specifically the binding and frontend.

User-provided services are external services which need to be consumed by
Kubernetes applications through bindings.

Create a Kubernetes deployment. Even though this deployment is hosted in the
Kubernetes cluster, for the walkthrough it assumes the role of an external,
user-provided service:

    kubectl create -f contrib/examples/user-bookstore-mysql/bookstore.yaml

Create a User-provided Service instance to make your service bindable. You will
need to specify the hostname (either the service name from above or the IP
address of the service), port, username, and password.

    kubectl create -f contrib/examples/walkthrough/ups-backend.yaml

Now you can use the same steps as above to create a binding and a consuming
service.

    kubectl create -f contrib/examples/walkthrough/frontend.yaml
    kubectl create -f contrib/examples/walkthrough/binding.yaml
