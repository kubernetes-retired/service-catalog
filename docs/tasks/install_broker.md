---
title: Install a Broker
layout: docwithnav
---

A service broker must be added to the catalog before it can be used. This is
done by creating either a Cluster Service Broker or Service Broker (also called
a namespaced broker) object, and then allowing Service Catalog to sync with
the broker. A Service Broker exposes the broker's available services to users
in a single namespace, while a Cluster Service Broker makes them available
to all users in all namespaces. By default, `svcat register` creates
namespaced Service Brokers:
```console
$ svcat register foobarbroker --url http://foobarbroker.com
```

To register a broker as a Cluster Service Broker, set the scope to cluster:
```console
$ svcat register foobarbroker --url http://foobarbroker.com --scope cluster
```

Most actual brokers will require authentication of some kind. To add this, use the
`--basic-secret` or `--bearer-secret` flags.
```console
$ svcat register foobarbroker --url http://foobarbroker.com --basic-secret broker-creds
$ svcat register foobarbroker --url http://foobarbroker.com --bearer-secret broker-creds --namespace creds-namespace
```
