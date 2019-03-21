---
title: Install a Broker
layout: docwithnav
---

A service broker must be added to the catalog before it can be used. This is
done by creating either a ClusterServiceBroker or ServiceBroker object, and
then allowing Service Catalog to sync with the broker. This is typically
referred to as "registering" the broker.

To register a broker, use the `svcat register` command:
```console
$ svcat register foobarbroker --url http://foobarbroker.com
```

By default, this will register the broker as a ServiceBroker object in your
current namespace. To make it available cluster-wide, set the scope to cluster:
```console
$ svcat register foobarbroker --url http://foobarbroker.com --scope cluster
```

Most actual brokers will require authenticaiton of some kind. To add this, use the
`--basic-secret` or `--bearer-secret` flags.
```console
$ svcat register foobarbroker --url http://foobarbroker.com --basic-secret broker-creds
$ svcat register foobarbroker --url http://foobarbroker.com --bearer-secret broker-creds --namespace creds-namespace
```
