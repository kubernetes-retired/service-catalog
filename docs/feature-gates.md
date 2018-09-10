---
title: Feature Gates
layout: docwithnav
---

## Overview

Feature gates are a set of key=value pairs that describe alpha or experimental
features and can be turned on or off by specifying the key=value pair in the
arguments list when launching the Service Catalog executable.  A new geature
gate should be created when introducing new features that may break existing
functionality or introduce instability.

The following table is a summary of the feature gates that you can set on
different Service Catalog.

- The "Since" column contains the Service Catalog release when a feature is
  introduced or its release stage is changed.
- The "Until" column, if not empty, contains the last Service Catalog release in
  which you can still use a feature gate.

| Feature | Default | Stage | Since | Until |
|---------|---------|-------|-------|-------|
| `AsyncBindingOperations` | `false` | Alpha | v0.1.7 | |
| `NamespacedServiceBroker` | `false` | Alpha | v0.1.10 | v0.1.28 |
| `NamespacedServiceBroker` | `true` | GA | v0.1.29 | |
| `OriginatingIdentity` | `false` | Alpha | v0.1.7 | v0.1.29 |
| `OriginatingIdentity` | `true` | GA | v0.1.30 | |
| `OriginatingIdentityLocking` | `true` | Alpha | v0.1.14 | |
| `PodPreset` | `false` | Alpha | v0.1.6 | |
| `ResponseSchema` | `false` | Alpha | v0.1.12 | |
| `ServicePlanDefaults` | `false` | Alpha | v0.1.32 | |
| `UpdateDashboardURL` | `false` | Alpha | v0.1.13 | |


## Using a Feature

### Feature Stages

A feature can be in *Alpha*, *Beta* or *GA* stage.
An *Alpha* feature means:

* Disabled by default.
* Might be buggy. Enabling the feature may expose bugs.
* Support for feature may be dropped at any time without notice.
* The API may change in incompatible ways in a later software release without
  notice.
* Recommended for use only in short-lived testing clusters, due to increased
  risk of bugs and lack of long-term support.

A *Beta* feature means:

* Enabled by default.
* The feature is well tested. Enabling the feature is considered safe.
* Support for the overall feature will not be dropped, though details may change.
* The schema and/or semantics of objects may change in incompatible ways in a
  subsequent beta or stable release. When this happens, we will provide
  instructions for migrating to the next version. This may require deleting,
  editing, and re-creating API objects. The editing process may require some
  thought. This may require downtime for applications that rely on the feature.
* Recommended for only non-business-critical uses because of potential for
  incompatible changes in subsequent releases. If you have multiple clusters
  that can be upgraded independently, you may be able to relax this restriction.

**Note:** Please do try *Beta* features and give feedback on them!
After they exit beta, it may not be practical for us to make more changes.

A *GA* feature is also referred to as a *stable* feature. It means:

* The corresponding feature gate is no longer needed.
* Stable versions of features will appear in released software for many
  subsequent versions.

### Feature Gates

Each feature gate is designed for enabling/disabling a specific feature:

- `AsyncBindingOperations`: Controls whether the controller should attempt
 asynchronous binding operations

- `NamespacedServiceBroker`: Enables namespaced variants of ServiceBrokers,
ServiceClasses, and ServicePlans.

- `OriginatingIdentity`: Controls whether the controller should include
originating identity in the header of requests sent to brokers

- `OriginatingIdentityLocking`:  Controls whether we lock OSB API resources
for updating while we are still processing the current spec.

 - `PodPreset`: Controls whether PodPreset resource is enabled or not in the
 API server.

- `ResponseSchema`:  Enables the storage of the binding response schema in
ServicePlans

- `ServicePlanDefaults`: Enables applying default values to service instances
and bindings

- `UpdateDashboardURL`:  Enables the update of DashboardURL in response to
update service instance requests to brokers.

