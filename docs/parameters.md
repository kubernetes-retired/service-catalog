# Passing parameters to brokers

Table of Contents
- [Overview](#overview)
- [Design](#design)
  - [Basic example](#basic-example)
  - [Passing sensitive data](#passing-sensitive-data)
  - [Value types](#value-types)
    - [String](#string)
    - [JSON object](#json-object)
    - [Other primitive types (int, bool, float)](#other-primitive-types-int-bool-float)
    - [Arrays](#arrays)
  - [Passing all parameters from a single source at once](#passing-all-parameters-from-a-single-source-at-once)
  - [Conflict resolution](#conflict-resolution)

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
    - name: ami_id
      value: ami-ecb68a84
```

### Passing sensitive data

To pass a parameter value stored in a `Secret` or a `ConfigMap`, use the `valueFrom` 
field:
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

### Value types

#### String

By default any parameter value is treated as a string, and passed to the underlying 
broker "as is", without further validation or processing. If the user needs to pass
a value of different JSON data type, the field `type` is required to explicitly
specify data type.

#### JSON object

`type: json` field value declares a value as a JSON object.
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
      # Explicitly tells that the value is a valid JSON which should be 
      # passed as a JSON object to the broker
      type: json 
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

If the JSON object is stored in a `Secret` or a `ConfigMap` key, the combination of 
`valueFrom` and `type: json` fields allows to declare this as well:
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
      # Explicitly tells that the value is a valid JSON which should be 
      # passed as a JSON object to the broker
      type: json
      valueFrom:
        secretKeyRef:
          name: mysecret
          key: credentials
```

#### Other primitive types (int, bool, float)

```yaml
  ...
  parameters:
    - name: text
      value: 1 # string by default, will be sent as "1"
    - name: number
      type: int
      value: 1 # will be sent without quotes
    - name: flag
      type: bool
      value: true
    - name: fl
      type: float
      value: 123.5
```

#### Arrays
A separate field `array: true` can be used to declare a value as an array. The
element type is defined by the `type` value as described above.

```yaml
  ...
  parameters:
    - name: tags
      array: true
      value:
      - foo
      - bar
    - name: jsonArray
      type: json
      array: true
      value: [{"a": 1, "b": 2}]
```

### Passing all parameters from a single source at once

In addition to support for explicitly specifying every parameter value fetched from 
a particular source, Service Catalog also supports passing multiple parameters at 
once using a `parametersFrom` field (similar to how `envFrom` works for `Pod`s).
Instead of specifying a reference to a single `Secret` (or `ConfigMap`) key, you 
just set the reference to the `Secret` (or `ConfigMap`), and all keys will be 
automatically processed as separate parameters.

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
you can just set it once in a `parametersFrom` field:
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

Parameter values could be passed as JSON objects, the same way as for [individual parameters](#json-object):
```yaml
  ...
  parametersFrom:
    - type: json
      secretRef:
        name: mysecret
```

### Conflict resolution

Within the scope of a `parameters` or a `parametersFrom` block, if there are 
duplicate top-level names then the last one in each section will take precedence.
If there are duplicate top-level names across `parameters` and `parametersFrom` 
blocks then the `parameters` always takes precedence over `parametersFrom`.

This conflict resolution strategy is consistent with the one supported by Kubernetes
for environment variables (`env` and `envFrom` fields) in the `Pod` resource.

For example, for the structure like this
```yaml
  ...
  parameters:
    - name: param
      value: 1
    - name: param
      valueFrom:
        secretKeyRef:
          name: mysecret
          key: key
    - name: param
      value: 2
  parametersFrom:
    - name: param
      secretRef:
        name: mysecret
```
the parameters JSON structure to be sent to the broker is going to be
```json
{
  "param": "2"
}
```

To avoid conflicts for `parametersFrom` (for example, in the case when multiple 
secrets of the same structure need to be passed to the broker), the user can specify 
a `name` field which will produce an extra nesting level.

For example, let's say we have the following structure in a `Secret`:
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

To inject credentials from two instances of such a `Secret`, we can avoid the 
naming conflict by providing unique names:
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