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

set -ux

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

. "${ROOT}/hack/run_utilities.sh" || { echo 'Cannot load run utilities.'; exit 1; }

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

trap print_logs EXIT

GCR="${GCR:-gcr.io/${PROJECT}/catalog}"
VERSION="${VERSION:-$(git rev-parse --verify HEAD)}" \
  || error_exit 'Cannot determine Git commit SHA'

# Deploy to Kubernetes cluster
echo 'Deploying to Kubernetes cluster...'

# Create the namespace
kubectl create namespace "${NAMESPACE}"

retry -n 10 -s 10 -t 60 \
    helm install "${ROOT}/deploy/catalog" \
    --set "registry=${GCR},version=${VERSION}" \
  || error_exit 'Error deploying to Kubernetes cluster.'

echo 'Waiting on services to spin up...'

# CREATE SERVICES
wait_for_expected_output -x -e 'ContainerCreating' -n 20 -s 10 -t 60  \
    kubectl get pods --namespace ${NAMESPACE} \
  || error_exit 'Services took an unexpected amount of time to spin up.'

# Check to see if Kubernetes resources have finished being created.
retry kubectl get serviceclasses,serviceinstances,servicebrokers,servicebindings \
  || error_exit 'Kubernetes resources took an unexpected amount of time to spin up.'

echo 'Creating resources...'

kubectl create -f "${ROOT}/examples/walkthrough/broker.yaml" \
  || error_exit 'Cannot create broker.'

sleep 10 #TODO: check that the broker actually came up.

kubectl create -f "${ROOT}/examples/walkthrough/backend.yaml" \
  || error_exit 'Cannot create backend.'

sleep 10

kubectl create -f "${ROOT}/examples/walkthrough/binding.yaml" \
  || error_exit 'Cannot create binding.'

sleep 10

kubectl create -f "${ROOT}/examples/walkthrough/frontend.yaml" \
  || error_exit 'Cannot create frontend.'

echo 'Waiting for frontend service to come up...'
wait_for_expected_output -e 'booksfe' -n 20 -s 10 -t 60 \
    kubectl get services \
  || error_exit 'Frontend service took unexpected amount of time to come up.'

echo 'Waiting for external IP for frontend...'
wait_for_expected_output -x -e 'pending' -n 20 -s 10 -t 60 \
    kubectl get services \
  || error_exit 'Frontend service took unexpected amount of time to get external IP.'

IP=$(echo $(kubectl get services) | sed 's/.*booksfe [0-9.]* \([0-9.]*\).*/\1/')

echo 'Waiting for frontend service to unblock...'
wait_for_expected_output -x -e 'blocked' -n 20 -s 30 -t 60 \
    curl "http://${IP}:8080/shelves" \
  || error_exit 'Access to frontend service still blocked after unexpected amount of time.'

echo 'Deployment to Kubernetes cluster succeeded.'
