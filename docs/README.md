---
title: Documentation
layout: docsportal
cid: docs-home
permalink: /docs/
---

# Service Catalog Documentation

This page is an index for the articles in here. We recommend you start by reading the 
[introduction](./concepts/index.md#introduction), and then move on to the 
[installation instructions](./install.md). After you install, see our
[walkthrough](./walkthrough.md) document to get started using Service Catalog.

Afterward, see the topics below.

**Note:** Between versions 0.2.0 and 0.3.0, Service Catalog changed its internal storage mechanism.
Versions 0.2.0 and older used its own API Server and etcd. 

Starting from version 0.3.0, Service Catalogs moved to a solution based on Custom Resource Definitions, which is a native K8S feature.

The API Server implementation will be supported by providing bug fixes for the next 9 months.
If you still use Service Catalog version 0.2.0, read the [migration guide](./migration-apiserver-to-crds.md).

## Topics for users:

- [Installation instructions](install.md)
- [Walkthrough](walkthrough.md)
- [Service Catalog CLI](cli.md)
- [The Service Catalog Resources In Depth](./resources.md)
- [Passing parameters to ServiceInstances and ServiceBindings](parameters.md)

## Topics for developers:

- [Code conventions](./code-standards.md)
- [Developer guide](./devguide.md)
- [Design](./design.md)
- [Notes on authentication](./auth.md)

## Topics for operators:

- [Using Namespaced Broker Resources](./namespaced-broker-resources.md)
- [Filtering Broker Catalogs](./catalog-restrictions.md)
- [Setting Defaults for Service Instances](./service-plan-defaults.md)

## Request for Comments

As Service Catalog is in beta, we are collecting use-cases as issues.
Please file an issue and bring it up in the weekly meeting.
