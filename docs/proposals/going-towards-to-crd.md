# Service Catalog migration to CRDs

In the Service Catalog, we would like to replace the Aggregated API Server with the CustomResourceDefinitions (CRDs) solution. This document presents all the concerns and possible solutions regarding the combination of two approaches (supporting both Aggregated API Server and CRDs), and the standalone CRDs solution.

## Support for both Aggregated API Server and CRDs in the same code-base

Below you can find our concerns about having a single executable binary that supports both Aggregated API Server and CRDs in the same code-base, and changing the behavior by using a feature gate. 

### Business concerns

Adding CRDs via a feature flag directly in a Helm chart can be misleading for the client. 
The CRDs approach is not a feature. It is just a new implementation for already existing features. The current [`apiserver.storage.type`: crds/etcd](https://github.com/kubernetes-incubator/service-catalog/blob/master/charts/catalog/values.yaml#L61-L62) configuration indicates that etcd is rather expanding than deprecating. We do not want to make customers think that **CRD** or **etcd** are storage options. Customers should know that the etcd is depreciating and that they should consider switching for CRDs approach as soon as possible.  

Moreover, fixing bugs and adding new features both in Aggregated API Server and CRDs may slow down the development process.   

Another problem is to provide consistent and up-to-date documentation for both solutions. It may be misleading for customers to see that we are going in two directions at once.

### Technical concerns

We need to clearly state that in the Service Catalog, Aggregated API Server and CRDs are not only about the underlying storage backend. Around those approaches, we have business logic. Because of that, we will end up with a lot of `if` statements in the following areas:

- controller reconcile process 

  |                                   | Aggregated API Server               | CRDs                                                                                      |
  |-----------------------------------|-------------------------------------|-------------------------------------------------------------------------------------------|
  | Queries                           | uses FieldSelector                  | use LabelSelector, cause the CRD does not support queries via Fields                      |
  | Removing finalizes                | in `UpdateStatus` method            | in `Update` method                                                                        |
  | ServiceInstance references fields | via custom `reference` sub-resource | do not support the generic sub-resources, setting them directly via `Update` method |

- `svcat` CLI 

  |         | Aggregated API Server | CRDs              |
  |---------|--------------------|-------------------|
  | Queries | uses FieldSelector | use LabelSelector |
  
- unit tests coverage - we had to also adjust unit tests because we have a slightly different approach, as you can see above. If we want to have one code base, we need to double tests or support both flows in each test.

- validation of incoming Custom Resources (CRs)

  | Aggregated API Server                                 | CRDs              |
  |-------------------------------------------------------|-------------------|
  | `ValidateUpdate` methods and some validation in plugins | Unified and performed only via **ValidatingWebhookConfiguration** |

- defaulting fields of incoming CRs

  | Aggregated API Server                                               | CRDs                                                           |
  |---------------------------------------------------------------------|----------------------------------------------------------------|
  | `PrepareForUpdate` methods and default schemas via `defaulter-gen` | Unified and performed only via **MutatingWebhookConfiguration** |

- defining services in Helm charts - different RBACs, secrets, deployments, services, etc.

- underlying constraint of different Kubernetes versions

As you can see above, the Aggregated API Sever use the validation and mutation in a different way. Sometimes the same logic is split across different pkgs. In CRDs, we unified that and adjusted/copied the validation and mutation logic directly to the webhook domain. If we want to support both the API Server and CRDs concepts, we must invest some time to extract this logic to some common libraries. Then, in both places use generic validation and mutation with an overhead of adjusting the generic interfaces to a custom one. 
Additionally, after removing the API Server, we need to migrate them back to the domain and make concrete method interfaces. Having this logic extracted will only mess the code as the additional abstraction layer can confuse other developers.
 
> **NOTE:** The described differences are those that we've noticed from the general walkthrough. Finally, there may be even more differences, especially if we will support adding features both for API Server and CRDs.
 
**From the technical point of view, having both the Aggregated API Server and CRDs in the same code-base is not a feature gate only in the `service-catalog` binary but in the whole ecosystem.** 

## Alternative CRDs solution

Before merging the CRDs solution into the master, create the `release-0.1` branch with the latest release of the Service Catalog with Aggregated API Server. On the `release-0.1` branch, we still **fix bugs** for this version of the Service Catalog but we do not introduce any new features. When any bugs will be found, we can fix them and still easily create a release with the `0.1.x` version. 

In the master branch, we introduce the Service Catalog with the CRDs approach. We fix bugs and add new features there. New releases are created with the `0.2.x` version.

Such strategy will help us get new CRDs customers fast and will show that the CRDs solution is our new direction.
 
On the other hand, existing Service Catalog customers will see that the old solution is deprecating and will exist only till a specific date (e.g. January 2020), and that since then all new features will be available with CRDs. Thanks to that, they will consider updating their system to the newest version as soon as possible because the goal is to use the CRDs solution recommended by the Kubernetes community.

We can set a reasonable time for supporting bug fixes for the Service Catalog with Aggregated API Server (e.g. Jan 2020).

The described solution mitigates concerns from the previous section.
     
## Migration Support

We need to create a migration guide/scripts to convince existing customers to use the new Service Catalog version. The migration needs to be as simple as possible.

The migration will be simpler if we will support only bug fixing in Aggregated API Server.   

### Details

The migration logic can be placed in the Service Catalog Helm chart. 
Customers just need to run `helm upgrade {release-name} svc-cat/catalog`.

Raw scenario:
- **pre-upgrade hook**
  - replace the api-server deployment with the api-server that has a read-only mode
  - backup the Service Catalog resources to the persistent volume
- **upgrade**
  - remove the api-server
  - remove the etcd storage
  - adjust secrets, RBAC etc.
  - upgrade controller manager
  - install webhook server
- **post-upgrade**
  - scale down the controller manager to 0  
  - restore the Service Catalog resources - Spec and Status (status is important because we do not want to trigger provisioning for already process items)
  - scale up controller manager


## Conclusions

We need to be sure that the customer will exactly know in which direction we want to go and why. Supporting only bug fixing for API Server will simplify adding new features and will give customers time for migration. The migration process will be easier when we will stop developing features in Aggregated API Server solution. For new contributors, it will be much easier to get familiar with the code base where only one approach is used. 

From our perspective, supporting both approaches at the same time, with bug fixing and new features, will just postpone the whole process. Sooner or later, we will have to take all these steps and back to the discussion. The later we migrate, the more difficult the whole process will become (technical debt, confused customers, etc.).

What we need:
- set a date (e.g. Jan 2020) when API Server will be erased from the Service Catalog repository 
- communicate via Service Catalog SIG and all channels about the decision as soon as possible, provide CRDs solution and migration guide beforehand
