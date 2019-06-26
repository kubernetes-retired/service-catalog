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

set -o errexit

KUBERNETES_VERSION=1.11.5
VM_DRIVER=hyperkit
MEMORY=6144

function waitForMinikubeToBeUp() {
    echo "Waiting for minikube to be up..."

	LIMIT=15
    COUNTER=0

    while [ ${COUNTER} -lt ${LIMIT} ]; do
      if checkMinikubeStatus $1; then
        echo "Minikube is up"
        return 0
      else
        echo "Minikube is not ready"
      fi
      (( COUNTER++ ))
      echo -e "Keep calm, there are $LIMIT possibilities and so far it is attempt number $COUNTER"
      sleep 1
    done

    set +o errexit

    # In case apiserver is not available get minikube logs
    if [[ "$VM_DRIVER" = "none" ]]; then
      cat /var/lib/minikube/minikube.err
    fi

    set -o errexit
}

function checkMinikubeStatus() {
    MINIKUBE_STATUS="$(minikube status --format {{.Host}})"
    KUBELET_STATUS="$(minikube status --format {{.Kubelet}})"
    APISERVER_STATUS="$(minikube status --format {{.ApiServer}})"

    if [[ "$MINIKUBE_STATUS" == "Running" ]] &&
       [[ "$KUBELET_STATUS" == "Running" ]] &&
       [[ "$APISERVER_STATUS" == "Running" ]]; then
        return 0
    else
        return 1
    fi
}

function start() {
    minikube start \
    --memory $MEMORY \
    --cpus 4 \
    --extra-config=apiserver.authorization-mode=RBAC \
    --extra-config=apiserver.cors-allowed-origins="http://*" \
    --extra-config=apiserver.enable-admission-plugins="DefaultStorageClass,LimitRanger,MutatingAdmissionWebhook,NamespaceExists,NamespaceLifecycle,ResourceQuota,ServiceAccount,ValidatingAdmissionWebhook" \
    --kubernetes-version=v$KUBERNETES_VERSION \
    --vm-driver=$VM_DRIVER \
    --bootstrapper=kubeadm

    waitForMinikubeToBeUp
}

start
