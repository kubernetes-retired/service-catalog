# Service Catalog WIP Helm Chart

The Helm Chart deploys Service Catalog into an existing Kubernetes cluster.

## Prerequisites

- [Helm](https://github.com/kubernetes/helm) must be installed in the cluster
- Images must be built from source and pushed to a registry accessible to the
  cluster if deploying to a non-local cluster; see the [dev guide](../../docs/DEVGUIDE.md) for instructions

## Usage

Supported template parameters (values):

### Deployment Knobs

- `registry`: Container registry with Service Catalog images; optional; defaults
  to unset
- `namespace`: A Kubernetes namespace to use for deployment of this
  chart into k8s; optional; defaults to `default`
- `version`: Version of Service Catalog (container images) to deploy; optional;
  defaults to `latest`
- `apiServerVersion`: The version of the API server image to deploy; optional;
  defaults to unset; overrides `version` for the API server if set
- `controllerManagerVersion`: The version of the controller-manager image to
  deploy; optional; defaults to unset; overrides `version` for the
  controller-manager if set
- `etcdImage`: The name of the etcd image to use; optional; defaults to
  `quay.io/coreos/etcd`
- `etcdVersion`: The version of the etcd image to deploy; optional; defaults to 
  unset; overrides `version` for etcd if set

### Catalog knobs

- `verbosity`: The verbosity of logs; optional; defaults to `10`
- `debug`: Whether to create a load balancer for the apiserver and
  controller-manager services; optional; defaults to `false`
- `insecure`: Whether the API server should serve insecurely; optional; defaults
  to `false`
- `imagePullPolicy`: The image pull policy to use for all pods in this chart;
  optional; defaults to `Always`

### API Server knobs

- `insecure`: Whether the API server should serve insecurely; optional; defaults
  to `false`
- `insecurePort`: When `insecure` is true, the API server pod serves insecurely
  on this container port; optional; defaults to `8081`
- `insecureServicePort`: When `insecure` is true, the API server is fronted by a
  k8s service serving on this node port; optional; defaults to `30001`
- `storageType`: The type of storage for the API server to use; optional;
  defaults to `etcd`, also accepts `tpr`
- `globalNamespace`: The namespace to store global resources in when the API
  server is backed by TPR

## Examples

### Local cluster installation, backed with etcd

This helm command installs into a local cluster, backing the API server with
etcd, and serving insecurely on node port 30000:

```console
helm install \
    --namespace=service-catalog \
    --set storageType=etcd,insecure=true,debug=true,nodePort=30000,imagePullPolicy=Never \
    deploy/wip-catalog
```

### Local cluster installation, backed with TPR

This helm command installs into a local cluster, backing the API server with
third party resources in the main Kubernetes API server:

```console
helm install \
    --namespace=service-catalog-tpr \
    --set version=${VERSION},storageType=tpr,debug=true,insecure=true,imagePullPolicy=Never,globalNamespace=service-catalog-global \
    deploy/wip-catalog
```