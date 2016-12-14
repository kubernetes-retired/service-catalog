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

set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

. "${ROOT}/hack/utilities.sh" || { echo 'Cannot load Bash utilities'; exit 1; }

# Cleanup services in Kubernetes to prevent network resource leaking
function wipe_cluster() {
  local namespace
  for namespace in $(kubectl get namespaces -oname | grep -v kube-system); do
    namespace="${namespace##*/}"
    kubectl delete deployments,services,configmaps,pods,replicasets \
        --all --namespace "${namespace}"

    wait_for_expected_output -x -e 'Terminating' -n 20 -s 2 -t 60 \
        kubectl get pods --namespace "${namespace}" \
      || echo "WARNING: Some Kubernetes resources in namespace "${namespace}" failed to terminate."

    if [[ "${namespace}" != "default" ]]; then
      kubectl delete namespace "${namespace}"
    fi
  done

  kubectl delete serviceinstances,serviceclasses,servicebindings,servicebrokers --all #TODO: Eventually this should work.

  # Temporarily, delete all by name.
  kubectl delete serviceinstances backend frontend
  kubectl delete serviceclasses booksbe user-provided-service
  kubectl delete servicebindings database
  kubectl delete servicebrokers k8s ups

  return 0
}
