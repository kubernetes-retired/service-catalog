---
title: Sync a Broker
layout: docwithnav
---

A broker's catalog may occasionaly change the services it offers. A broker
may resync automatically or may need to be resynced manually. By default,
brokers are resynced automatically based on an interval that is globally
set in Service Catalog. If a broker must be resynced immeadiately or
`spec.relistBehavior` has been set to manual, then it can be resynced
manually by incrementing `spec.relistRequests`. This can be done using svcat:
```console
$ svcat sync broker foobar --scope cluster
Synchronization requested for broker: foobar
```
