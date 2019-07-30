---
title: CLI
layout: docwithnav
---

This is a command-line interface (CLI) for interacting with Service Catalog
resources. svcat is a domain-specific tool to make interacting with the Service Catalog easier.
While many of its commands have analogs to `kubectl`, our goal is to streamline and optimize
the operator experience.

svcat communicates with Kubernetes cluster by directly using REST API - just like kubectl.

This document assumes that you've installed Service Catalog and the Service Catalog CLI
onto your cluster. If you haven't, please see the [installation instructions](install.md#installing-the-service-catalog-cli).

## Plugin Mode
To use svcat as a kubectl plugin, run the following command after downloading:

```console
$ svcat install plugin
Plugin has been installed to ~/.kube/plugins/svcat. Run kubectl plugin svcat --help for help using the plugin.
```

When operating as a plugin, the commands are the same with the addition of the global
kubectl configuration flags. One exception is that boolean flags aren't supported
when running in plugin mode, so instead of using `--flag` you must specify a value `--flag=true`.

# Use

Run `svcat --help` to see the available commands.

Below are some common tasks made easy with svcat. The example output assumes that the
[User Provided Service Broker](../charts/ups-broker) is installed on the cluster.

## Register a broker
```console 
$ svcat register ups-broker --url http://ups-broker-ups-broker.ups-broker.svc.cluster.local --scope cluster
  Name:     ups-broker                                                  
  URL:      https://ups-broker-ups-broker.ups-broker.svc.cluster.local  
  Status:  
```

## Find brokers installed on the cluster

This lists all brokers available in the current namespace and at the cluster scope.

```console
$ svcat get brokers
     NAME      NAMESPACE                              URL                              STATUS  
+------------+-----------+-----------------------------------------------------------+--------+
  ups-broker               http://ups-broker-ups-broker.ups-broker.svc.cluster.local   Ready   
```

## Trigger a sync of a broker's catalog

```console
$ svcat sync broker ups-broker
Synchronization requested for broker: ups-broker
```

## List available service classes

This lists all classes available in the current namespace and at the cluster scope.
```console
$ svcat get classes
                 NAME                  NAMESPACE         DESCRIPTION        
+------------------------------------+-----------+-------------------------+
  user-provided-service                            A user provided service  
  user-provided-service-single-plan                A user provided service  
  user-provided-service-with-schemas               A user provided service  
  ```

## See all services offered in the current namespace and at the cluster scope.
```console
$ svcat marketplace
                CLASS                   PLANS          DESCRIPTION        
+------------------------------------+---------+-------------------------+
  user-provided-service                default   A user provided service  
                                       premium                            
  user-provided-service-single-plan    default   A user provided service  
  user-provided-service-with-schemas   default   A user provided service 
```

## Provision a service

```console
$ svcat provision ups-instance --class user-provided-service --plan default
  Name:        ups-instance           
  Namespace:   default 
  Status:                             
  Class:       user-provided-service  
  Plan:        default                

Parameters:
  No parameters defined
```

Additional parameters and secrets can be provided using the `--param` and `--secret` flags:

```
--param p1=foo --param p2=bar --secret creds[db]
```

You can also provide provision parameters in the form of a JSON string using the `--params-json` flag:

```console
$ svcat provision secure-instance --class user-provided-service --plan premium --params-json '{
    "encrypt" : true,
    "firewallRules" : [
        {
            "name": "AllowSome",
            "startIPAddress": "75.70.113.50",
            "endIPAddress" : "75.70.113.131"
        },
        {
            "name": "AllowMore",
            "startIPAddress": "13.54.0.0",
            "endIPAddress" : "13.56.0.0"
        }
    ]
}
'
Name:        secure-instance
Namespace:   default
Status:
Class:       user-provided-service
Plan:        premium

Parameters:
  encrypt: true
  firewallRules:
  - endIPAddress: 75.70.113.131
    name: AllowSome
    startIPAddress: 75.70.113.50
  - endIPAddress: 13.56.0.0
    name: AllowMore
    startIPAddress: 13.54.0.0
```

Note: You may not combine the `--params-json` flag with individual `--param` flags.


## List all service instances in a namespace

```console
$ svcat get instances
      NAME       NAMESPACE           CLASS            PLAN     STATUS  
+--------------+-----------+-----------------------+---------+--------+
  ups-instance   default     user-provided-service   default   Ready 
```

## Bind an instance

```console
$ svcat bind ups-instance --name ups-binding
  Name:        ups-binding   
  Namespace:   default       
  Status:                    
  Secret:      ups-binding   
  Instance:    ups-instance  

Parameters:
  No parameters defined
```

When omitted, the names of the binding and secret are defaulted to the name of the instance.

```console
$ svcat bind ups-instance
  Name:        ups-instance
  Namespace:   default
  Status:
  Instance:    ups-instance
```

## View the details of a service instance

```console
$ svcat describe instance ups-instance
  Name:        ups-instance                                                                       
  Namespace:   default                                                                            
  Status:      Ready - The instance was provisioned successfully @ 2018-11-01 18:31:16 +0000 UTC  
  Class:       user-provided-service                                                              
  Plan:        default                                                                            

Parameters:
  No parameters defined

Bindings:
     NAME       STATUS  
+-------------+--------+
  ups-binding   Ready 
```

## Remove all bindings from an instance

```console
$ svcat unbind ups-instance
deleted ups-binding
```

## Remove a single binding from an instance

```console
$ svcat unbind  --name ups-binding
deleted ups-binding
```

## Delete a service instance

Deprovisioning is the process of preparing an instance to be removed, and then deleting it.
You must unbind delete all bindings before deprovisioning an instance.

```console
$ svcat deprovision ups-instance
deleted ups-instance
```

## Deregister a broker
Deregistering is the process of removing a broker and its associated classes and plans from the cluster.
You must delete all active instances of its classes before deregistering a broker.
```console
$ svcat deregister ups-broker --scope cluster
Successfully removed broker "ups-broker"
```

# Namespaced Resource Support

svcat supports interaction with the namespaced versions of Service Catalog resources. The `scope` flag is
used to indicate whether to execute the command in the `cluster`, `namespace`, or `all` scope. 

## Register a Namespaced Broker

svcat registers/deregisters brokers as namespaced brokers in the namespace of your current context by default.
You can choose another namespace with the `-n` flag.
```console
$ svcat register ups-broker --url http://ups-broker-ups-broker.ups-broker.svc.cluster.local -n foobar
  Name:     ups-broker                                                 
  URL:      http://ups-broker-ups-broker.ups-broker.svc.cluster.local  
  Status:
  ```

## Provision an Instance of a Namespaced Class/Plan

svcat does not currently support provisioning instances of namespaced classes and plans.

## Querying State

svcat commands that get information about the current state of the system also support the `--scope` flag, which is set
to `all` by default. This will return all of the resources visible in the cluster scope, and the namespace of your current
context. In the following examples, [minibroker](https://github.com/osbkit/minibroker) is installed as a cluster broker, while `ups-broker` is installed in the
`default` namespace.

```console
$ svcat get brokers
     NAME      NAMESPACE                              URL                              STATUS
+------------+-----------+-----------------------------------------------------------+--------+
  minibroker               http://minibroker-minibroker.minibroker.svc.cluster.local   Ready
  ups-broker   default     http://ups-broker-ups-broker.ups-broker.svc.cluster.local   Ready
$ svcat get brokers -n foobar
       NAME      NAMESPACE                              URL                              STATUS
+------------+-----------+-----------------------------------------------------------+--------+
minibroker               http://minibroker-minibroker.minibroker.svc.cluster.local   Ready
$ svcat get brokers --scope namespace
     NAME      NAMESPACE                              URL                              STATUS
+------------+-----------+-----------------------------------------------------------+--------+
  ups-broker   default     http://ups-broker-ups-broker.ups-broker.svc.cluster.local   Ready
```

## Describing a Namespaced Resource

`svcat describe` does not currently support namespaced resources.
