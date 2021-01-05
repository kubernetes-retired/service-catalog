#!/usr/bin/env bash
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


# This script provisions testing environment using 'kind'(kubernetes-in-docker)
# and execute end-to-end Service Catalog tests.
#
# It requires Docker to be installed.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=${CURRENT_DIR}/../../../
readonly TMP_DIR=$(mktemp -d)

source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
source "${CURRENT_DIR}/lib/deps_ver.sh" || { echo 'Cannot load dependencies versions.'; exit 1; }

SKIP_DEPS_INSTALLATION=${SKIP_DEPS_INSTALLATION:-}

DUMP_CLUSTER_INFO="${DUMP_CLUSTER_INFO:-false}"

SC_CHART_NAME="catalog"
SC_NAMESPACE="catalog"

export GOFLAGS=-mod=vendor

cleanup() {
    if [[ "${DUMP_CLUSTER_INFO}" == true ]]; then
        shout '- Creating artifacts...'

        export DUMP_NAMESPACE=${SC_NAMESPACE}
        dump_logs || true
    fi

    kind::delete_cluster || true

    rm -rf "${TMP_DIR}" || true
}

trap cleanup EXIT

install::cluster::service_catalog_latest() {
    pushd "${REPO_ROOT_DIR}"
    shout "- Building Service Catalog image from sources..."
    env REGISTRY="" VERSION=canary ARCH=amd64 \
        make service-catalog-image

    shout "- Loading Service Catalog image into cluster..."
    kind::load_image service-catalog:canary

    shout "- Installing Service Catalog via helm chart from sources..."
    helm install ${SC_CHART_NAME} charts/catalog \
        --set imagePullPolicy=IfNotPresent \
        --set image=service-catalog:canary \
        --namespace=${SC_NAMESPACE} \
        --create-namespace \
        --wait
    popd
}

test::prepare_data() {
    shout "- Building User Broker image from sources..."
    pushd "${REPO_ROOT_DIR}"
    env REGISTRY="" VERSION=canary ARCH=amd64 \
        make user-broker-image
    popd

    shout "- Load User Broker image into cluster..."
    kind::load_image user-broker:canary
}

test::execute() {
    shout "- Executing e2e test..."
    pushd "${REPO_ROOT_DIR}/test/e2e/"
    env SERVICECATALOGCONFIG="${KUBECONFIG}" go test -v ./... -broker-image="user-broker:canary"
    popd
}

main() {
    shout "Starting E2E test."

    if [[ "${SKIP_DEPS_INSTALLATION}" == "" ]]; then
        export INSTALL_DIR=${TMP_DIR} KIND_VERSION=${STABLE_KIND_VERSION} HELM_VERSION=${STABLE_HELM_VERSION}
        install::local::kind
        install::local::helm
    else
        echo "Skipping kind and helm installation cause SKIP_DEPS_INSTALLATION is set to true."
    fi

    export KUBERNETES_VERSION=${KUBERNETES_VERSION:-${STABLE_KUBERNETES_VERSION}} KUBECONFIG="${TMP_DIR}/kubeconfig"
    kind::create_cluster

    # Cluster is already created, and all below operation are performed against that cluster,
    # so we should dump cluster info for debugging purpose in case of any error
    DUMP_CLUSTER_INFO=true

    install::cluster::service_catalog_latest

    test::prepare_data
    test::execute

    # Test completed successfully. We do not have to dump cluster info
    DUMP_CLUSTER_INFO=false
    shout "E2E test completed successfully."
}

main
