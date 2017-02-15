# Service Catalog Helm Chart

The Helm Chart deploys Service Catalog into an existing Kubernetes cluster.

## Prerequisites

  - [Helm](https://github.com/kubernetes/helm) must be installed in the cluster.
  - Images must be built from source and pushed to a registry accessible to
    the cluster. Example (from this directory):

    ```
    $ export REGISTRY=hub.docker.com/<username>
    $ make -C ../.. push
    ```

    __Note:__ If deploying locally (e.g. in Minikube), it can be more efficient
    to use a _local_ Docker registry.

## Usage

Supported template parameters (values):

  - `registry`  (required): Container registry with Service Catalog images.
  - `version`   (optional): Version of Service Catalog (container images) to deploy.

Example:

```
$ export VERSION=$(git describe --tags --always --abbrev=7 --dirty)
$ helm install \
    --namespace catalog \
    --set registry=${REGISTRY},version=${VERSION} .
```
