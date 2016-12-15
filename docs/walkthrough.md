# An End-to-End Walkthrough

## Prerequisites and Setup

You will need to meet the following prerequisites:

* [Docker](https://www.docker.com) installed locally
* [Go](https://golang.org) set up locally (with proper `$GOPATH`
  and (optionally) `$GOPATH/bin` in your local `PATH`)
* [Glide](https://github.com/masterminds/glide) v0.12.3 or higher installed
  and on your `PATH`
* Working Kubernetes cluster and `kubectl` installed in your local `PATH`.
  The version of Kubernetes and `kubectl` must be >= 1.4
* [Helm](https://helm.sh) installed in your Kubernetes cluster,
  and the `helm` binary in your `PATH`
* Cluster credentials in ./kubeconfig file
* GNU Make
* [git](https://git-scm.com)
* Must be pre-authenticated to any Docker registry to which you intend to push
  images

## Getting Sources

With these prerequisites, clone the [Git repository
](https://github.com/kubernetes-incubator/service-catalog.git).

    cd $GOPATH
    git clone https://github.com/kubernetes-incubator/service-catalog.git \
        src/github.com/kubernetes-incubator/service-catalog

# Running in Kubernetes Cluster

## Building

Build the source. In order to build and push Docker images to Google
Container Registry you are going to need Google Cloud Project ID:

    cd "${GOPATH}/src/github.com/kubernetes-incubator/service-catalog"
    export VERSION=$(git rev-parse --short --verify HEAD)
    # Registry URL is the portion up to, but excluding the image name and its
    # preceding slash, e.g., "gcr.io/my-repo", "my-docker-id"
    export REGISTRY=<registry URL>

    make init build docker push

This will build Docker images for the service controller, Kubernetes service
broker, and service classes registry. The images are also pushed to Google
Container Registry (gcr.io) so they can be accessed by your Kubernetes cluster:

    docker images

The images are tagged with the current Git commit SHA.

## Deploying to Kubernetes

**NOTE**: Do not forget to specify a Kubernetes namespace where the system will
be deployed. Here, we will use `catalog`.

Use Helm to create the Kubernetes deployments:

    helm install \
        --set "registry=${REGISTRY},version=${VERSION}" \
        --namespace catalog \
        ./deploy/catalog

After the deployment, observe the deployments and services:

    kubectl get deployments,services --namespace catalog

## Walkthrough

Now that the system has been deployed to our Kubernetes cluster, multiple
new Kubernetes resources were registered. Service brokers, classes, instances,
and bindings. These resources are building blocks for composing services.

Because we didn't create any services yet, `kubectl get` will return an empty
list:

    kubectl get servicebrokers,serviceclasses,serviceinstances,servicebindings

**NOTE**: If there are any resouces, for example left over from an earlier
walkthrough, you can delete them using `script/cleanup.sh`.

Now we are ready to use service catalog. First, register service broker with the
catalog:

    cd contrib/examples/walkthrough/

    kubectl create -f broker.yaml

Confirm that service types are now available for instantiation:

    kubectl get serviceclasses

This will output available service types, for example:

    NAME               LABELS    DATA
    binding-consumer   <none>    {"apiVersion":"...
    booksbe            <none>    {"apiVersion":"...
    booksfe            <none>    {"apiVersion":"...

We can now create instances of these service classes and connect them
using bindings:

    # Create backend (MySQL) instance.
    kubectl create -f backend.yaml

    # Create binding called 'database'.
    kubectl create -f binding.yaml

Creating the binding will create a `ConfigMap` for binding consumption. This
can now be used by a native kubernetes app. You can inspect the config map
using `kubectl get configmap database`:

    NAME       DATA      AGE
    database   4         55s

Now you can deploy the application that consumes the binding.

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

## Consume a user-provided service

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
