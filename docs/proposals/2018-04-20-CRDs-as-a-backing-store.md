# Use CRDs as a Backing Store

Proposed: 2018-04-20
Approved: 

## Abstract

There is a need to be able to deploy Service Catalog without the overhead of an
additional etcd cluster to manage.


## Motivation

The Google Cloud SRE team has made it a requirement for Service Catalog to not
have a second etcd cluster for managed Kubernetes. We use an API Server that
does not have direct access to the main k8s etcd cluster and the SRE team is
unwilling to grant access to Service Catalog OOB.


Alternatives to using CRDs are not quite ready yet, such as the idea for
Kubernetes to provide persistants for API Servers. This work is in the schedule
but not staffed, and could change. The work to move to CRDs entirely is
interesting and will be explored independently of this proposal.


## Constraints and Assumptions

 - CRDs as a backing store will be opt-in.
 - No changes to the API Server or current resources required.
 - One Custom Resource per Service Catalog resource. 
 - Service Catalog will only produce relatively small amount of CRs. (100's)
 - The contents of a CR will be relatively small. (~=10kBs or smaller)
 - Service Catalog uses CRDs as a backing store a constrained amount of time, <2 years.

## Proposed Design

Service Catalog will take advantage of storing json representations of the current
service catalog resources as Custom Resources (via CRDs). This might come from an
external package that allows API servers to do this generically. Service Catalog
will operate exactly the same with no noticeable changes based on this implementation.

rest.Storage is the interface that the API server uses to communicate with etcd. More 
research needs to happen to work out the exact details of the Implementation. There 
was a previous effort to do this with TPRs and a few PRs to do it with an original version of 
CRDs. Not sure why this was abandoned. 

TODO(n3wscott): MORE DETAILS.

## Issues and History

 - Current Issue requesting [CRDs as a Storage Backend](https://github.com/kubernetes-incubator/service-catalog/issues/1088).
 - [[WIP] CRD storage support](https://github.com/kubernetes-incubator/service-catalog/pull/1105) (Reopen / good starting place.)
 - [WIP: complete rewrite of tpr-based storage](https://github.com/kubernetes-incubator/service-catalog/pull/612) (closed, not merged)