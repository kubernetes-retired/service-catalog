# V1 API

This document contains the Kubernetes resource types for the v1 service catalog.

# Resource Types

This section lists descriptions of the resources in the service catalog API.

_ __Note:__ All names herein are tentative, and should be considered placeholders
and used as placeholders for purposes of discussion only. This note will be
removed when names are finalized._

## `Broker`

This resource is created by an administrator to instruct the service catalog's
controller event loop to do the following:

1. Make a request against a given CF service broker's catalog endpoint
   (`GET /v2/catalog`)
2. Translate the response to a list of `ServiceClass`es
3. Write each translated `ServiceClass` to stable storage

TODO: should we call out what happens when a `Broker` resource is deleted?

## `ServiceClass`

This resource is created by the service catalog's controller event loop to
represent a service ID & plan ID that a CF service broker has made available.

## `Instance`

This resource is created by a service consumer to indicate that the service
catalog's controller event loop should provision the backing CF service broker
and write the provision response back into the `Instance`'s `status` section.

## `Binding`

This resource is created by a service consumer to indicate that the service
catalog's controller event loop should issue a bind request on the backing CF
service broker.
