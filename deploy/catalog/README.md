# Service Catalog Helm Chart

The Helm Chart deploys Service Catalog into an existing GKE cluster.

## Prerequisites

[Helm](https://github.com/kubernetes/helm) must be installed in the cluster.

## Usage

Supported template parameters (values):

  - `registry`  (required): Container registry with Service Catalog images.
  - `version`   (optional): Version of Service Catalog (container images) to deploy.
  - `namespace` (optional): A Kubernetes namespace to use for Service Catalog deployment.
