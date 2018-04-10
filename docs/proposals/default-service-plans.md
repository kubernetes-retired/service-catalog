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
         * [Modify existing defaults](#modify-existing-defaults)
         * [Define a new set of defaults](#define-a-new-set-of-defaults)
         * [Override the broker defaults](#override-the-broker-defaults)
         * [Select a default plan for a service type](#select-a-default-plan-for-a-service-type)
   * [Developer](#developer)
      * [Service Class Discovery](#service-class-discovery)
      * [Legacy Provisioning](#legacy-provisioning)
      * [Provisioning with Defaults](#provisioning-with-defaults)
      * [Binding with Defaults](#binding-with-defaults)
      * [Provisioning by Service Type](#provisioning-by-service-type)
      * [Service Instance Management](#service-instance-management)
      * [Service Binding Management](#service-binding-management)
   * [Helm Charts](#helm-charts)
      * [Author](#author)
      * [Consumer](#consumer)
* [Implementation](#implementation)
   * [Service Class](#service-class)
   * [Service Plan](#service-plan)
   * [Default Plan Resolution](#default-plan-resolution)
   * [Service Instance](#service-instance)
   * [Service Binding](#service-binding)
   * [OSB Compatibility](#osb-compatibility)
      * [Plan Tags](#plan-tags)
      * [Service Types](#service-types)
      * [Suggested Plan](#suggested-plan)
      * [Parameter Defaults](#parameter-defaults)
      * [Schemas](#schemas)
      * [Plan Attributes](#plan-attributes)
* [Milestones](#milestones)
   * [Default Parameters](#default-parameters)
   * [Default Secret Transform](#default-secret-transform)
   * [User-Defined Plans](#user-defined-plans)
   * [Default Plans by Service Type](#default-plans-by-service-type)
   * [Suggested Parameters](#suggested-parameters)
   * [Suggested Plans](#suggested-plans)

## Abstract

Let's wholesale rip off how default storage classes for volumes work and apply
it to Service Catalog! 😎 This allows us to specify a dependency on a type of
service without requiring upfront knowledge of the underlying broker.

> "My application requires a mysql database."

> "I just want to try out this application using the recommended defaults."

> "I need to encourage everyone to use the cheapest plans possible in our
> integration environment by making a small database the default."

## Summary

1.  Add default provisioning and binding parameters to Service Classes and Plans.
1.  Add a "Service Type" field to classes and plans, e.g. mysql or redis.
1.  Add an is-default-plan annotation to Service Plans to flag it as the default
    plan for a particular service type.
1.  Service Instances and Bindings fall back to the defaults set on the plan or class  
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

Otherwise, the broker can provide additional manifests, or a helm chart, to include out-of-band support for the features in this proposal. See [details](#osb-compatibility).

#### Cluster Customization

The cluster operator can define a new class or plan, and provide a different set of defaults than what was included by the broker. The key details such as the broker, class and plan remain the same.

##### Modify existing defaults

```console
$ svcat set class NAME [--provision-params] [--bind-params] [--secret-transform]

$ svcat set plan NAME [--provision-params] [--bind-params] [--secret-transform] [--default]
```

This updates the existing resources with the specified changes.

NOTE: Once the broker starts providing defaults, if a cluster operator patches a service class/plan
to set defaults, their changes will be overwritten at the next relist.
Ideally svcat would warn about that when we finally get to that state.

##### Define a new set of defaults

```console
$ svcat create class NAME --from EXISTING_NAME [--type] [--provision-params] [--bind-params] [--secret-transform]

$ svcat create plan NAME --from EXISTING_NAME [--type] [--provision-params] [--bind-params] [--secret-transform] [--default]
```

1.  Copies an existing resource definition.
1.  Sets a new name on the resource.
1.  Sets metadata on the resource to indicate that it is managed by the user (and not the broker).
1.  Replace any defaults set on the original resource.

##### Override the broker defaults

1.  Define a new class or plan (above).
1.  Blacklist the original resource so that it cannot be used anymore. _This relies upon functionality that is still in progress._

\* The operator should not directly modify the defaults defined on a broker provided class or plan, because the changes are overwritten during relist.

##### Select a default plan for a service type

```console
$ svcat set plan NAME --default
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
  postgres     postgres-dev       Dev DB   	           ns (dev)

$ svcat describe class postgres-dev
  Name: postgres-dev
  Scope: namespace
  Namespace: dev
  ...

$ svcat get classes --namespace qa
	TYPE     	NAME           	DESCRIPTION          	SCOPE
+------------+------------------+---------------------+-----------+
  postgres     postgres-qa        QA DB   	           ns (qa)

$ svcat get classes --cluster
	TYPE     	NAME           	DESCRIPTION          	SCOPE
+------------+------------------+---------------------+-----------+
  redis        azure-redis        Azure Redis       	broker (azure)
  postgres     azure-postgresql   Azure PostgreSQL  	broker (azure)
  postgres     postgres-dev       Dev DB   	           cluster

$ svcat describe class postgres-dev --cluster
  Name: postgres-dev
  Scope: cluster
  ...

$ svcat describe class azure-postgresql --cluster
  Name: azure-postgresql
  Scope: broker
  Broker: azure
  ...

$ svcat get classes --cluster --all-namespaces
	TYPE     	NAME           	DESCRIPTION          	SCOPE
+------------+------------------+---------------------+-----------+
  redis        azure-redis        Azure Redis       	broker (azure)
  postgres     azure-postgresql   Azure PostgreSQL  	broker (azure)
  postgres     postgres-dev       Dev DB   	           cluster
  postgres     postgres-dev       Dev DB   	           ns (dev)
  postgres     postgres-qa        QA DB   	           ns (qa)

$ svcat get classes --type redis
	TYPE     	NAME           	DESCRIPTION          	SCOPE
+------------+------------------+---------------------+-----------+
  redis        azure-redis    	 Azure Redis       	broker (azure)

$ svcat get plans --default
	TYPE  NAME     CLASS DESCRIPTION          	SCOPE
+------------+--------+--------------+---------------+-----------+
  redis*     basic    azure-redis    Azure Redis     broker (azure)
  postgres*  tiny     postgres-dev   Tiny Dev DB   	cluster
```

#### Legacy Provisioning

This is unchanged. Create a service instance using a class+plan and define every parameter.

Note: The parameters defined on the instance are merged with the parameters defined on the class, but since every value is set by the instance, all class-level defaults are overridden.


#### Provisioning with Defaults

Create a service instance using a class+plan. Individual parameters may be set on the instance to override a class+plan default. Any unspecified parameters are defaulted using the values defined on the class+plan.


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


The requested service instance is combined with the defaults defined on the class+plan:


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
---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServicePlan
metadata:
  name: "427559f1-bf2a-45d3-8844-32374a3e58aa"
  annotations:
    servicecatalog.kubernetes.io/is-default-plan: true
spec:
  serviceType: mysql
  clusterServiceBrokerName: azure
  clusterServiceClassRef:
    name: "997b8372-8dac-40ac-ae65-758b4a5075a5"
  description: Basic Tier, 50 DTUs.
  externalID: "427559f1-bf2a-45d3-8844-32374a3e58aa"
  externalName: basic50
  free: false
  # No defaults defined, so the defaults on the class are used
```


Yielding a final service instance. Notice that the default location parameter value "eastus" was overridden to "westus" and the entire final set of parameters were persisted on the instance. From this point forward, the instance manages its parameters independently of the defaults on the class+plan, just as Service Catalog works today.


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
      resourceGroup: default
      sslEnforcement: disabled
      firewallRules:
      - name: "AllowAll"
        startIPAddress: "0.0.0.0"
        endIPAddress: "255.255.255.255"
```


#### Binding with Defaults

Create a service binding. Individual parameters may be set on the binding to override a class+plan default. Any unspecified parameters are defaulted using the values defined on the class+plan.


```console
$ svcat bind mydb --name mydb-admin --param user=admin
```

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  name: mydb-admin
spec:
  instanceRef:
    name: mydb
  parameters:
    user: admin
```


The requested service binding is combined with the defaults defined on the class+plan:


```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceClass
metadata:
  name: "997b8372-8dac-40ac-ae65-758b4a5075a5"
spec:
  externalName: azure-mysql
  # spec truncated for clarity
  # No binding defaults defined, so the defaults on the plan are used
---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServicePlan
metadata:
  name: "427559f1-bf2a-45d3-8844-32374a3e58aa"
spec:
  clusterServiceClassRef:
    name: "997b8372-8dac-40ac-ae65-758b4a5075a5"
  externalName: basic50
  # spec truncated for clarity
  defaultBindingParameters:
    user: reader
  defaultSecretTransform:
  - renameKey:
      from: db-name
      to: database
```


Yielding a final service binding. From this point forward, the binding manages its spec independently of the defaults on the class+plan, just as Service Catalog works today.


```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  name: mydb-admin
spec:
  instanceRef:
    name: mydb
  parameters:
    user: admin
  secretTransform:
  - renameKey:
      from: db-name
      to: database
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
  name: "f9a3cc8e-a6e2-474d-b032-9837ea3dfcaa"
  annotations:
    servicecatalog.kubernetes.io/is-default-plan: true
spec:
  serviceType: mysql
  clusterServiceBrokerName: azure
  clusterServiceClassRef:
    name: "997b8372-8dac-40ac-ae65-758b4a5075a5"
  description: "PremiumP1 Tier, 125 DTUs, 500GB, 35 days point-in-time restore"
  externalID: "f9a3cc8e-a6e2-474d-b032-9837ea3dfcaa"
  externalName: premium-p1
  free: false
  defaultProvisionParameters:
    backup-schedule: "1d"
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


Yielding a final service instance. Notice that the default location parameter value "eastus" from the class was overridden by the instance to "westus" and also the "backup-schedule" parameter from the plan was merged with the class defaults as well. The entire final set of parameters were persisted on the instance. From this point forward, the instance manages its parameters independently of the defaults on the class+plan, just as Service Catalog works today.


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
    name: "f9a3cc8e-a6e2-474d-b032-9837ea3dfcaa"
  parameters:
    backup-schedule: "1d"
    location: westus
    resourceGroup: default
    sslEnforcement: disabled
    firewallRules:
    - name: "AllowAll"
      startIPAddress: "0.0.0.0"
      endIPAddress: "255.255.255.255"
status:
  externalProperties:
    clusterServicePlanExternalName: premium-p1
    clusterServicePlanExternalID: "f9a3cc8e-a6e2-474d-b032-9837ea3dfcaa"

```

#### Service Instance Management

The service type and final set of parameters should be displayed when viewing a service instance.

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

#### Service Binding Management

The service type and final set of parameters should be displayed when viewing a service binding.

```console
$ svcat describe binding mydb-admin
  Name:    	mydb-admin
  Namespace:    default
  Status:  	Ready - The binding was injected successfully @
                2017-11-30 13:11:49 -0600 CST
  Type:    	mysql
  Instance: mydb

Parameters:
  user: admin

Secret Transform:
  db-name -> database
```

### Helm Charts

#### Author

Chart authors can use service types, instead of specifying specific broker's class, plan and parameters. Optional variables can be defined so an advanced user can specify a class, plan and parameters just like they can today with Service Catalog. [See OSB Service Types](#service-types) for more information.


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
| Default Secret Transform | Default json key mapping to transform a non-standard broker response into something that can be relied upon. See #1868 for implementation details. |
| Controller Reference |<p>Set by Service Catalog when populating the catalog from a broker list to indicate that it is managed by the controller.</p> <p>When not present, the class is ignored during a relist. This gives operators a way to define custom classes.</p>|

Notes

* Relist needs to take into account if the class is managed by the controller or was created by the cluster operator.
* Defaults can be defined at the Cluster or Namespace level.


### Service Plan

| Field | Description |
|-------|-------------|
| Service Type | Examples: mysql, redis |
| annotation: is-default-plan=true | <p>Specifies whether or not a Service Plan is the default for its Service Type.</p> <p>When annotated as the default and a Service Type is set, this plan is used to dynamically provision an instance for that type.</p> |
| annotation: is-suggested-plan=true | <p>Specifies whether or not a Service Plan is the broker suggested default for its Service Type.</p> <p>When annotated as the default and a Service Type is set, *and no plans with `is-default-plan` are defined, this plan is used to dynamically provision an instance for that type.</p> |
| <p>Default Provision Parameters</p><p>Default Bind Parameters</p> | <p>Default parameters to supply to the broker during provisioning and binding.</p><p>Overrides the defaults defined on the class.</p>|
| Default Secret Transform | Default json key mapping to transform a non-standard broker response into something that can be relied upon. See #1868 for implementation details. |
| Controller Reference|<p>Set by Service Catalog when populating the catalog from a broker list to indicate that it is managed by the controller.</p> <p>When not present, the plan is ignored during a relist. This gives operators a way to define custom plans.</p> |


Notes:

* The controller will enforce that only one default is set per service type.
* When a plan does not define any defaults, the defaults from the class are used. Otherwise the values are merged, with the plan level defaults "winning".

### Default Plan Resolution

1. During a list, brokers provide classes+plans. A broker can suggest a plan as
   a good default by including a pre-defined tag, such as `SuggestedPlan=true`.
1. Service catalog looks for that tag during List and populates an annotation,
   `is-suggested-plan`.
1. Cluster operators can flag a plan as the default using svcat. This will add a
   different annotation `is-default-plan`, and Service catalog will handle
   unsetting the `is-default-plan` annotation on any other plan of the same service
   type.

When service catalog searches for the default plan it will:

1. List all plans that have that are flagged as suggested or default for that service type.
1. When more than one plan has `is-default-plan` set, the operation should fail.
1. When there is one plan with `is-default-plan` set, that plan is used as the default.
1. Otherwise, when more than one plan has `is-suggested-plan` set, the operation
    should fail. This happens when multiple brokers exist on the cluster, provide
    a default plan for a particular service type, and the operator has not picked
    a default.
1. When there is on plan with `is-suggested-plan` set, that plan is used as the default.

In the simplest case, there is one broker that provides default plans, the operator does nothing and people can still get a default plan resolved without any extra work. Otherwise, the operator must step in and pick a default (or the user needs to be specific instead of relying on a default).

### Service Instance


| Field | Description |
|-------|-------------|
| Service Type | Examples: mysql, redis |
| Parameters | Overrides the default provision parameters defined on the class+plan. |

Notes:
* Merge provision defaults and overrides before sending to the broker.


### Service Binding

| Field | Description |
|-------|-------------|
| Service Type | Examples: mysql, redis |
| Service Plan | When a ServicePlan is not specified, the default ServicePlan for the type is selected. |
| Parameters| Overrides the default bind parameters defined on the class+plan. |
| Secret Transform| Overrides the default secret transform set on the class+plan. |

Notes:

* Merge bind defaults and overrides and before sending to the broker.

### OSB Compatibility

These are the various changes that I am proposing to OSB API v3, and what we can
do today to be compatible with v2.

For OSB v2 brokers, or v3 brokers that chose to not implement optional features,
brokers (or a motivated operator) can include defaults for their Service Classes
in a Helm chart.

Note: The defaults must be created by Helm before Service Catalog performs a list
against the broker because Helm cannot modify resources that it did not create.
To ensure proper ordering, the ServiceClass resources should include a "pre-install"
Helm hook.

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceClass
metadata:
  # Must match the name set during relist
  name: "997b8372-8dac-40ac-ae65-758b4a5075a5"
  # Ensure this the resource is created by Helm before broker list
  annotations:
    "helm.sh/hooks": pre-install
  # Indicates that this is a broker-provided class
  ownerReferences:
  - apiVersion: servicecatalog.k8s.io/v1beta1
  	controller: true
Spec:
  externalId: "997b8372-8dac-40ac-ae65-758b4a5075a5"
  # Only need specify new fields not set during List
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

#### Plan Tags
The following changes all mention tags, but currently tags are only at the class level.
It would be much easier if plans also had tags.

#### Service Types
Optionally have brokers include the service type for a class during List,
e.g. mysql database, redis cache.

This can be accomplished without a schema change by having the broker specify the
service type using a pre-defined format, for example `ServiceType=mysql`.
Service Catalog would scan the tags for the service type and parse it.

Brokers must agree upon a standard set of names to use for those service types
so that people can rely upon using the same one across brokers.

Below is one possible format:

* **{{ service-name }}-database-only**: Represents a database only
  (requires an existing database server Service Instance).
  Example: mysql-database-only
* **{{ service-name }}-server-only**: Represents a database management system or server only.
  Example: mysql-server-only
* **{{ service-name }}-database-and-server**: Represents both a database server and a logical database.
  Example: mysql-database-and-server
* **{{ service-name }}**: Represents a service where there are not parent-child relationships between service instances.
  Example: redis, rabbitmq

_Parent-child relationships are a new concept in OSB that is still being discussed
but is already implemented by some brokers, such as Azure._

#### Suggested Plan
Optionally have brokers suggest a default plan for a particular service type.
During a list, the broker can include a tag using a pre-defined format, for example
`SuggestedPlan=true`. Service Catalog would scan the tags for suggested plans,
and flag them when found.

#### Parameter Defaults
Optionally have brokers include default parameters for provisioning and binding
during List.

*Open Question:* How does a broker distinguish between a default that it will apply
when no value is provided, and a suggested default that the platform can provide when
no value is provided. i.e. broker defaults vs, the broker's suggestion of a complete working
request that will successfully provision.

#### Schemas
Require the broker to return a schema for the provision and bind parameters.

Additionally, brokers should agree upon a standard json schema for common binding responses so
that people can rely upon using the same schema across brokers (like CF does).

While isn't strictly necessary for this proposal, but would be so incredibly helpful
that I am pushing for it regardless. Service Catalog has ways to work around this
now with #1868 but not having to do it in the first place is preferable.

#### Plan Attributes
Optionally have brokers include tags for plans, which can be used to specify
attributes about the plan such as database cores (or whatever the term is to indicate how huuuuuuuge a db is).

Examples:
`Cores=4`
`Memory=16`

Again, while not strictly necessary, it would be useful. Right now an operator
usually ends up going to the broker's website to look up the attributes of a plan
and this would save us all a lot of time.

## Milestones
This breaks down the proposed changes into smaller chunks of useful work:

### Default Parameters
An operator can define default parameters for classes and plans either
by installing Helm charts before installing the broker, or by editing existing
resources.

* Add default provisioning parameters field to service classes and plans.
* Add default binding parameters field to classes and plans.
* Merge the instance's parameters with the class+plan defaults
  during service instance reconciliation.
* Merge the binding's parameters with the class+plan defaults
  during service binding reconciliation.
* Output default parameters in `svcat describe` for classes and plans.
* Make sure that the controller doesn't completely overwrite class and plan resources
  during a relist so that defaults applied out-of-band are not lost. This may not
  require any changes, but we should check.
* Support `svcat set class|plan --provision-params --bind-params`.

Details:
* [Provisioning with Default Parameters](#provisioning-with-defaults)
* [Binding with Default Parameters](#binding-with-defaults)
* [Modify existing defaults](#modify-existing-defaults)
* [Service Instance Management](#service-instance-management)
* [Service Binding Management](#service-binding-management)

### Default Secret Transform
An operator can define default secret transformations for classes and plans either
by installing Helm charts before installing the broker, or by patching existing
resources.

* Add default secret transformation field to service classes and plans.
* Merge the bindings's secret transformation with the class+plan defaults during
  service binding reconciliation.
* Output default secret transforms in `svcat describe` for classes and plans.
* Support `svcat set class|plan --secret-transform`.

Details:
* [Binding with Default Secret Transforms](#binding-with-defaults)
* [Modify existing defaults](#modify-existing-defaults)

### User-Defined Plans
An operator can create custom service classes and plans based on an existing resource.

* Set a controller reference on broker-managed service classes and plans during relist.
* Only reconcile service classes or plans that have a controller reference to avoid
  overwriting or deleting user-defined resources.
* Output a scope column in `svcat get` for classes and plans that indicates the
  origin+scope of the resource. For example: `broker (azure)`, `cluster`, `ns (dev)`.
* Output a scope and either a broker or namespace field in `svcat describe` for classes
  and plans.
* Support `svcat create` for classes and plans.

Details:
* [Create a New Class or Plan](#define-a-new-set-of-defaults)
* [Override a Class or Plan](#override-the-broker-defaults)
* [Service Class Discovery](#service-class-discovery)

### Default Plans by Service Type
An operator can mark a plan as the default for a particular service type, allowing
users to provision specifying only the service type.

* Add the ServiceType field to service classes, plans, instances and bindings.
* Output the service type in `svcat get` and `svcat describe`.
* Indicate default service plans in `svcat get` with an `*` after the service type, e.g. `redis*`.
* Add a default: true/false field to `svcat describe` for service plans.
* Resolve the default when during service instance reconciliation when no plan is specified.
* Support `svcat set plan --default`.

Details:
* [Modify existing defaults](#modify-existing-defaults)
* [Provisioning by Service Type](#provisioning-by-service-type)
* [Default Plan Resolution](#default-plan-resolution)

### Suggested Parameters
Brokers can provide a set of suggested parameters that can be used by the platform to provision or
bind when no parameters are specified. The service class and plan default provision
and bind parameters are set during relist automatically when present.

Details are TBD based on the open questions in [Default Parameters](#default-parameters).

### Suggested Plans
Brokers can indicate that a plan is the suggested default plan for a service type.

* Parse tags and mark the plan as suggested during relist.
* Parse tags and populate service type on classes and plans during relist.
* Default plan resolution falls back to suggested plans when no default plans are set.
* Indicate suggested service plans in `svcat get` with an `*` after the service type
  when no default for that type is set.

Details:
* [Default Plan Resolution](#default-plan-resolution)
* [Service Class Discovery](#service-class-discovery)
