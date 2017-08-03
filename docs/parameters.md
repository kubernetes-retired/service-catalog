# Passing parameters to brokers

Table of Contents
- [Overview](#overview)
- [Design](#design)
  - [Basic example](#basic-example)
  - [Passing sensitive data](#passing-sensitive-data)
    - [Reference to a secret](#reference-to-a-secret)
    - [The entire payload with sensitive data](#the-entire-payload-with-sensitive-data)
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
field allows to reference the external parameters source.
The following sections describe the two ways in which secrets can be used to 
populate the parameters sent to a broker.

#### Reference to a secret

Use the `secretRef` field to pass a reference to a secret with parameter contents.

```yaml
apiVersion: servicecatalog.k8s.io/v1alpha1
kind: Instance
metadata:
  name: qwerty-instance
  namespace: test-ns
spec:
  serviceClassName: qwerty
  planName: default
  parametersFrom:
    - secretRef:
        name: mysecret
```

Each secret key will be transformed into a top-level parameter name, and 
corresponding data will be represented as a value for this parameter.

Note that every value is always treated as a **string**, other value types are 
not supported for `secretRef`.

#### The entire payload with sensitive data

If the user has the entire JSON payload prepared to be sent, and it 
contains sensitive data, this payload can be stored in a single `Secret` key, and 
passed using a `secretKeyRef` field:

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

**one-two-secret**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: map-secret
type: Opaque
stringData:
  oneSecret: a
  twoSecret: b
```

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
  - secretRef:
      name: one-two-secret
  - secretKeyRef:
      name: blob-secret
      key: blob
```

### Parameters payload to be passed to the broker

```json
{
  "inlineField": "abc",
  "sampleLabels": ["foo", "bar"],
  "oneSecret": "a",
  "twoSecret": "b",
  "blobSecretString": "text",
  "blobSecretObj": {
    "json": true
  }
}
```
