#!/usr/bin/env bash
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

set -o errexit

EXIT_CODE=0

function checkIfClassExist(){
    local externalName=$1
    local classesNames=$2

    for class in ${classesNames}
    do
        if [[ "${class}" = "${externalName}" ]]; then
            return 0
        fi
    done

    return 1
}

function checkIfClassExistForInstance(){
    serviceClassesNames=$(kubectl get serviceclasses --all-namespaces -ojsonpath="{.items[*].spec.externalName}")
    clusterServiceClassesNames=$(kubectl get clusterserviceclasses -ojsonpath="{.items[*].spec.externalName}")

    for className in $(kubectl get serviceinstances --all-namespaces -ojsonpath="{.items[*].status.userSpecifiedClassName}")
    do
        type=$(echo "${className}" | cut -d'/' -f1)
        externalName=$(echo "${className}" | cut -d'/' -f2)

        set +o errexit
        if [[ "${type}" = "ClusterServiceClass" ]]; then
            checkIfClassExist "${externalName}" "${clusterServiceClassesNames}"
            if [[ $? -eq 1 ]]; then
                echo "${className} not exist in the cluster for the ServiceInstances:"
                kubectl get serviceinstances --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,STATUS:.status.provisionStatus,CLASS\ NAME:.status.userSpecifiedClassName | grep "${className}"
                setExitCode
            fi
        fi
        if [[ "${type}" = "ServiceClass" ]]; then
            checkIfClassExist "${externalName}" "${serviceClassesNames}"
            if [[ $? -eq 1 ]]; then
                echo "${className} not exist in the cluster for the ServiceInstances:"
                kubectl get serviceinstances --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,STATUS:.status.provisionStatus,CLASS\ NAME:.status.userSpecifiedClassName | grep "${className}"
                setExitCode
            fi
        fi
    done
    set -o errexit
}

function setExitCode() {
    echo
    EXIT_CODE=1
}
# checks if there are some instances with not existing classes
checkIfClassExistForInstance

#
# Check if any class/plans were removed from broker's catalog
#
CSC=$(kubectl get clusterserviceclasses -ojsonpath="{.items[*].metadata.deletionTimestamp}")
if [[ -n ${CSC} ]]; then
    echo "There are being deleted ClusterServiceClasses:"
    kubectl get clusterserviceclasses -o custom-columns=NAME:.metadata.name,DELETION\ TIME:.metadata.deletionTimestamp
    setExitCode
fi

SC=$(kubectl get serviceclasses --all-namespaces -ojsonpath="{.items[*].metadata.deletionTimestamp}")
if [[ -n ${SC} ]]; then
    echo "There are being deleted ServiceClasses:"
    kubectl get serviceclasses --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,DELETION\ TIME:.metadata.deletionTimestamp
    setExitCode
fi

CSP=$(kubectl get clusterserviceplans -ojsonpath="{.items[*].metadata.deletionTimestamp}")
if [[ -n ${CSP} ]]; then
    echo "There are being deleted ClusterServicePlans:"
    kubectl get clusterserviceplans -o custom-columns=NAME:.metadata.name,DELETION\ TIME:.metadata.deletionTimestamp
    setExitCode
fi

SP=$(kubectl get serviceplans --all-namespaces -ojsonpath="{.items[*].metadata.deletionTimestamp}")
if [[ -n ${SP} ]]; then
    echo "There are being deleted ServicePlans:"
    kubectl get serviceplans --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,DELETION\ TIME:.metadata.deletionTimestamp
    setExitCode
fi

#
# Check if any instance/binding is in progress or is being deleted
#
SI=$(kubectl get serviceinstances --all-namespaces -ojsonpath="{.items[*].metadata.deletionTimestamp}")
if [[ -n "${SI}" ]]; then
    echo "There are being deleted ServiceInstances:"
    kubectl get serviceinstances --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,DELETION\ TIME:.metadata.deletionTimestamp
    setExitCode
fi
for status in $(kubectl get serviceinstances --all-namespaces -ojsonpath="{.items[*].status.asyncOpInProgress}")
do
if [[ -n "${status}" ]] && ${status}; then
    echo "There are ServiceInstance in progress:"
    kubectl get serviceinstances --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,IN\ PROGRESS:.status.asyncOpInProgress
    setExitCode
fi
done

SBI=$(kubectl get servicebindings --all-namespaces -ojsonpath="{.items[*].metadata.deletionTimestamp}")
if [[ -n ${SBI} ]]; then
    echo "There are being deleted ServiceBindings:"
    kubectl get servicebindings --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,DELETION\ TIME:.metadata.deletionTimestamp
    setExitCode
fi
for status in $(kubectl get servicebindings --all-namespaces -ojsonpath="{.items[*].status.asyncOpInProgress}")
do
if [[ -n "${status}" ]] && ${status}; then
    echo "There are ServiceBinding in progress:"
    kubectl get servicebindings --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,IN\ PROGRESS:.status.asyncOpInProgress
    setExitCode
fi
done

#
# Check if any broker is being deleted
#
SB=$(kubectl get servicebrokers --all-namespaces -ojsonpath="{.items[*].metadata.deletionTimestamp}")
if [[ -n ${SB} ]]; then
    echo "There are being deleted ServiceBrokers:"
    kubectl get servicebrokers --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,DELETION\ TIME:.metadata.deletionTimestamp
    setExitCode
fi
CSB=$(kubectl get clusterservicebrokers -ojsonpath="{.items[*].metadata.deletionTimestamp}")
if [[ -n ${CSB} ]]; then
    echo "There are being deleted ClusterServiceBrokers:"
    kubectl get clusterservicebrokers -o custom-columns=NAME:.metadata.name,DELETION\ TIME:.metadata.deletionTimestamp
    setExitCode
fi

if [[ ${EXIT_CODE} -eq 0 ]]; then
    echo "Your Service Catalog resources are ready to migrate!"
else
    echo "Please prepare your Service Catalog resources before migration. Check docs/migration-apiserver-to-crds.md#preparation"
fi

exit ${EXIT_CODE}
