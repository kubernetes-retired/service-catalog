#!/usr/bin/env bash

RESULT=
EXIT_CODE=0

#
# Check if any class/plan was removed from broker's catalog
#

CSC=$(kubectl get clusterserviceclasses -ojsonpath="{.items[*].metadata.deletionTimestamp}")
CSC_STATUS=$(kubectl get clusterserviceclasses -ojsonpath="{.items[*].status.removedFromBrokerCatalog}")
if [[ -n ${CSC} ]] || [[ -n ${CSC_STATUS} ]]; then
    RESULT="Some ClusterServiceClasses are dangling"
    EXIT_CODE=1
fi

SC=$(kubectl get serviceclasses --all-namespaces -ojsonpath="{.items[*].metadata.deletionTimestamp}")
SC_STATUS=$(kubectl get serviceclasses --all-namespaces -ojsonpath="{.items[*].status.removedFromBrokerCatalog}")
if [[ -n ${SC} ]] || [[ -n ${SC_STATUS} ]]; then
    RESULT="$RESULT Some ServiceClasses are dangling"
    EXIT_CODE=1
fi

CSP=$(kubectl get clusterserviceplans -ojsonpath="{.items[*].metadata.deletionTimestamp}")
CSP_STATUS=$(kubectl get clusterserviceplans -ojsonpath="{.items[*].status.removedFromBrokerCatalog}")
if [[ -n ${CSP} ]] || [[ -n ${CSP_STATUS} ]]; then
    RESULT="$RESULT Some ClusterServicePlans are dangling"
    EXIT_CODE=1
fi

SP=$(kubectl get serviceplans --all-namespaces -ojsonpath="{.items[*].metadata.deletionTimestamp}")
SP_STATUS=$(kubectl get serviceplans --all-namespaces -ojsonpath="{.items[*].status.removedFromBrokerCatalog}")
if [[ -n ${SP} ]] || [[ -n ${SP_STATUS} ]]; then
    RESULT="$RESULT Some ServicePlans are dangling"
    EXIT_CODE=1
fi

SI=$(kubectl get serviceinstances --all-namespaces -ojsonpath="{.items[*].metadata.deletionTimestamp}")
SI_STATUS=$(kubectl get serviceinstances --all-namespaces -ojsonpath="{.items[*].status.asyncOpInProgress}")
if [[ -n ${SI} ]] || [[ "${SI_STATUS}" -eq "true" ]]; then
    RESULT="$RESULT Some ServiceInstances are dangling"
    EXIT_CODE=1
fi

SBI=$(kubectl get servicebindings --all-namespaces -ojsonpath="{.items[*].metadata.deletionTimestamp}")
SBI_STATUS=$(kubectl get servicebindings --all-namespaces -ojsonpath="{.items[*].status.asyncOpInProgress}")
if [[ -n ${SBI} ]]; then
    RESULT="$RESULT Some ServiceBindings are dangling"
    EXIT_CODE=1
fi

SB=$(kubectl get servicebrokers --all-namespaces -ojsonpath="{.items[*].metadata.deletionTimestamp}")
if [[ -n ${SB} ]]; then
    RESULT="$RESULT Some ServiceBrokers are dangling"
    EXIT_CODE=1
fi
CSB=$(kubectl get clusterservicebrokers -ojsonpath="{.items[*].metadata.deletionTimestamp}")
if [[ -n ${CSB} ]]; then
    RESULT="$RESULT Some ClusterServiceBrokers are dangling"
    EXIT_CODE=1
fi


echo "${RESULT}"
echo "EXIT CODE: $EXIT_CODE"

exit ${EXIT_CODE}
