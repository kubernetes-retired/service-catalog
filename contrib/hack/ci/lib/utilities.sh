# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Library of useful utilities for CI purposes.
#

readonly LIB_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# Prints first argument as header. Additionally prints current date.
shout() {
    echo -e "
#################################################################################################
# $(date)
# $1
#################################################################################################
"
}


# Installs kind dependency locally.
# Required envs:
#  - KIND_VERSION
#  - INSTALL_DIR
#
# usage: env INSTALL_DIR=/tmp KIND_VERSION=v0.4.0 install::local::kind
install::local::kind() {
    mkdir -p "${INSTALL_DIR}/bin"
    export PATH="${INSTALL_DIR}/bin:${PATH}"

    pushd "${INSTALL_DIR}"

    os=$(host::os)
    arch=$(host::arch)

    shout "- Install kind ${KIND_VERSION} locally to a tempdir..."

    curl -sSLo kind "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-${os}-${arch}"
    chmod +x kind
    mv kind "${INSTALL_DIR}/bin"

    popd
}

host::os() {
  local host_os
  case "$(uname -s)" in
    Darwin)
      host_os=darwin
      ;;
    Linux)
      host_os=linux
      ;;
    *)
      kube::log::error "Unsupported host OS.  Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}

host::arch() {
  local host_arch
  case "$(uname -m)" in
    x86_64*)
      host_arch=amd64
      ;;
    i?86_64*)
      host_arch=amd64
      ;;
    amd64*)
      host_arch=amd64
      ;;
    aarch64*)
      host_arch=arm64
      ;;
    arm64*)
      host_arch=arm64
      ;;
    arm*)
      host_arch=arm
      ;;
    ppc64le*)
      host_arch=ppc64le
      ;;
    *)
      kube::log::error "Unsupported host arch. Must be x86_64, arm, arm64, or ppc64le."
      exit 1
      ;;
  esac
  echo "${host_arch}"
}


# Installs kind and helm dependencies locally.
# Required envs:
#  - HELM_VERSION
#  - INSTALL_DIR
#
# usage: env INSTALL_DIR=/tmp HELM_VERSION=v2.14.3 install::local::kind
install::local::helm() {
    mkdir -p "${INSTALL_DIR}/bin"
    export PATH="${INSTALL_DIR}/bin:${PATH}"

    pushd "${INSTALL_DIR}"

    shout "- Install helm ${HELM_VERSION} locally to a tempdir..."
    curl -LO https://git.io/get_helm.sh > ${INSTALL_DIR}/get_helm.sh
    chmod 700 ${INSTALL_DIR}/get_helm.sh
    env HELM_INSTALL_DIR="${INSTALL_DIR}/bin" ./get_helm.sh \
        --version ${HELM_VERSION} \
        --no-sudo

    popd
}

# Installs tiller on cluster
install::cluster::tiller() {
    shout "- Installing Tiller..."
    kubectl create -f ${LIB_DIR}/../assets/tiller-rbac.yaml
    helm init --service-account tiller --wait
}

# Installs Service Catalog from newest 0.2.x release on k8s cluster.
# Required envs:
#  - SC_CHART_NAME
#  - SC_NAMESPACE
install::cluster::service_catalog_v2() {
    shout "- Installing Service Catalog in version 0.2.x"
    helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
    # install always the newest service catalog with apiserver
    helm install svc-cat/catalog-v0.2 --name ${SC_CHART_NAME} --namespace ${SC_NAMESPACE} --wait
}

#
# 'kind'(kubernetes-in-docker) functions
#
readonly KIND_CLUSTER_NAME="kind-ci"

kind::create_cluster() {
    shout "- Create k8s cluster..."
    kind create cluster --name=${KIND_CLUSTER_NAME} --image=kindest/node:${KUBERNETES_VERSION} --wait=5m
    export KUBECONFIG="$(kind get kubeconfig-path --name=${KIND_CLUSTER_NAME})"
}

kind::delete_cluster() {
    kind delete cluster --name=${KIND_CLUSTER_NAME}
}

# Arguments:
#   $1 - image name to copy into cluster nodes
kind::load_image() {
    kind load docker-image $1 --name=${KIND_CLUSTER_NAME}
}
