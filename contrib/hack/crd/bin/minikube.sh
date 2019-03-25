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

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCES_DIR="${CURRENT_DIR}/../resources"

MINIKUBE_VERSION=0.33.0
KUBERNETES_VERSION=1.11.5
KUBECTL_CLI_VERSION=1.11.0
VM_DRIVER=hyperkit
DISK_SIZE=20g
MEMORY=8192

#TODO refactor to use minikube status!
function waitForMinikubeToBeUp() {
    set +o errexit

    echo "Waiting for minikube to be up..."

	LIMIT=15
    COUNTER=0

    while [ ${COUNTER} -lt ${LIMIT} ] && [ -z "$STATUS" ]; do
      (( COUNTER++ ))
      echo -e "Keep calm, there are $LIMIT possibilities and so far it is attempt number $COUNTER"
      STATUS="$(kubectl get namespaces || :)"
      sleep 1
    done

    # In case apiserver is not available get minikube logs
    if [[ -z "$STATUS" ]] && [[ "$VM_DRIVER" = "none" ]]; then
      cat /var/lib/minikube/minikube.err
    fi

    set -o errexit

    echo "Minikube is up"
}

function increaseFsInotifyMaxUserInstances() {
    # Default value of 128 is not enough to perform “kubectl log -f” from pods, hence increased to 524288
    if [[ "$VM_DRIVER" != "none" ]]; then
        minikube ssh -- "sudo sysctl -w fs.inotify.max_user_instances=524288"
        echo "fs.inotify.max_user_instances is increased"
    fi
}

function applyDefaultRbacRole() {
    kubectl apply -f "${RESOURCES_DIR}/default-sa-rbac-role.yaml"
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
    --disk-size=$DISK_SIZE \
    --bootstrapper=kubeadm

    waitForMinikubeToBeUp

    increaseFsInotifyMaxUserInstances

}

start
