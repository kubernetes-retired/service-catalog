---
title: Service Plan Defaults
layout: docwithnav
---

Service Plan Defaults is a new feature in Service Catalog that gives operators
the ability to configure default provision parameters for a plan. When a new
instance of that plan is created, the default provision parameters defined
on the plan are merged with parameters defined on the instance before it is
provisioned

For example, the operator could define a default set of IP addresses allowed to
connect to databases, or require TLS by default.

## Enable Service Plan Defaults

Service Plan Defaults is an alpha-feature of Service 
Catalog that is off by default. To enable this feature, you will need 
to pass an argument to the API Server when you install Service Catalog:
 `--feature-gates ServicePlanDefaults=true`.

If you are using Helm, you can use the `servicePlanDefaultsEnabled` setting
 to control that flag:

```
helm install svc-cat/catalog --name catalog --set servicePlanDefaultsEnabled=true
```

## Define Default Provision Parameters

You can either define default provision parameters on an existing plan,
or create a custom plan with the parameters defined.

### Create and modify a copy of an existing plan

1. Using an existing plan as a template, save its definition to a yaml file.
    
    ```
    kubectl get clusterserviceplan -o yaml PLAN > custom-plan.yaml
    ```
1. Edit the yaml file and remove the `ownerReferences` node from the metadata.
    This indicates to Service Catalog that it is a user-managed plan.
1. Change the `name` and `externalName` of the plan to a unique value.
1. Add a `defaultProvisionParameters` node to the spec and define the default
    parameters:
                                
    ```yaml
    apiVersion: servicecatalog.k8s.io/v1beta1
    kind: ClusterServicePlan
    metadata:
      name: custom-mysql
    spec:
      clusterServiceBrokerName: minibroker
      externalID: mysql-5-7-14
      externalName: custom-mysql
      defaultProvisionParameters:
        port: 5000
    ```
1. Save the yaml file and apply it using kubectl:

    ```
    kubectl apply -f custom-plan.yaml
    ```
    
### Modify an existing plan

1. Use kubectl to modify the spec of an existing plan:

    ```
    kubectl edit clusterserviceplan mysql-5-7-14
    ```
1. Add a `defaultProvisionParameters` node to the spec and define the default
    parameters:
                                
    ```yaml
    apiVersion: servicecatalog.k8s.io/v1beta1
    kind: ClusterServicePlan
    metadata:
      name: 5-7-14
    spec:
      clusterServiceBrokerName: minibroker
      externalID: mysql-5-7-14
      externalName: 5-7-14
      defaultProvisionParameters:
        port: 5000
    ```
1. Save the updated resource definition.

## Provision a service instance with default parameters

Once you have a plan with default provision parameters set, provision an instance
of that plan:

```console
$ svcat provision mydb --class mysql --plan custom-mysql
  Name:        mydb
  Namespace:   default
  Status:
  Class:       mysql
  Plan:        custom-mysql

Parameters:
  No parameters defined

$ svcat describe instance mydb
  Name:        mydb
  Namespace:   default
  Status:      Ready - The instance was provisioned successfully @ 2018-09-11 20:26:58 +0000 UTC
  Class:       mysql
  Plan:        custom-mysql

Parameters:
  port: 5000
```

Note that the service instance initially did not have any parameters defined, 
but after it was provisioned it has the parameters defined on the custom
service plan that we created above.
