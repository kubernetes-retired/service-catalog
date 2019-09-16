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

SC_CHART_NAME="catalog"
SC_NAMESPACE="catalog"

cleanup() {
    kind::delete_cluster || true
    rm -rf "${TMP_DIR}" > /dev/null 2>&1 || true
}

trap cleanup EXIT

install::cluster::service_catalog_latest() {
    pushd ${REPO_ROOT_DIR}
    shout "- Building Service Catalog image from sources..."
    env REGISTRY="" VERSION=canary ARCH=amd64 \
        make service-catalog-image

    shout "- Loading Service Catalog image into cluster..."
    kind::load_image service-catalog:canary

    shout "- Installing Service Catalog via helm chart from sources..."
    helm install charts/catalog \
        --set imagePullPolicy=IfNotPresent \
        --set image=service-catalog:canary \
        --namespace=${SC_NAMESPACE} \
        --name=${SC_CHART_NAME} \
        --wait
    popd
}

test::prepare_data() {
    shout "- Building User Broker image from sources..."
    pushd ${REPO_ROOT_DIR}
    env REGISTRY="" VERSION=canary ARCH=amd64 \
        make user-broker-image
    popd

    shout "- Load User Broker image into cluster..."
    kind::load_image user-broker:canary
}

test::execute() {
    shout "- Executing e2e test..."
    pushd ${REPO_ROOT_DIR}/test/e2e/
    env SERVICECATALOGCONFIG="${KUBECONFIG}" go test -v ./... -broker-image="user-broker:canary"
    popd
}

exit_and_dump_logs_for_failed_test() {
    # The $ARTIFACTS environment variable is set by prow.
    # It's a directory where job artifacts can be dumped for automatic upload to GCS upon job completion.
    # source: https://github.com/kubernetes/test-infra/blob/2ccde6c957e5de7603faa43399167b18a41b496b/prow/pod-utilities.md#what-the-test-container-can-expect
    LOGS_DIR=${ARTIFACTS:-${TMP_DIR}}/logs
    mkdir -p ${LOGS_DIR}

    echo "Executing test failed, dumping logs from namespace ${SC_NAMESPACE} into ${LOGS_DIR}"
    env DUMP_NAMESPACE= OUTPUT_DIR= dump_k8s_resources

    kubectl cluster-info dump --namespace=${SC_NAMESPACE} --output-directory=${LOGS_DIR}
    exit 1
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

    kind::create_cluster

    install::cluster::tiller
    install::cluster::service_catalog_latest

    test::prepare_data
    test::execute \
        || exit_and_dump_logs_for_failed_test

    shout "E2E test completed successfully."
}

main
