# Passing parameters to the brokers

Table of Contents
- [Overview](#overview)
- [Design](#design)
  - [Basic example](#basic-example)
  - [Passing sensitive data](#passing-sensitive-data)
  - [Passing parameter value as a JSON object](#passing-parameter-value-as-a-json-object)
  - [Passing all parameters from a single source at once](#passing-all-parameters-from-a-single-source-at-once)
  - [Conflict resolution](#conflict-resolution)

## Overview
`parameters` and `parametersFrom` properties of `Instance` and `Broker` resources provide support for passing 
parameters to the broker relevant to the corresponding
[provisioning](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#provisioning) or
[binding](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#binding) request. 
The resulting structure represents an arbitrary JSON object which is assumed to be valid for a particular broker. 
Service Catalog does not enforce any extra limitations on the format and content of this structure.

## Design

When you create an `Instance` or a `Binding`, you can set parameters to be passed to the corresponding broker.
To set parameters, include the `parameters` or `parametersFrom` field in the spec.

### Basic example

Let's say we want to create an `Instance` of EC2 running on AWS using 
[corresponding broker](https://github.com/cloudfoundry-samples/go_service_broker) which implements Open Service Broker API.

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

Note that broker accepts `ami_id` parameter ([AMI](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html) identifier).
To configure a provisioning request in Service Catalog, we need to declare an `Instance` resource with specifying AMI 
identifier in the `parameters` field of its spec:
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
    - name: ami_id
      value: ami-ecb68a84
```

### Passing sensitive data

To pass a parameter value stored in a `Secret` or a `ConfigMap`, use `valueFrom` field:
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
    - name: username
      value: admin
    - name: password
      valueFrom:
        secretKeyRef:
          name: mysecret
          key: password
```

### Passing parameter value as a JSON object

By default any parameter value is treated as a string, and passed to the underlying broker "as is", without further
validation or processing. If the user needs to pass a JSON object, it can be achieved by explicitly setting the `type: json` field.
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
    - name: txt
      value: Plain text, passed as a string
    - name: meta
      type: json # Explicitly tells that the value is JSON which should be validated and passed as a JSON object
      value: >
          {
            "attributes": [
                {
                  "type": "a",
                  "A": "1",
                },
                {
                  "type": "b",
                  "B": "2",
                },
            ],
            "labels": ["foo", "bar"]
          }
```

If the JSON object is stored in a `Secret` or `ConfigMap` key, the combination of `valueFrom` and `type: json` fields 
allows to declare this as well:
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
    - name: txt
      value: Plain text, passed as a string
    - name: meta
      type: json # Explicitly tells that the value is JSON which should be validated and passed as a JSON object
      valueFrom:
        secretKeyRef:
          name: mysecret
          key: password
```

### Passing all parameters from a single source at once

In addition to support for explicitly specifying every parameter value fetched from a particular source, Service Catalog
also supports passing multiple parameters at once using `parametersFrom` field (similar to how `envFrom` works for `Pod`s).
Instead of specifying a reference to a single `Secret` 
(or `ConfigMap`) key, you just set the reference to the `Secret` (or `ConfigMap`), and all keys will be automatically
processed as separate parameters.

In other words, instead of 
```yaml
  ...
  parameters:
    - name: key1
      valueFrom:
        secretKeyRef:
          name: mysecret
          key: key1
    - name: key2
      valueFrom:
        secretKeyRef:
          name: mysecret
          key: key2
     ...
```
you can just set it once in `parametersFrom` field:
```yaml
  ...
  parametersFrom:
    - secretRef:
        name: mysecret
```

You can also pass several `Secret`s and/or `ConfigMap`s at once:
```yaml
  ...
  parametersFrom:
    - secretRef:
        name: secret1
    - secretRef:
        name: secret2
    - configMapRef:
        name: configmap1
```

Parameter values could be passed as JSON objects, the same way as for [individual parameters]((#passing-parameter-value-as-a-json-object)):
```yaml
  ...
  parametersFrom:
    - type: json
      secretRef:
        name: mysecret
```

### Conflict resolution

When a key exists in multiple sources defined in `parametersFrom` or `parameters` field, the value associated with 
the last source will take precedence.
Values defined in `parameters` with a duplicate key will take precedence over `parametersFrom`.

To avoid conflicts for `parametersFrom` (for example, in the case when multiple secrets of the same structure need
to be passed to the broker), the user can specify `name` field which will produce an extra nesting level.

For example, let's say we have a following `Secret` structure:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm
```

To inject credentials from two instances of such `Secret`, we can resolve the conflict by providing unique names:
```yaml
  ...
  parametersFrom:
    - name: master
      secretRef:
        name: secret1
    - name: slave
      secretRef:
        name: secret2
```
which will result in the following JSON parameters object:
```json
{
  "master": {
    "username": "root",
    "password": "letmein"
  },
  "slave": {
    "username": "foo",
    "password": "bar"
  }
}

```