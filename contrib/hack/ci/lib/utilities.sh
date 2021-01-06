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

# Ensure the go tool exists and is a viable version.
golang::verify_go_version() {
  shout "Verify Go version"
  if [[ -z "$(command -v go)" ]]; then
    echo "Can't find 'go' in PATH, please fix and retry.
See http://golang.org/doc/install for installation instructions."
    exit 1
  fi

  local go_version
  IFS=" " read -ra go_version <<< "$(go version)"
  local minimum_go_version
  minimum_go_version=go1.13
  if [[ "${minimum_go_version}" != $(echo -e "${minimum_go_version}\n${go_version[2]}" | sort -s -t. -k 1,1 -k 2,2n -k 3,3n | head -n1) && "${go_version[2]}" != "devel" ]]; then
    echo "Detected go version: ${go_version[*]}.
Kubernetes requires ${minimum_go_version} or greater.
Please install ${minimum_go_version} or later."
    exit 1
  fi
}

# Checks whether jq is installed.
require-jq() {
  if ! command -v jq &>/dev/null; then
    echo "jq not found. Please install." 1>&2
    return 1
  fi
}

dump_logs() {
    # The $ARTIFACTS environment variable is set by prow.
    # It's a directory where job artifacts can be dumped for automatic upload to GCS upon job completion.
    # source: https://github.com/kubernetes/test-infra/blob/2ccde6c957e5de7603faa43399167b18a41b496b/prow/pod-utilities.md#what-the-test-container-can-expect
    LOGS_DIR=${ARTIFACTS:-${TMP_DIR}}/logs
    mkdir -p ${LOGS_DIR}

    echo "Dumping logs from namespace ${DUMP_NAMESPACE} into ${LOGS_DIR}"
    kubectl cluster-info dump --namespace=${DUMP_NAMESPACE} --output-directory=${LOGS_DIR}
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
      echo "Unsupported host OS. Must be Linux or Mac OS X."
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
      echo "Unsupported host arch. Must be x86_64, arm, arm64, or ppc64le."
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
      curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
      chmod 700 ./get_helm.sh
      env HELM_INSTALL_DIR="${INSTALL_DIR}/bin" ./get_helm.sh \
          --version "${HELM_VERSION}" \
          --no-sudo
    popd
}

# Installs Service Catalog from newest 0.2.x release on k8s cluster.
# Required envs:
#  - SC_CHART_NAME
#  - SC_NAMESPACE
install::cluster::service_catalog_v2() {
    shout "- Installing Service Catalog in version 0.2.x"
    helm repo add svc-cat https://kubernetes-sigs.github.io/service-catalog
    # always install the newest service catalog with apiserver
    helm repo update
    helm install "$SC_CHART_NAME" svc-cat/catalog-v0.2 --namespace "$SC_NAMESPACE" --create-namespace --wait
}

#
# 'kind'(kubernetes-in-docker) functions
#
readonly KIND_CLUSTER_NAME="kind-ci"

kind::create_cluster() {
    shout "- Create k8s cluster..."
    kind create cluster --name=${KIND_CLUSTER_NAME} --image="kindest/node:${KUBERNETES_VERSION}" --wait=5m
}

kind::delete_cluster() {
    kind delete cluster --name=${KIND_CLUSTER_NAME}
}

# Arguments:
#   $1 - image name to copy into cluster nodes
kind::load_image() {
    kind load docker-image $1 --name=${KIND_CLUSTER_NAME}
}
