# Kubernetes resources for interacting directly with kubectl / k8s API server.

This handles creation of a Third Party Resource as catalog.k8s.io/v1alpha1 and
creates these resources under it:

* ServiceBroker
* ServiceClass
* ManagedService
* ServiceInstance
* ServiceBinding

## Prerequisites

Kubernetes had a bug prior to 1.4 which prevented it from handling more than one
Third Party Resource under a given group. Hence it is impossible to have our
resources there, it would only see one (latest?). So you must run against
a cluster that's >=1.4 as well as have the kubectl client be of version >=
1.4. To check the version of the cluster you're running at, you can use:

    kubectl version

Both Server and Client Version should be >= 1.4.

## Getting the credentials for the cluster

You need to fetch the credentials that you want to use for the k8s cluster so
that the watch can actually talk to the k8s api. It's easiest to just spin a new
GKE cluster (so you get >=1.4) and grab the credentials for it. From the ../
directory, you could do this:

    gcloud container clusters create <clustername>
    env KUBECONFIG=./kubeconfig gcloud container clusters get-credentials <clustername>

## To create a service instance

    kubectl create -f watch/typedata/service_instance_app.yaml

## Known issues

This is SUPER early work, so the list is long

* watch times out
* error handling is minimal at best

