# Declarative Delete Field

Concept committed to [during 2018 April 12F2F](https://docs.google.com/document/d/1O7_fws7hwZ6qV3okAjbV5qdFaEq4bt1LYJ9bTu-IwvU)

This document 2018-05.

## Abstract

Add a field to instance and binding to indicate that they have been asked to
send a DELETE to the backend.

## Motivation

When the Originating-Identity header is used, the broker and refuse a delete
call when the user info in the Originating-Identity header does not have
permissions to delete the instance or binding. As this happens after kubernetes
has accepted the DELETE as an true indication to do a delete, it has set the
DeletionTimestamp, which is an irreversible step on the way to carrying out the
necessary actions to delete a kubernetes resource.

## Proposed Design

 - Add a field to indicate deletion
 - Add a condition to indicate permanent end state of deletion success.

Once the field is activated the controller deprovisions the broker
resource. When successfully deleted the controller then deletes the
kubernetes resource.

A feature flag to enable/disable this behavior.

RBAC changes to allow the controller to issue deletes.

Helm chart changes to enable the flag and the rbac rules.

### API Resource Changes

Add a condition for ServiceInstanceConditionType, ServiceBindingConditionType.
```go
  ServiceInstanceConditionDeleted ServiceBindingConditionType = "Deleted"
  ServiceBindingConditionDeleted ServiceBindingConditionType = "Deleted"
```

Add a field to spec of ServiceInstance, ServiceBinding.
```go
  ExistsExternally bool `json:"existsExternally,omitempty"`
```

This field is set to true by default. The user indicates that it wants
to delete the backing resource by setting this field to false. The
controller will see this update and attempt to delete the backing
broker resource. If the broker delete succeeds, it sets the deleted
status condition.

If the field is set to false, no updates are allowed besides
setting the field back to true. 

Once the status is changed to deleted the controller should issue the
DELETE to the apiserver removing the resource.

### API Server Changes

Additional field results in generated code changes.

Validation/Strategy to prevent changes after a deletion is started and is successful or
until the deletion is rolled back.

### Controller-Manager Changes

Additional condition to determine delete flow is added. This condition is
dependent on the state of the new field. Reuse all applicable status
objects as appropriate.

Proceed through the normal deletion flow and if the broker allows the delete,
set a new condition indicating the delete has occurred. If the delete has
occurred, when running the finalizer, allow the delete to finish and occur
without any additional actions involving broker communication.



### SVCat CLI Changes

New types result in test input/output changes to golden files.
