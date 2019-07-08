---
title: Walkthrough
layout: docwithnav
---

This document assumes that you've installed Service Catalog onto your cluster.
If you haven't, please see the [installation instructions](./install.md). Optionally you may install
the Service Catalog CLI, svcat. Examples for both svcat and kubectl are provided
so that you may follow this walkthrough using svcat or using only kubectl.

All commands in this document assume that you're operating out of the root
of this repository.

<a id="install" />

# Step 1 - Installing the minibroker Server

Since the Service Catalog provides a Kubernetes-native interface to an
[Open Service Broker API](https://www.openservicebrokerapi.org/) compatible broker
server, we'll need to install one in order to proceed with a demo.

We plan on using the minibroker for demo purposes. The codebase for that broker is
[here](https://github.com/kubernetes-sigs/minibroker).

We're going to deploy the minibroker to our Kubernetes cluster before
proceeding, and we'll do so with the minibroker helm chart. You can find details about the chart in the minibroker README
[README](https://github.com/kubernetes-sigs/minibroker#install-minibroker).

Otherwise, to install with sensible defaults, run the following command:

**NOTE:** The walkthrough installs a cluster-wide Broker with the defaults from minibroker.

```console
helm repo add minibroker https://minibroker.blob.core.windows.net/charts
helm install --name minibroker --namespace minibroker minibroker/minibroker
```

# Step 2 - Viewing ClusterServiceClasses and ClusterServicePlans

The controller created a `ClusterServiceClass` for each service that the minibroker
provides. We can view the `ClusterServiceClass` resources available:

```console
$ svcat get classes
     NAME      NAMESPACE          DESCRIPTION         
+------------+-----------+---------------------------+
  mariadb                  Helm Chart for mariadb     
  mongodb                  Helm Chart for mongodb     
  mysql                    Helm Chart for mysql       
  postgresql               Helm Chart for postgresql  
  redis                    Helm Chart for redis 

$ kubectl get clusterserviceclasses
NAME         EXTERNAL-NAME   BROKER       AGE
mariadb      mariadb         minibroker   5m50s
mongodb      mongodb         minibroker   5m50s
mysql        mysql           minibroker   5m50s
postgresql   postgresql      minibroker   5m50s
redis        redis           minibroker   5m50s
```

**NOTE:** The above kubectl command uses a custom set of columns.  The `NAME` field is
the Kubernetes name of the `ClusterServiceClass` and the `EXTERNAL NAME` field is the
human-readable name for the service that the broker returns.

The minibroker provides a service with the external name
`mariadb`. View the details of this offering:

```console
$ svcat describe class mariadb
  Name:              mariadb                        
  Scope:             cluster                        
  Description:       Helm Chart for mariadb         
  Kubernetes Name:   mariadb                        
  Status:            Active                         
  Tags:              mariadb, mysql, database, sql  
  Broker:            minibroker                     

Plans:
        NAME                  DESCRIPTION            
+------------------+--------------------------------+
  10-1-26            Fast, reliable, scalable,       
                     and easy to use open-source     
                     relational database system.     
                     MariaDB Server is intended      
                     for mission-critical,           
                     heavy-load production systems   
                     as well as for embedding into   
                     mass-deployed software.         
  10-1-28            Fast, reliable, scalable,       
                     and easy to use open-source     
                     relational database system.     
                     MariaDB Server is intended      
                     for mission-critical,           
                     heavy-load production systems   
                     as well as for embedding into   
                     mass-deployed software. 
.
.
.
$ kubectl get clusterserviceclasses mariadb -o yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceClass
metadata:
  creationTimestamp: "2019-06-18T01:52:07Z"
  name: mariadb
  ownerReferences:
  - apiVersion: servicecatalog.k8s.io/v1beta1
    blockOwnerDeletion: false
    controller: true
    kind: ClusterServiceBroker
    name: minibroker
    uid: 6a9a047e-916b-11e9-bfe5-0242ac110008
  resourceVersion: "9"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterserviceclasses/mariadb
  uid: adf6e194-916b-11e9-bfe5-0242ac110008
spec:
  bindable: true
  bindingRetrievable: false
  clusterServiceBrokerName: minibroker
  description: Helm Chart for mariadb
  externalID: mariadb
  externalName: mariadb
  planUpdatable: false
  tags:
  - mariadb
  - mysql
  - database
  - sql
status:
  removedFromBrokerCatalog: false
```

Additionally, the controller created a `ClusterServicePlan` for each of the
plans for the broker's services. We can view the `ClusterServicePlan`
resources available in the cluster:

```console
$ svcat get plans
       NAME         NAMESPACE     CLASS                DESCRIPTION            
+------------------+-----------+------------+---------------------------------+
  10-1-26                        mariadb      Fast, reliable, scalable,        
                                              and easy to use open-source      
                                              relational database system.      
                                              MariaDB Server is intended       
                                              for mission-critical,            
                                              heavy-load production systems    
                                              as well as for embedding into    
                                              mass-deployed software.          
  10-1-28                        mariadb      Fast, reliable, scalable,        
                                              and easy to use open-source      
                                              relational database system.      
                                              MariaDB Server is intended       
                                              for mission-critical,            
                                              heavy-load production systems    
                                              as well as for embedding into    
                                              mass-deployed software.          
.
.
.
$ kubectl get clusterserviceplans
NAME                       EXTERNAL-NAME      BROKER       CLASS        AGE                                                                                                                                                                                                    
mariadb-10-1-26            10-1-26            minibroker   mariadb      34m                                                                                                                                                                                                    
mariadb-10-1-28            10-1-28            minibroker   mariadb      34m                                                                                                                                                                                                    
mariadb-10-1-29            10-1-29            minibroker   mariadb      34m                                                                                                                                                                                                    
mariadb-10-1-30            10-1-30            minibroker   mariadb      34m                                                                                                                                                                                                    
mariadb-10-1-31            10-1-31            minibroker   mariadb      34m                                                                                                                                                                                                    
mariadb-10-1-32            10-1-32            minibroker   mariadb      34m                                                                                                                                                                                                    
mariadb-10-1-33            10-1-33            minibroker   mariadb      34m                                                                                                                                                                                                    
mariadb-10-1-34            10-1-34            minibroker   mariadb      34m                                                                                                                                                                                                    
mariadb-10-1-34-debian-9   10-1-34-debian-9   minibroker   mariadb      34m
```

You can view the details of a `ClusterServicePlan` with this command:

```console
$ svcat describe plan 10-1-26 --scope cluster
  Name:              10-1-26                                                                                                                                                                                                                 
  Description:       Fast, reliable, scalable, and easy to use open-source relational database system. MariaDB Server is intended for mission-critical, heavy-load production systems as well as for embedding into mass-deployed software.  
  Kubernetes Name:   mariadb-10-1-26                                                                                                                                                                                                         
  Status:            Active                                                                                                                                                                                                                  
  Free:              true                                                                                                                                                                                                                    
  Class:             mariadb                                                                                                                                                                                                                 

Instances:
No instances defined

$ kubectl get clusterserviceplans mariadb-10-1-26 -o yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServicePlan
metadata:
  creationTimestamp: "2019-06-18T03:32:31Z"
  name: mariadb-10-1-26
  ownerReferences:
  - apiVersion: servicecatalog.k8s.io/v1beta1
    blockOwnerDeletion: false
    controller: true
    kind: ClusterServiceBroker
    name: minibroker
    uid: 6ee99fc5-9179-11e9-8cb6-0242ac110009
  resourceVersion: "28"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterserviceplans/mariadb-10-1-26
  uid: b496eaf5-9179-11e9-8cb6-0242ac110009
spec:
  clusterServiceBrokerName: minibroker
  clusterServiceClassRef:
    name: mariadb
  description: Fast, reliable, scalable, and easy to use open-source relational database
    system. MariaDB Server is intended for mission-critical, heavy-load production
    systems as well as for embedding into mass-deployed software.
  externalID: mariadb-10-1-26
  externalName: 10-1-26
  free: true
status:
  removedFromBrokerCatalog: false
```

# Step 4 - Creating a New ServiceInstance

Now that a `ClusterServiceClass` named `mariadb` exists within our
cluster's service catalog, we can create a `ServiceInstance` that points to
it.

Unlike `ClusterServiceBroker` and `ClusterServiceClass` resources, `ServiceInstance`
resources must be namespaced. Create a namespace with the following command:

```console
$ kubectl create namespace test-ns
namespace "test-ns" created
```

Then, create the `ServiceInstance`:

```console
$ kubectl create -f contrib/examples/walkthrough/mini-instance.yaml
serviceinstance.servicecatalog.k8s.io/mini-instance created
```

After the `ServiceInstance` is created, the service catalog controller will
communicate with the appropriate broker server to initiate provisioning.
Check the status of that process:

```console
$ svcat describe instance -n test-ns mini-instance
  Name:        mini-instance                                                                      
  Namespace:   test-ns                                                                            
  Status:      Ready - The instance was provisioned successfully @ 2019-06-18 02:42:55 +0000 UTC  
  Class:       mariadb                                                                            
  Plan:        10-1-26                                                                            

Parameters:
  param-1: value-1
  param-2: value-2

Bindings:
No bindings defined

$ kubectl get serviceinstances -n test-ns mini-instance -o yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  creationTimestamp: "2019-06-18T02:42:50Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  name: mini-instance
  namespace: test-ns
  resourceVersion: "93"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/test-ns/serviceinstances/mini-instance
  uid: c3b56b7e-9172-11e9-bfe5-0242ac110008
spec:
  clusterServiceClassExternalName: mariadb
  clusterServiceClassRef:
    name: mariadb
  clusterServicePlanExternalName: 10-1-26
  clusterServicePlanRef:
    name: mariadb-10-1-26
  externalID: c3b56b2e-9172-11e9-bfe5-0242ac110008
  parameters:
    param-1: value-1
    param-2: value-2
  updateRequests: 0
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: minikube-user
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: "2019-06-18T02:42:55Z"
    message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
  deprovisionStatus: Required
  externalProperties:
    clusterServicePlanExternalID: mariadb-10-1-26
    clusterServicePlanExternalName: 10-1-26
    parameterChecksum: 4fa544b50ca7a33fe5e8bc0780f1f36aa0c2c7098242db27bc8a3e21f4b4ab55
    parameters:
      param-1: value-1
      param-2: value-2
    userInfo:
      groups:
      - system:masters
      - system:authenticated
      uid: ""
      username: minikube-user
  observedGeneration: 1
  orphanMitigationInProgress: false
  provisionStatus: Provisioned
  reconciledGeneration: 1
```

# Step 5 - Requesting a ServiceBinding to use the ServiceInstance

Now that our `ServiceInstance` has been created, we can bind to it.
Create a `ServiceBinding` resource:

```console
$ kubectl create -f contrib/examples/walkthrough/mini-binding.yaml
servicebinding.servicecatalog.k8s.io/mini-binding created
```

After the `ServiceBinding` resource is created, the service catalog controller will
communicate with the appropriate broker server to initiate binding. Generally,
this will cause the broker server to create and issue credentials that the
service catalog controller will insert into a Kubernetes `Secret`. We can check
the status of this process like so:

```console
$ svcat describe binding -n test-ns mini-binding
  Name:        mini-binding                                                  
  Namespace:   test-ns                                                       
  Status:      Ready - Injected bind result @ 2019-06-18 02:45:41 +0000 UTC  
  Secret:      mini-binding                                                  
  Instance:    mini-instance                                                 

Parameters:
  No parameters defined

Secret Data:
  Protocol                5 bytes   
  host                    47 bytes  
  mariadb-password        10 bytes  
  mariadb-root-password   10 bytes  
  password                10 bytes  
  port                    4 bytes   
  uri                     76 bytes  
  username                4 bytes

$ kubectl get servicebindings -n test-ns mini-binding -o yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  creationTimestamp: "2019-06-18T02:45:40Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  name: mini-binding
  namespace: test-ns
  resourceVersion: "97"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/test-ns/servicebindings/mini-binding
  uid: 28d115b0-9173-11e9-bfe5-0242ac110008
spec:
  externalID: 28d11555-9173-11e9-bfe5-0242ac110008
  instanceRef:
    name: mini-instance
  secretName: mini-binding
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: minikube-user
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: "2019-06-18T02:45:41Z"
    message: Injected bind result
    reason: InjectedBindResult
    status: "True"
    type: Ready
  externalProperties:
    userInfo:
      groups:
      - system:masters
      - system:authenticated
      uid: ""
      username: minikube-user
  orphanMitigationInProgress: false
  reconciledGeneration: 1
  unbindStatus: Required
```

Notice that the status has a `Ready` condition set.  This means our binding is
ready to use!  If we look at the `Secret`s in our `test-ns` namespace, we should
see a new one:

```console
$ kubectl get secrets -n test-ns
NAME                    TYPE                                  DATA   AGE
default-token-n2j75     kubernetes.io/service-account-token   3      10m
mini-binding            Opaque                                8      91s
```

Notice that a new `Secret` named `mini-binding` has been created.
At this point,we could use this secret to connect the running MariaDB instance to our application running on Kubernetes.

# Step 6 - Deleting the ServiceBinding

Now, let's unbind the instance:

```console
$ svcat unbind -n test-ns mini-instance
deleted mini-binding
```

After the deletion is complete, we should see that the `Secret` is gone:

```console
$ kubectl get secrets -n test-ns
NAME                    TYPE                                  DATA   AGE
default-token-n2j75     kubernetes.io/service-account-token   3      11m
```

# Step 7 - Deleting the ServiceInstance

Now, we can deprovision the instance:

```console
$ svcat deprovision -n test-ns mini-instance
deleted mini-instance
```

# Step 8 - Deleting the ClusterServiceBroker

Next, we should remove the `ClusterServiceBroker` resource. This tells the service
catalog to remove the broker's services from the catalog. Do so with this
command:

```console
$ kubectl delete clusterservicebrokers minibroker
clusterservicebroker.servicecatalog.k8s.io "minibroker" deleted
```

We should then see that all the `ClusterServiceClass` resources that came from that
broker have also been deleted:

```console
$ svcat get classes
  NAME   NAMESPACE   DESCRIPTION
+------+-----------+-------------+

$ kubectl get clusterserviceclasses
No resources found.
```

# Step 9 - Final Cleanup

## Cleaning up the User Provided Service Broker

To clean up, delete the helm deployment:

```console
helm delete --purge minibroker
```

Then, delete all the namespaces we created:

```console
kubectl delete ns test-ns minibroker
```

## Cleaning up the Service Catalog

Delete the helm deployment and the namespace:

```console
helm delete --purge catalog
kubectl delete ns svc-cat
```

# Troubleshooting

## Firewall rules

If you are using Google Cloud Platform, you may need to run the following
command to setup proper firewall rules to allow your traffic get in.

```console
gcloud compute firewall-rules create allow-service-catalog-secure --allow tcp:30443 --description "Allow incoming traffic on 30443 port."
```
