---
title: Synchronize a Broker
layout: docwithnav
---

A broker's catalog may occasionally change the services it offers. Therefore,
Service Catalog must resynchronize with the broker to get the updated services.
A broker may resynchronize automatically or may need to be resynchronized
manually. By default, brokers are resynchronized automatically based on
the `brokerRelistInterval` global setting in Service Catalog. If a broker must be
resynchronized immediately or if `.spec.relistBehavior` on the broker has been
set to manual, then it can be resynchronized manually by incrementing 
`.spec.relistRequests`. This can be done using svcat:
```console
$ svcat get classes
     NAME      NAMESPACE          DESCRIPTION
+------------+-----------+---------------------------+
  mariadb                  Helm Chart for mariadb
  mongodb                  Helm Chart for mongodb
  mysql                    Helm Chart for mysql
  postgresql               Helm Chart for postgresql
  
$ svcat sync broker foobar --scope cluster
Synchronization requested for broker: foobar

$ svcat get classes
     NAME      NAMESPACE          DESCRIPTION
+------------+-----------+---------------------------+
  mariadb                  Helm Chart for mariadb
  mongodb                  Helm Chart for mongodb
  mysql                    Helm Chart for mysql
  postgresql               Helm Chart for postgresql
  redis                    Helm Chart for redis
```
