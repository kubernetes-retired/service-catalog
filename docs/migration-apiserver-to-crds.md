## Overview

Service Catalog upgrade from version 0.2.x (and earlier) to 0.3.x needs a data migration. This document describes how the migration works and what action must be performed.

## Upgrade Service Catalog as a Helm release

The Service Catalog helm release can be upgraded using `helm upgrade` command, which runs all necessary actions.

### Details of an upgrade and migration

The upgrade to CRDs contains the following steps:
1. Make API Server read only. Before any backup we should block any resource changes to be sure the backup makes a snapshot. We need to avoid any changes while migration tool is backuping resources.
2. Scale down controller manager to avoid resources processing, for example secret deletion.
3. Backup ServiceCatalog custom resources to files in a Persistent Volume.
4. Remove `OwnerReference` fields in all secrets pointed by any ServiceBinding. This is needed to avoid Secret deletion.
5. Remove all Service-Catalog resources. This must be done if Service Catalog uses the main Kubernetes ETCD instance.
6. Upgrade Service-Catalog: remove API Server, install CRDs, webhook and roll up the controller manager.
7. Scale down controller-manager to avoid any resource processing while applying resources.
8. Restore all resources. The migration tool sets all necessary fields added in Service Catalog 0.3.0. Creating resources triggers all logic implemented in webhooks so we can be sure all data are consistent.
Service instances are created and then updated because of class/plan refs fields. The validation webhooks denies creating service instances if the reference to ClusterServiceClass or ServiceClass is not set: Spec.ClusterServiceClassRef, Spec.ClusterServicePlanRef, Spec.ServiceClassRef, Spec.ServicePlanRef.
These fields are set during an update operation.
9. Add proper owner reference to all secrets pointed by service bindings.
10. Scale up controller-manager. 

>Note: There is no difference between upgrade Service Catalog using own ETCD or main Kubernetes ETCD.
## Manual Service Catalog upgrade

### Backup and deleting resources

Execute `backup` action to scale down the controller, remove owner referneces in secrets and store all resources in a specified folder, then delete all Service Catalog resources.

```bash
./service-catalog migration --action backup --storage-path=data/ --service-catalog-namespace=catalog --controller-manager-deployment=catalog-catalog-controller-manager
```

### Upgrade

Uninstall old Service Catalog and install the new one (version 0.3.0).

### Restore

Execute `restore action` to restore all resources and scale up the controller.

```bash
./service-catalog migration --action restore --storage-path=data/ --service-catalog-namespace=catalog --controller-manager-deployment=catalog-catalog-controller-manager

```
## Migration tool

The `service-catalog` binary can be run with `migration` parameter which triggers the migration process, for example:

```bash
./service-catalog migration --action restore --storage-path=data/ --service-catalog-namespace=catalog --controller-manager-deployment=catalog-catalog-controller-manager
```

| flag   | Description  |
| ------------    | ------------ |
| action | Specifies the action which must be executed, can be `backup` or `restore`.|
| storage-path | Points to a folder, where resoruces will be saved. |
| service-catalog-namespace | The namespace, where the Service Catalog is installed. |
| controller-manager-deployment | The controller manager deployment name. |

