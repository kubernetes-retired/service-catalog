#!/bin/bash
# Copyright 2017 The Kubernetes Authors.
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

set -o nounset
set -o errexit
set -x

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

. "${ROOT}/contrib/hack/utilities.sh" || { echo 'Cannot load bash utilities.'; exit 1; }

while [[ $# -gt 0 ]]; do
  case "${1}" in
    --registry)   REGISTRY="${2:-}"; shift ;;
    --version)    VERSION="${2:-}"; shift ;;
    --with-tpr)   WITH_TPR=1 ;;
    --cleanup)    CLEANUP=1 ;;
    *) error_exit "Unrecognized command line parameter: $1" ;;
  esac
  shift
done

BROKER_RELEASE="ups-broker"
CATALOG_RELEASE="catalog"

K8S_KUBECONFIG="${KUBECONFIG:-~/.kube/config}"
SC_KUBECONFIG="/tmp/sc-kubeconfig"

function cleanup() {
  echo 'Cleaning up resources...'
  export KUBECONFIG="${K8S_KUBECONFIG}"
  helm delete --purge "${BROKER_RELEASE}" || true
  helm delete --purge "${CATALOG_RELEASE}" || true
  rm -f "${SC_KUBECONFIG}"
  kubectl delete secret -n test-ns my-secret > /dev/null 2>&1 || true
  kubectl delete namespace test-ns || true

  wait_for_expected_output -x -e 'test-ns' -n 10 \
    kubectl get namespaces

  # TODO: Hack in order to delete TPRs. Will need to be removed when TPRs can be deleted
  # by the catalog API server.
  if [[ -n "${WITH_TPR:-}" ]]; then
    kubectl delete thirdpartyresources binding.servicecatalog.k8s.io
    kubectl delete thirdpartyresources instance.servicecatalog.k8s.io
    kubectl delete thirdpartyresources broker.servicecatalog.k8s.io
    kubectl delete thirdpartyresources service-class.servicecatalog.k8s.io
  fi

  echo 'Cleanup done.'
}

VERSION="${VERSION:-"$(git describe --tags --always --abbrev=7 --dirty)"}" \
  || error_exit 'Cannot determine Git commit SHA'

# Deploying to cluster

if [[ -n "${CLEANUP:-}" ]]; then
  trap cleanup EXIT
fi

kubectl create namespace test-ns

echo 'Deploying user-provided-service broker...'

VALUES="version=${VERSION}"
if [[ -n "${REGISTRY:-}" ]]; then
  VALUES+=",registry=${REGISTRY}"
fi

retry -n 10 \
    helm install "${ROOT}/charts/ups-broker" \
    --name "${BROKER_RELEASE}" \
    --namespace "ups-broker" \
    --set "${VALUES}" \
  || error_exit 'Error deploy ups broker to cluster.'

echo 'Deploying service catalog...'

VALUES+=',debug=true'
VALUES+=',insecure=true'
VALUES+=',apiserver.service.type=LoadBalancer'
if [[ -n "${WITH_TPR:-}" ]]; then
  VALUES+=',apiserver.storage.type=tpr'
  VALUES+=',apiserver.storage.tpr.globalNamespace=test-ns'
fi

NAME="$(openssl rand -base64 100 | tr -dc a-z | cut -c -6)"

retry -n 10 \
    helm install "${ROOT}/charts/catalog" \
    --name "${CATALOG_RELEASE}" \
    --namespace "catalog" \
    --set "${VALUES}" \
  || error_exit 'Error deploying service catalog to cluster.'

# Waiting for everything to come up

echo 'Waiting on pods to come up...'

wait_for_expected_output -x -e 'ContainerCreating' -n 10 \
    kubectl get pods --namespace ups-broker \
  || error_exit 'User provided service broker pod took an unexpected amount of time to come up.'

[[ "$(kubectl get pods --namespace ups-broker | grep ups-broker-ups-broker | awk '{print $3}')" == 'Running' ]] \
  || error_exit 'User provided service broker pod did not come up successfully.'

wait_for_expected_output -x -e 'ContainerCreating' -n 10 \
    kubectl get pods --namespace catalog \
  || error_exit 'Service catalog pods did not come up successfully.'

[[ "$(kubectl get pods --namespace catalog | grep catalog-catalog-apiserver | awk '{print $3}')" == 'Running' ]] \
  || error_exit 'API server pod did not come up successfully.'
[[ "$(kubectl get pods --namespace catalog | grep catalog-catalog-controller | awk '{print $3}')" == 'Running' ]] \
  || error_exit 'Controller pod did not come up successfully.'

echo 'Waiting on external IP for service catalog API Server...'

wait_for_expected_output -x -e 'pending' -n 10 \
    kubectl get services --namespace catalog \
  || error_exit 'Could not get external IP for service catalog API Server.'

# Create kubeconfig for service catalog API server

echo 'Connecting to service catalog API Server...'

API_SERVER_HOST="$(kubectl get services -n catalog | grep 'apiserver' | awk '{print $3}')"

[[ "${API_SERVER_HOST}" =~ ^[0-9.]*$ ]] \
  || error_exit 'Error when fetching service catalog API Server IP address.'

export KUBECONFIG="${SC_KUBECONFIG}"

kubectl config set-credentials service-catalog-creds --username=admin --password=admin
kubectl config set-cluster service-catalog-cluster --server="http://${API_SERVER_HOST}:80"
kubectl config set-context service-catalog-ctx --cluster=service-catalog-cluster --user=service-catalog-creds
kubectl config use-context service-catalog-ctx

retry -n 10 \
  kubectl get brokers,serviceclasses,instances,bindings \
  || error_exit 'Issue listing resources from service catalog API server.'

# Create the broker

echo 'Creating broker...'

kubectl create -f "${ROOT}/contrib/examples/walkthrough/ups-broker.yaml" \
  || error_exit 'Error when creating ups-broker.'

wait_for_expected_output -e 'FetchedCatalog' -n 10 \
    kubectl get brokers ups-broker -o yaml \
  || error_exit 'Did not receive expected condition when creating ups-broker.'

[[ "$(kubectl get brokers ups-broker -o yaml)" == *"status: \"True\""* ]] \
  || error_exit 'Failure status reported when attempting to fetch catalog from ups-broker.'

[[ "$(kubectl get serviceclasses)" == *user-provided-service* ]] \
  || error_exit 'user-provided-service not listed when fetching service classes.'

# Provision an instance

echo 'Provisioning instance...'

kubectl create -f "${ROOT}/contrib/examples/walkthrough/ups-instance.yaml" \
  || error_exit 'Error when creating ups-instance.'

wait_for_expected_output -e 'ProvisionedSuccessfully' -n 10 \
  kubectl get instances -n test-ns ups-instance -o yaml \
  || error_exit 'Did not receive expected condition when provisioning ups-instance.'

[[ "$(kubectl get instances -n test-ns ups-instance -o yaml)" == *"status: \"True\""* ]] \
  || error_exit 'Failure status reported when attempting to provision ups-instance.'

# Bind to the instance

echo 'Binding to instance...'

kubectl create -f "${ROOT}/contrib/examples/walkthrough/ups-binding.yaml" \
  || error_exit 'Error when creating ups-binding.'

wait_for_expected_output -e 'InjectedBindResult' -n 10 \
  kubectl get bindings -n test-ns ups-binding -o yaml \
  || error_exit 'Did not receive expected condition when injecting ups-binding.'

[[ "$(kubectl get bindings -n test-ns ups-binding -o yaml)" == *"status: \"True\""* ]] \
  || error_exit 'Failure status reported when attempting to inject ups-binding.'

[[ "$(KUBECONFIG="${K8S_KUBECONFIG}" kubectl get secrets -n test-ns)" == *my-secret* ]] \
  || error_exit '"my-secret" not present when listing secrets.'


# TODO: Cannot currently test TPR deletion; only delete if using an etcd-backed
# API server
if [[ -z "${WITH_TPR:-}" ]]; then
  # Unbind from the instance

  echo 'Unbinding from instance...'

  kubectl delete -n test-ns bindings ups-binding \
    || error_exit 'Error when deleting ups-binding.'

  export KUBECONFIG="${K8S_KUBECONFIG}"
  wait_for_expected_output -x -e "my-secret" -n 10 \
      kubectl get secrets -n test-ns \
    || error_exit '"my-secret" not removed upon deleting ups-binding.'
  export KUBECONFIG="${SC_KUBECONFIG}"

  # Deprovision the instance

  echo 'Deprovisioning instance...'

  kubectl delete -n test-ns instances ups-instance \
    || error_exit 'Error when deleting ups-instance.'

  # Delete the broker

  echo 'Deleting broker...'

  kubectl delete brokers ups-broker \
    || error_exit 'Error when deleting ups-broker.'

  wait_for_expected_output -x -e 'user-provided-service' -n 10 \
      kubectl get serviceclasses \
    || error_exit 'Service classes not successfully removed upon deleting ups-broker.'

  [[ "$(kubectl get serviceclasses 2>&1)" == "No resources found." ]] \
    || error_exit 'Service classes not successfully removed upon deleting ups-broker.'
fi

echo 'Walkthrough completed successfully.'
