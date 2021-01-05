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

# This script execute such test flow:
#
# 1. Provision testing environment using 'kind'(kubernetes-in-docker)
# 2. Install Service Catalog in version 0.2.x
# 3. Create a sample resources (broker, instances, bindings etc)
# 4. Upgrade Service Catalog to version build from sources
# 5. Execute tests to check if everything still working
#

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

SC_CHART_NAME="catalog"
export SC_NAMESPACE="catalog"
export SC_APISERVER="${SC_CHART_NAME}-catalog-apiserver"
export SC_CONTROLLER="${SC_CHART_NAME}-catalog-controller-manager"

TB_CHART_NAME="test-broker"
export TB_NAME="${TB_CHART_NAME}-test-broker"
export TB_NAMESPACE="test-broker"

DUMP_CLUSTER_INFO="${DUMP_CLUSTER_INFO:-false}"

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

upgrade::cluster::service_catalog() {
    shout "- Building Service Catalog image from sources..."
    pushd "${REPO_ROOT_DIR}"

    make service-catalog-image

    shout "- Load Service Catalog image into cluster..."
    kind::load_image service-catalog:canary

    shout "- Upgrade Service Catalog..."
    helm upgrade ${SC_CHART_NAME} charts/catalog \
        --set imagePullPolicy=IfNotPresent \
        --set image=service-catalog:canary \
        --namespace ${SC_NAMESPACE} \
        --wait
    popd
}

examiner::prepare_resources() {
    shout "- Building Test Broker image from sources..."
    pushd "${REPO_ROOT_DIR}"

    make test-broker-image

    shout "- Load Service Catalog image into cluster..."
    kind::load_image test-broker:canary

    shout "- Installing Test broker..."
    helm install "$TB_CHART_NAME" charts/test-broker \
        --set imagePullPolicy=IfNotPresent \
        --set image=test-broker:canary \
        --namespace ${TB_NAMESPACE} \
        --create-namespace \
        --wait

    shout "- Create sample resources for testing purpose..."
    go run "${REPO_ROOT_DIR}"/test/upgrade/examiner/main.go --action prepareData
}

examiner::execute_test() {
    shout "- Execute upgrade tests..."
    # Required environment variables:
    # SC_APISERVER, SC_CONTROLLER, SC_NAMESPACE, TB_NAME, TB_NAMESPACE
    go run "${REPO_ROOT_DIR}/test/upgrade/examiner/main.go" --action executeTests
    pushd "${REPO_ROOT_DIR}"
}

main() {
    shout "Starting migration test"

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

    install::cluster::service_catalog_v2

    examiner::prepare_resources

    upgrade::cluster::service_catalog

    examiner::execute_test

    # Test completed successfully. We do not have to dump cluster info
    DUMP_CLUSTER_INFO=false
    shout "Migration test completed successfully."
}

main
