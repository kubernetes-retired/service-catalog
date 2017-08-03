# Passing parameters to brokers

Table of Contents
- [Overview](#overview)
- [Design](#design)
  - [Basic example](#basic-example)
  - [Passing sensitive data](#passing-sensitive-data)
  - [Merging multiple sources and conflict resolution](#merging-multiple-sources-and-conflict-resolution)
- [Example with multiple sources](#example-with-multiple-sources)

## Overview
`parameters` and `parametersFrom` properties of `Instance` and `Broker` resources 
provide support for passing parameters to the broker relevant to the corresponding
[provisioning](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#provisioning) or
[binding](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#binding) request. 
The resulting structure represents an arbitrary JSON object, which is assumed to 
be valid for a particular broker. 
The Service Catalog does not enforce any extra limitations on the format and content 
of this structure.

## Design

When you create an `Instance` or a `Binding`, you can set parameters to be passed 
to the corresponding broker.
To set parameters, include the `parameters` or `parametersFrom` field in the spec.

### Basic example

Let's say we want to create an `Instance` of EC2 running on AWS using a
[corresponding broker](https://github.com/cloudfoundry-samples/go_service_broker) 
which implements the Open Service Broker API.

A typical provisioning request for this broker looks [like this](https://github.com/cloudfoundry-samples/go_service_broker/blob/master/bin/curl_broker.sh):
```bash
curl -X PUT http://username:password@localhost:8001/v2/service_instances/instance_guid-111 -d '{
  "service_id":"service-guid-111",
  "plan_id":"plan-guid",
  "organization_guid": "org-guid",
  "space_guid":"space-guid",
  "parameters": {"ami_id":"ami-ecb68a84"}
}' -H "Content-Type: application/json"
```

Note that the broker accepts an `ami_id` parameter ([AMI](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html) 
identifier).
To configure a provisioning request in Service Catalog, we need to declare an `Instance` 
resource with an AMI identifier declared in the `parameters` field of its spec:
```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: Instance
metadata:
  name: ami-instance
  namespace: test-ns
spec:
  serviceClassName: aws-ami
  planName: default
  parameters:
    ami_id: ami-ecb68a84
```

### Passing sensitive data

`Secret` resources can be used to store sensitive data. The `parametersFrom`
field allows the user to reference the external parameters source.

If the user has sensitive data in their parameters, the entire JSON payload can 
be stored in a single `Secret` key, and passed using a `secretKeyRef` field:

```yaml
  ...
  parametersFrom:
    - secretKeyRef:
        name: mysecret
        key: mykey
```

The value stored in a secret key must be a valid JSON.

### Merging multiple sources and conflict resolution

If multiple sources in `parameters` and `parametersFrom` blocks are specified,
the final payload is a result of merging all of them at the top level.
If there are any duplicate properties defined at the top level, the specification
is considered to be invalid, the further processing of the `Instance`/`Binding` 
resource stops and its `status` is marked with error condition.

## Example with multiple sources

### Sources

**blob-secret**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: blob-secret
type: Opaque
stringData:
  blob: >
    {
      "blobSecretString": "text",
      "blobSecretObj": {
        "json": true
      }
    }
```

### Instance specification

```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: Instance
metadata:
  name: qwerty-instance
  namespace: test-ns
spec:
  serviceClassName: qwerty
  planName: default
  parameters:
    inlineField: abc
    sampleLabels:
    - foo
    - bar
  parametersFrom:
  - secretKeyRef:
      name: blob-secret
      key: blob
```

### Parameters payload to be passed to the broker

```json
{
  "inlineField": "abc",
  "sampleLabels": ["foo", "bar"],
  "blobSecretString": "text",
  "blobSecretObj": {
    "json": true
  }
}
```
