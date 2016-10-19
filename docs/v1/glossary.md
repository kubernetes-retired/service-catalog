# Service Catalog Glossary

This document defines common terminology that other documents in this
repository commonly use.

- __Consumer__: an application that will use a backing service (see below).
- __Backing service__: the service that a consumer utilizes, after following steps
  to make it available to them. Backing services are commonly, but not
  necessarily hosted off-cluster. They include white-box implementations, such
  as open source software, and black-box implementations, such as SaaS products.
  Examples of both include but are not limited to:
  - Amazon AWS resources (
    [ELB](https://aws.amazon.com/elasticloadbalancing/) instances,
    [RDS](https://aws.amazon.com/rds/) databases,
    [S3](https://aws.amazon.com/s3/) buckets)
  - [Hosted elastic search](https://www.elastic.co/cloud)
  - Open source database systems
  - Open source message queueing/streaming systems
- __Service broker__: a web server system that adheres to the standard CloudFoundry
  [service broker API](https://docs.cloudfoundry.org/services/api.html), and is
  responsible for performing standard operations to make backing services
  available to consumers
- __Cluster operator__: the team of people responsible for operating a Kubernetes
  cluster. This team may operate the cluster on behalf of other users, or may
  operate the cluster to facilitate their own work.
  The cluster operator "team" may be one person
- __Application developer__: the team of people responsible for developing consumer
  applications. This "team" may be one person. Additionally,
  it may overlap (partially or completely) with the cluster operator team
- __Resource type__: a logical Kubernetes concept. Examples include:
  - [Pod](http://kubernetes.io/docs/user-guide/pods/)s
  - [Service](http://kubernetes.io/docs/user-guide/services/)s
  - [Secret](http://kubernetes.io/docs/user-guide/secrets/)s
- __Resource__: a specific instantiation of an aforementioned resource type,
  often represented as a YAML or JSON file that is submitted or retrieved
  via the standard Kubernetes API (or via `kubectl`)
