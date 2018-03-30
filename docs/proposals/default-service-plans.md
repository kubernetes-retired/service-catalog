# Default Service Plans in Service Catalog

* [Abstract](#abstract)
* [Summary](#summary)
* [Motivation](#motivation)
* [Goals and Non-Goals](#goals-and-non-goals)
* [Storage Classes vs. Service Classes](#storage-classes-vs-service-classes)
* [User Experience](#user-experience)
  * [Cluster Operator](#cluster-operator)
     * [Broker Installation](#broker-installation)
     * [Cluster Customization](#cluster-customization)
  * [Developer](#developer)
     * [Service Class Discovery](#service-class-discovery)
     * [Legacy Provisioning](#legacy-provisioning)
     * [Provisioning with Defaults](#provisioning-with-defaults)
     * [Provisioning by Service Type](#provisioning-by-service-type)
     * [Service Instance Management](#service-instance-management)
  * [Helm Charts](#helm-charts)
     * [Broker](#broker)
     * [Author](#author)
     * [Consumer](#consumer)
* [Implementation](#implementation)
  * [Service Class](#service-class)
  * [Service Plan](#service-plan)
  * [Service Instance](#service-instance)
  * [Service Binding](#service-binding)

## Abstract

Let's wholesale rip off how default storage classes for volumes work and apply
it to Service Catalog! 😎 This allows us to specify a dependency on a type of
service without requiring upfront knowledge of the underlying broker.

> "My application requires a mysql database."

> "I just want to try out this application using the recommended defaults."

> "I need to encourage everyone to use the cheapest plans possible in our
> integration environment by making a small database the default."

## Summary

1.  Add default provisioning and binding parameters to Service Classes.
1.  Add a "Service Type" field to classes and plans, e.g. mysql or redis.
1.  Add a is-default-plan annotation to Service Plans to flag it as the default
    plan for a particular service type.
1.  Service Instances and Bindings fall back to the defaults set on the class  
    when parameters are not provided.
1.  Services can be provisioned using only the Service Type, e.g. mysql, and the
    default plan and parameters are used.

## Motivation

Consider a Helm chart author who is interested in updating their chart to be
"Service Catalog Ready"\*. Their chart requires a MySQL database and they'd like
to help their users by setting up the database in their chart, just like the
upstream Helm charts do today.

So they add a ServiceInstance to their chart, and when they get to the class and
plan fields in the spec, they hit a wall. The author doesn't know the
capabilities of the cluster on which their chart will be deployed and aren't in
a position to provide the class and plan. Neither do they know the secret keys,
as it's broker specific. The author is forced to add required arguments to their
chart, passing the buck on figuring out the right class, plan, instance
parameters and secret keys to the person deploying the chart.

Now consider a user who wants to get up and running quickly, so naturally they
chose to use a "Service Catalog Ready"* Helm chart. They don't know or care
about the details, and want to get the application deployed with sane defaults
so that they can try it out. Before they can even install the application they
need to:

* Realize that they need to specify a class and plan.
* Figure out what a broker is and which one is installed on their cluster.
* Hunt down a list of suitable classes and plans.
* Decide which class and plan is right for a trying out the application.
* Fill in the class and plan, install the chart only to later learn that there
    is a set of "magic" parameters that they should have specified when
    provisioning the instance.
* Scour the internet looking for undocumented, totally optional but it won't
    work without them, parameters for their broker because brokers aren't
    required to provide a schema.
* Blindly copy/paste some parameters they found in a gist somewhere.
* Realize that the broker is returning MYSQL_PASSWORD, while the app expects
    DB_PASSWORD, so they update the pod spec with that and redeploy the app.

**This is not the user experience that we want for Service Catalog.**

\* _I am trying to make "Service Catalog Ready" charts and "Service Catalog Enabled" clusters a thing._ 😇

This is a solved problem in Kubernetes for volumes: storage classes. They
encapsulate provisioner-specific details and allow users to either select one
using a friendly name like "slow", or omit the class and use the default set by
the operator. Service Catalog is very close to this model today, and I am
proposing that we "fill in the gaps" following the precedent set by storage.

## Goals and Non-Goals

* Enable using the same manifest across cloud providers and brokers.
* Reduce the burden of using Service Catalog by having the brokers and cluster
  operators to define defaults.
* Non-invasively add optional fields to existing resources to support this
  feature. This should not interfere with existing functionality.
* Avoid making changes to Service Catalog that aren't _strictly_ required to
  support this feature. For example, we aren't going to rename Service Bindings
  to Service Claims.
* Dynamic provisioning concepts is out of scope. Concepts such as binding to
  existing unbound instances, reclaiming instances, instance capacity, binding
  resources will be considered in a [follow-on proposal](https://docs.google.com/document/d/1U7YpZ1XHLjiXDPg0o31JC3GZDPXpQkD0PZwJY7Z_8v0/edit?usp=sharing).

## Storage Classes vs. Service Classes

| Volumes | Service Catalog | Comments |
|---------|-----------------|----------|
| Storage Class | Service Class & Service Plan | Both are referenced by name when provisioning, and provide defaults to the provisioner/broker.|
| Persistent Volume | Service Instance | |
| Persistent Volume Claim | Service Binding | Service bindings are not exclusive (1:1) like PVCs. |

## User Experience


### Cluster Operator


#### Broker Installation

If the broker supports providing default provision and bind parameters with their catalog, no additional steps are required because everything is populated during a relist.

Otherwise, the broker can provide additional manifests, or a helm chart, to include out-of-band support for the features in this proposal. See [details](#broker).

#### Cluster Customization

The cluster operator can define a new class, and provide a different set of defaults than what was included by the broker. The key details such as the broker, class and plan remain the same.

##### Define a new set of defaults for a class

```console
$ svcat create class NAME --from EXISTING_NAME \
    [--type] [--provision-params] [--bind-params] [--bind-transform]

```

1.  Copies an existing class definition.
1.  Sets a new name on the class.
1.  Sets metadata on the class to indicate that it is a user managed class.
1.  Replace any defaults set on the original class.

##### Override the defaults for a broker level class

1.  Define a new class (above).
1.  Blacklist the original class so that it cannot be used anymore. _This relies upon functionality that is still in progress._

##### Select a default plan for a service type

```console
$ svcat set-default plan NAME
OLD_NAME is no longer the default plan for TYPE
NAME is the default plan for TYPE

```

1.  If an existing plan is already the default for the type represented by the specified plan, the annotation marking the existing plan as the default is removed.
1.  An annotation is added to the specified plan, marking it as the default for its type.

### Developer


#### Service Class Discovery

Defaults are indicated with `*` after the type.


```console
$ svcat get classes
	TYPE     	NAME           	DESCRIPTION          	SCOPE
+------------+------------------+---------------------+-----------+
  redis        azure-redis        Azure Redis       	broker (osba)
  postgres     azure-postgresql   Azure PostgreSQL  	broker (osba)
  postgres     postgres-dev       Dev DB   	           cluster

$ svcat get classes --type redis
	TYPE     	NAME           	DESCRIPTION          	SCOPE
+------------+------------------+---------------------+-----------+
  redis        azure-redis    	 Azure Redis       	broker (osba)

$ svcat get plans --default
	TYPE     	NAME     CLASS DESCRIPTION          	SCOPE
+------------+--------+--------------+---------------+-----------+
  redis*     basic    azure-redis    Azure Redis     broker (osba)
  postgres*  tiny     postgres-dev   Tiny Dev DB   	cluster
```

#### Legacy Provisioning

This is unchanged. Create a service instance using a class+plan and define every parameter.

Note: The parameters defined on the instance are merged with the parameters defined on the class, but since every value is set by the instance, all class-level defaults are overridden.


#### Provisioning with Defaults

Create a service instance using a class+plan. Individual parameters may be set on the instance to override a class-level default. Any unspecified parameters are defaulted using the values defined on the class.


```console
$ svcat provision mydb --class azure-mysql --plan basic50 \
    --param location=westus
```

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: mydb
spec:
  clusterServiceClassExternalName: azure-mysql
  clusterServicePlanExternalName: basic50
  parameters:
    location: westus
```


The requested service instance is combined with the defaults defined on the class:


```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceClass
metadata:
  name: "997b8372-8dac-40ac-ae65-758b4a5075a5"
spec:
  externalID: "997b8372-8dac-40ac-ae65-758b4a5075a5"
  externalName: azure-mysql
  defaultProvisionParameters:
    location: eastus
    resourceGroup: default
    sslEnforcement: disabled
    firewallRules:
    - name: "AllowAll"
  	  startIPAddress: "0.0.0.0"
  	  endIPAddress: "255.255.255.255"
```


Yielding a final service instance. Notice that the default location parameter value "eastus" was overridden to "westus" and persisted in the status field.


```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: mydb
spec:
  clusterServiceClassExternalName: azure-mysql
  clusterServicePlanExternalName: basic50
  parameters:
    location: westus
status:
  externalProperties:
      parameters:
        location: westus
        resourceGroup: default
        sslEnforcement: disabled
        firewallRules:
        - name: "AllowAll"
    	    startIPAddress: "0.0.0.0"
    	    endIPAddress: "255.255.255.255"
```



#### Provisioning by Service Type

Create a service instance specifying only the service type, so that the default plan for that type is used. Individual parameters may be set to override a default. Any unspecified parameters are defaulted using the values defined on the class.


```console
$ svcat provision mydb --type mysql --param location=westus
```

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: mydb
spec:
  serviceType: mysql
  parameters:
    location: westus
```


The requested service type is used to filter the plans by the type and select the one annotated with is-default-plan to find the default plan:


```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServicePlan
metadata:
  name: "427559f1-bf2a-45d3-8844-32374a3e58aa"
  annotations:
    servicecatalog.kubernetes.io/is-default-plan: true
spec:
  serviceType: mysql
  clusterServiceBrokerName: osba
  clusterServiceClassRef:
    name: "997b8372-8dac-40ac-ae65-758b4a5075a5"
  description: Basic Tier, 50 DTUs.
  externalID: "427559f1-bf2a-45d3-8844-32374a3e58aa"
  externalName: basic50
  free: false
```


Then the instance is combined with the defaults defined on the class:


```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceClass
metadata:
  name: "997b8372-8dac-40ac-ae65-758b4a5075a5"
spec:
  serviceType: mysql
  externalID: "997b8372-8dac-40ac-ae65-758b4a5075a5"
  externalName: azure-mysql
  defaultProvisionParameters:
    location: eastus
    resourceGroup: default
    sslEnforcement: disabled
    firewallRules:
    - name: "AllowAll"
  	  startIPAddress: "0.0.0.0"
  	  endIPAddress: "255.255.255.255"
```


Yielding a final service instance. Notice that the default location parameter value "eastus" was overridden to "westus" and persisted in the status field.


```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: mydb
spec:
  serviceType: mysql
  clusterServiceClassRef:
    name: "997b8372-8dac-40ac-ae65-758b4a5075a5"
  clusterServicePlanRef:
    name: "427559f1-bf2a-45d3-8844-32374a3e58aa"
  parameters:
    location: westus
status:
  externalProperties:
    clusterServicePlanExternalName: basic50
    clusterServicePlanExternalID: "427559f1-bf2a-45d3-8844-32374a3e58aa"
    parameters:
      location: westus
      resourceGroup: default
      sslEnforcement: disabled
      firewallRules:
      - name: "AllowAll"
    	  startIPAddress: "0.0.0.0"
    	  endIPAddress: "255.255.255.255"
```



#### Service Instance Management


```console
$ svcat describe instance mydb
  Name:    	mydb
  Namespace:    default
  Status:  	Ready - The instance was provisioned successfully @
                2017-11-30 13:11:49 -0600 CST
  Type:    	mysql
  Class:   	azure-mysql
  Plan:    	basic50

Parameters:
  location: westus
  resourceGroup: default
  sslEnforcement: disabled
  firewallRules:
  - name: "AllowAll"
    startIPAddress: "0.0.0.0"
    endIPAddress: "255.255.255.255"
```

### Helm Charts


#### Broker

In the interim between this proposal and any upstream changes to the OSB API, brokers should include defaults for their Service Classes in their Helm charts.

Note: The defaults must be created by Helm before Service Catalog performs a list against the broker because Helm cannot modify resources that it did not create. To ensure proper ordering, the ServiceClass resources should include a "pre-install" Helm hook.


```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceClass
metadata:
  # Must match the name set during relist
  name: b9978372-8dac-40ac-ae65-758b4a5075a5
  # Ensure this the resource is created by Helm before broker list
  annotations:
    "helm.sh/hooks": pre-install # this needs to be deleted when the broker is deleted because helm won't cleanup this itself due to the hook
  # indicates that this is a broker-provided class
  ownerReferences:
  - apiVersion: servicecatalog.k8s.io/v1beta1
  	controller: true \
Spec: \
  externalId: b9978372-8dac-40ac-ae65-758b4a5075a5
  # Only need specify new values not yet included in the OSB API
  serviceType: mysql
  defaultProvisionParameters:
  	location: eastus
  	resourceGroup: default
  	sslEnforcement: disabled
  	firewallRules:
  	- name: "AllowAll"
      startIPAddress: "0.0.0.0"
      endIPAddress: "255.255.255.255"
```



#### Author

Chart authors can use service types, instead of specifying specific broker's class, plan and parameters. Optional variables can be defined so an advanced user can specify a class/plan/parameters.


```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: {{ template "fullname" . }}-mysql-instance
spec:
  serviceType: mysql-database-and-server
  {{- if .Values.className != "" }}
  clusterServiceClassName: {{ .Values.className }} # e.g. mysql
  clusterServicePlanName: {{ .Values.planName }} # e.g. ha-db
  {{- end}}
  {{- if .Values.parameters != "" }}
  {{ include "toYaml" .Values.parameters }}
  {{- end}}

```


We will need to coordinate with OSB API to define an agreed upon set of service types. Below is just a suggestion:

* **{{ service-name }}-database-only**: Represents a database only (requires an existing database server Service Instance). Example: mysql-database-only
* **{{ service-name }}-server-only**: Represents a database management system or server only. Example: mysql-server-only
* **{{ service-name }}-database-and-server**: Represents both a database server and a logical database. Example: mysql-database-and-server
* **{{ service-name }}**: Represents a service where there are not parent-child relationships between service instances. _Parent-child relationships are a new concept in OSB that is still being discussed._ Example: redis, rabbitmq


#### Consumer

End users should be able to install a chart on a cluster with any broker that supports this proposal, and not be required to specify any broker-specific values:


``` console
$ helm install azure/wordpress
# Provision a new service of type mysql and inject it into Wordpress
```


If the cluster operator has created custom classes, that can be specified when installing the chart:


``` console
$ helm install azure/wordpress \
    --set className=custom-mysql --set planName=ha-mysql
```



## Implementation

A key goal of the implementation is to avoid going off and rewriting or requiring a long term branch or fork, etc. I'd like to see this developed using a feature flag, and making all new fields optional.


### Service Class


| Field | Description |
|-------|-------------|
| Service Type | Examples: mysql, redis|
| <p>Default Provision Parameters</p><p>Default Bind Parameters</p> | <p>Default parameters to supply to the broker during provisioning and binding.</p> <p>OSB API V3 - Make the schema required, and defaults optional.|
| Default Bind Response Transform | <p>Default json key mapping to transform a non-standard broker response into something that can be relied upon.</p> <p>For example, if a broker returns "database.name" (a nested value) in the response, it can be changed to "database" to un-nest it and apply a standard name.</p>|
| Controller Reference |<p>Set by Service Catalog when populating the catalog from a broker list to indicate that it is managed by the controller.</p> <p>When not present, the class is ignored during a relist. This gives operators a way to define custom classes.</p>|

Notes

* Relist needs to take into account if the class is managed by the controller or was created by the cluster operator.
* Defaults can be defined at the Cluster or Namespace level.
* For simplification, only broker bind response bodies can be transformed, not provision, since that's where we require that functionality today.


### Service Plan

| Field | Description |
|-------|-------------|
| Service Type | Examples: mysql, redis |
| annotation: is-default-plan=true | <p>Specifies whether or not a Service Plan is the default for its Service Type.</p> <p>When annotated as the default and a Service Type is set, this plan is used to dynamically provision an instance for that type.</p> |
| Controller Reference|<p>Set by Service Catalog when populating the catalog from a broker list to indicate that it is managed by the controller.</p> <p>When not present, the plan is ignored during a relist. This gives operators a way to define custom plans.</p> |


Notes:

* The controller will enforce that only one default is set per service type.
* For simplification, defaults are set at the class level only, plans cannot override them.


### Service Instance


| Field | Description |
|-------|-------------|
| Service Type | Examples: mysql, redis |
| Provision Parameters | Overrides the default provision parameters defined on the class/plan. |

Notes:
* Merge provision defaults and overrides before sending to the broker.


### Service Binding

| Field | Description |
|-------|-------------|
| Service Type | Examples: mysql, redis |
| Service Plan | When a ServicePlan is not specified, the default ServicePlan for the type is selected. |
| Bind Parameters|Overrides the default bind parameters defined on the class/plan. |
| Bind Response Transform|Allow overriding the default bind response transform set at the class level. |

Notes:

* Merge bind defaults and overrides and before sending to the broker.
