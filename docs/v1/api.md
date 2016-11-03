# V1 API

This document contains resource types and their standard usage in v1 service
catalog. Although this API will be implemented in Kubernetes, other systems
are not precluded from implementing it as well.

# Resource Types

This section lists descriptions of the resources in the service catalog API.

*__Note:__ All names herein are tentative, and should be considered placeholders
and used as placeholders for purposes of discussion only. This note will be
removed when names are finalized.*

## `Broker`

This resource is created by an administrator to instruct the service catalog's
controller event loop to do the following:

1. Make a request against a given CF service broker's catalog endpoint
   (`GET /v2/catalog`)
2. Translate the response to a list of `ServiceClass`es
3. Write each translated `ServiceClass` to stable storage

*TODO: should we call out what happens when a `Broker` resource is deleted?*

## `ServiceClass`

This resource is created by the service catalog's controller event loop after
it has received a `Broker` resource and successfully called the backing CF
broker's catalog endpoint. It represents a service ID & plan ID that a CF
service broker has made available.

*TODO: what happens in the below cases?*

1. *The `Broker` that caused the `ServiceClass` was deleted*
1. *The `ServiceClass` itself was deleted*

## `Instance`

This resource is created by a service consumer to indicate their desire to
provision a service. When the service catalog's controller event loop sees an
`Instance` created, it calls the provision endpoint on the backing CF service
broker and writes `provisioned` into the `status.status` field of the
`Instance`.

*TODO: what happens when an `Instance` resource is deleted?*

## `Binding`

This resource is created by a service consumer to indicate that an application
should be bound to an instance. When the service catalog's controller event
loop sees a `Binding` created, it calls the bind endpoint on the backing CF
service broker. When a successful response is returned, it does the following:

1. Writes `bound` into the `status.status` field of the `Binding`
1. Writes the contents of `credentials` map into a secret (naming of the secret
   to be discussed later)
1. Updates the `Binding`'s status section with the fully qualified path to the
   aforementioned secret

*TODO: what happens when a `Binding` resource is deleted?*
