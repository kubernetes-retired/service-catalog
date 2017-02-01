#!/bin/bash
# Copyright 2016 The Kubernetes Authors.
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

. "${ROOT}/contrib/jenkins/run_utilities.sh" || { echo 'Cannot load run utilities.'; exit 1; }

while [[ $# -gt 0 ]]; do
  case "${1}" in
    --project)    PROJECT="${2:-}"; shift ;;
    --kubeconfig) KUBECONFIG="${2:-}"; shift ;;
    --namespace)  NAMESPACE="${2:-}"; shift ;;
    --version)    VERSION="${2:-}"; shift ;;
    --gcr)        GCR="${2:-}"; shift ;;

    *) error_exit "Unrecognized command line parameter: $1" ;;
  esac
  shift
done

# kubectl accepts kubeconfig on the command line but HELM doesn't
# so we need to export the variable :(
[[ -n "${KUBECONFIG:-}" ]] \
  && export KUBECONFIG

[[ -n "${PROJECT:-}" ]] \
  || error_exit "Missing required --project parameter"

[[ -n "${NAMESPACE:-}" ]] \
  || error_exit "Missing required --namespace parameter"

function print_logs() {
  # Print logs for controller & broker on error
  if [[ $? -ne 0 ]]; then
    local pods_list="$(kubectl get pods --namespace "${NAMESPACE}")"
    local controller_pod="$(echo ${pods_list} | sed 's/.*\(controller[a-z0-9\-]*\).*/\1/')"
    local broker_pod="$(echo ${pods_list} | sed 's/.*\(k8s-broker[a-z0-9\-]*\).*/\1/')"

    echo '#### CONTROLLER LOGS ####'
    kubectl logs --namespace "${NAMESPACE}" "${controller_pod}"
    echo '#### BROKER LOGS ####'
    kubectl logs --namespace "${NAMESPACE}" "${broker_pod}"
  fi
}

function wait_for_service_host() {
    if [[ "${#}" -ne 2 ]]; then
      echo 'Unexpected number of arguments passed.'
      return 1
    fi

    local name="${1}"
    local path="${2}"
    local services_output="$(kubectl get services --namespace "${NAMESPACE}" | xargs echo -n)"
    local host="$(echo "${services_output}" | sed "s/.*${name} [0-9.]* \([0-9.]*\) \([0-9]*\).*/\1:\2/")"

    retry -n 10 -s 10 -t 60 \
        curl --silent --fail "http://${host}/${path}" > /dev/null \
      && return 0

    echo "Could not communicate with ${host}/${path}. Response: $(curl "http://${host}/${path}")"
    return 1
}

GCR="${GCR:-gcr.io/${PROJECT}/catalog}"
VERSION="${VERSION:-"$(git describe --tags --always --abbrev=7 --dirty)"}" \
  || error_exit 'Cannot determine Git commit SHA'

# Deploy to Kubernetes cluster
echo 'Deploying to Kubernetes cluster...'

# Create the namespace
kubectl create namespace "${NAMESPACE}"

retry -n 10 -s 10 -t 60 \
    helm install "${ROOT}/deploy/catalog" \
    --set "registry=${GCR},version=${VERSION},storageType=etcd,debug=true" \
    --namespace "${NAMESPACE}" \
  || error_exit 'Error deploying to Kubernetes cluster.'

# CREATE SERVICES

trap print_logs EXIT

# Wait for pods to be up
echo 'Waiting on services to spin up...'

wait_for_expected_output -x -e 'ContainerCreating' -n 20 -s 10 -t 60 \
    kubectl get pods --namespace "${NAMESPACE}" \
  || error_exit 'Services took an unexpected amount of time to spin up.'

kubectl get pods --namespace "${NAMESPACE}" --no-headers  | grep -v Running \
  && error_exit 'Pods failed to spin up successfully.'

# Wait for services to respond
echo 'Waiting on services to get external IP...'

wait_for_expected_output -x -e 'pending' -n 20 -s 10 -t 60 \
    kubectl get services --namespace "${NAMESPACE}" \
  || error_exit 'Services took an unexpected amount of time to spin up.'

echo 'Waiting on services to respond...'

wait_for_service_host 'registry' 'services' \
  || error_exit 'Error when trying to communicate with registry.'

wait_for_service_host 'k8s-broker' 'v2/catalog' \
  || error_exit 'Error when trying to communicate with k8s broker.'

wait_for_service_host 'ups-broker' 'v2/catalog' \
  || error_exit 'Error when trying to communicate with ups broker.'

#wait_for_service_host 'controller' 'v2/service_brokers' \
#  || error_exit 'Error when trying to communicate with controller.'

echo 'Creating resources...'

kubectl create -f "${ROOT}/contrib/examples/walkthrough/broker.yaml" \
  || error_exit 'Cannot create brokers.'

wait_for_expected_output -e 'booksbe' -n 20 -s 1 -t 60 \
    kubectl get serviceclasses \
  || error_exit 'Could not retrieve service classes from broker.'

kubectl create -f "${ROOT}/contrib/examples/walkthrough/backend.yaml" \
  || error_exit 'Cannot create backend.'

sleep 10

kubectl create -f "${ROOT}/contrib/examples/walkthrough/binding.yaml" \
  || error_exit 'Cannot create binding.'

sleep 10

kubectl create -f "${ROOT}/contrib/examples/user-bookstore-client/bookstore.yaml" \
  || error_exit 'Cannot create frontend.'

wait_for_expected_output -e 'user-bookstore-fe' -n 20 -s 2 -t 60 \
    kubectl get services \
  || error_exit 'Frontend service took unexpected amount of time to come up.'

echo 'Waiting for external IP for frontend...'
wait_for_expected_output -x -e 'pending' -n 20 -s 10 -t 60 \
    kubectl get services \
  || error_exit 'Frontend service took unexpected amount of time to get external IP.'

IP="$(kubectl get services | xargs echo -n | sed 's/.*user-bookstore-fe [0-9.]* \([0-9.]*\).*/\1/')"

echo 'Waiting for frontend service to unblock...'
wait_for_expected_output -x -e 'blocked' -n 20 -s 30 -t 60 \
    curl "http://${IP}:8080/shelves" \
  || error_exit 'Access to frontend service still blocked after unexpected amount of time.'

echo 'Deployment to Kubernetes cluster succeeded.'
