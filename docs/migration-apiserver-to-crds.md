---
title: Migration from API server to CRDs
layout: docwithnav
---

Service Catalog upgrade from version 0.2.x (and earlier) to 0.3.x needs a data migration. 
This document describes how the migration works and what actions must be performed. 

![Service Catalog upgrade](images/sc-migration-to-crds.svg)

The above picture describes changes in the Service Catalog architecture made between versions 0.2.0 and 0.3.0:
- Custom Resource Definitions (native K8S feature) are now used to store Service Catalog objects 
- etcd and Aggregated API Server are no longer needed
- Webhook Server was added to perform data validation/mutation using the admission webhooks mechanism

## Upgrade Service Catalog as a Helm release

The Service Catalog Helm release can be upgraded using the `helm upgrade` command, which runs all necessary actions.

![Service Catalog upgrade steps](images/sc-migration-to-crds-steps.svg)

The upgrade to CRDs consists of the following steps:
1. Make API Server read-only. Before any backup, we should block any resource changes to be sure the backup makes a snapshot. We need to avoid any changes when the migration tool is backuping resources.
2. Check if Apiserver deployment with a given name exist. **If deployment was not found we skip the migration**.
3. Scale down the Controller Manager to avoid resources processing, such as Secret deletion.
4. Backup Service Catalog custom resources to files in a Persistent Volume.
5. Remove `OwnerReference` fields in all Secrets pointed by any ServiceBinding. This is needed to avoid Secret deletion.
6. Remove all Service Catalog resources. This must be done if the Service Catalog uses the main Kubernetes etcd instance.
7. Upgrade the Service Catalog: remove API Server, install CRDs, Webhook Server and roll up the Controller Manager.
8. Scale down the Controller Manager to avoid any resource processing while applying resources.
9. Restore all resources. The migration tool sets all necessary fields added in the Service Catalog 0.3.0. 
Creating resources triggers all logic implemented in webhooks so we can be sure all data are consistent.
ServiceInstances are created and then updated because of class/plan references fields. 
The validation webhooks denies creating ServiceInstances if the reference to ClusterServiceClass or ServiceClass is not set in following fields:
Spec.ClusterServiceClassRef, Spec.ClusterServicePlanRef, Spec.ServiceClassRef, Spec.ServicePlanRef.
These fields are set during an update operation.
10. Add proper `OwnerReference` to all Secrets pointed by ServiceBindings.
11. Scale up the Controller Manager. 

>**NOTE:** In step 6, there is no difference between Service Catalog upgrade using your own etcd or the main Kubernetes etcd.

## Upgrade Service Catalog manually

### Backup and deleting resources

Execute the `backup` action to scale down the Controller Manager, remove owner references in Secrets and store all resources in a specified folder, then delete all Service Catalog resources.

```bash
./service-catalog migration --action backup --storage-path=data/ --service-catalog-namespace=catalog --controller-manager-deployment=catalog-catalog-controller-manager --apiserver-deployment=catalog-catalog-apiserver
```

### Upgrade

Uninstall old Service Catalog and install the new one (version 0.3.0).

### Restore

Execute `restore action` to restore all resources and scale up the Controller Manager.

```bash
./service-catalog migration --action restore --storage-path=data/ --service-catalog-namespace=catalog --controller-manager-deployment=catalog-catalog-controller-manager
```

## Migration tool

### Build
To run the migration tool, compile the `service-catalog` binary by executing the following command:
```bash
make build
```

If you run the migration tool on OSX and want to get a native binary, add the `PLATFORM` environment variable:
```bash
PLATFORM=darwin make build
```

Resulting executable file can be found in the `bin` subdirectory.

### Execution

You can run the `service-catalog` binary with the `migration` parameter which triggers the migration process. For example, run:

```bash
./service-catalog migration --action restore --storage-path=data/ --service-catalog-namespace=catalog --controller-manager-deployment=catalog-catalog-controller-manager
```

| Flag   | Description  |
| ------------    | ------------ |
| action | Specifies the action which must be executed. The possible values are `backup` or `restore`.|
| storage-path | Points to a folder where resources will be saved. |
| service-catalog-namespace | Specifies the namespace in which the Service Catalog is installed. |
| controller-manager-deployment | Provides the Controller Manager deployment name. |
| apiserver-deployment | Provides the Apiserver deployment name. It is required only for the `backup` phase. |

## Troubleshooting

In case your migration job failed, you can check its logs using the following command:

```bash
kubectl logs -n catalog -l migration-job=true
```

## Cleanup

You can delete all the migration-related resources using this command:

```bash
kubectl delete pvc,clusterrole,clusterrolebinding,serviceaccount,job -n catalog -l migration-job=true
```
