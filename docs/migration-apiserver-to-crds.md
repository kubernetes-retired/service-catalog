---
title: Migration from API server to CRDs
layout: docwithnav
---

Service Catalog upgrade from version 0.2.x (and earlier) to 0.3.x needs a data migration. 
This document describes how the migration works and what actions must be performed. 

>**NOTE:**
Before starting the migration, make sure that you performed a full backup of your cluster.
You should also test the procedure on a testing environment first.

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

Migration tool is a set of helper functions integrated into the Service Catalog binary.

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

### Implementation details

In order to get a consistent backup, we have to make sure that no resources are modified during the backup process.
To achieve that, the migration tool creates `ValidatingWebhookConfiguration` at the beginning of the backup process 
to intercept and reject all attempts to mutate Service Catalog resources. Because of the limitation of the Aggregated API Server used in the previous version of Service Catalog, this webhook call fails with the following message:
```bash
failed calling webhook "validating.reject-changes-to-sc-crds.servicecatalog.k8s.io": 
webhook does not accept v1beta1 AdmissionReviewRequest
```
This error message is presented in case of a modification or creation attempt of any Service Catalog resource during the backup process, and it means that the write protection works as expected.

To test the mutation blocking feature, execute the following commands:
- to enable the write protection:
  ```bash
  ./service-catalog migration --action deploy-blocker --service-catalog-namespace=default
  ```
- to disable the write protection:
  ```bash
  ./service-catalog migration --action undeploy-blocker --service-catalog-namespace=default
  ```

## Cleanup

You can delete all the migration-related resources using this command:

```bash
kubectl delete pvc,clusterrole,clusterrolebinding,serviceaccount,job -n catalog -l migration-job=true
```

## Troubleshooting

In case your migration job failed, you can check its logs using the following command:

```bash
kubectl logs -n catalog -l migration-job=true
```

## Rollback

In case you want to revert the upgrade, use the `helm rollback` command which will restore the Service Catalog API Server version.

Before you proceed, you must delete all the Service Catalog resources and CRDs. You must also delete the resources that are not necessary for the Service Catalog API Server version. Use the following commands:
```
kubectl delete crd -l svcat=true
kubectl delete secret -n catalog catalog-catalog-webhook-cert
kubectl delete sa -n catalog service-catalog-webhook
kubectl delete sa -n catalog clean-job-account
```

Then you can execute the rollback using this command:
```
helm rollback catalog 1 --cleanup-on-fail --no-hooks
```

After the rollback is succeeded, you still have the backup of your Service Catalog resources from the previous upgrade stored in the persistence volume.  
