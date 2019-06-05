# Test Service Broker

Test Service Broker is an example
[Open Service Broker](https://www.openservicebrokerapi.org/)
for manually testing & demonstrating the Kubernetes
Service Catalog.

For more information,
[visit the Service Catalog project on github](https://github.com/kubernetes-sigs/service-catalog).

## Installing the Chart

To install the chart with the release name `test-broker`:

```bash
$ helm install charts/test-broker --name test-broker --namespace test-broker
```

## Registering the broker

To use the broker, register it with any of the following commands.

Using `svcat`: 

```bash
$ svcat register test-broker --url http://test-broker-test-broker.test-broker.svc.cluster.local
```

Using kubectl:
```bash
$ kubectl apply -f contrib/examples/walkthrough/test-broker/test-clusterservicebroker.yaml
```

To register the broker into your namespace instead, use:
```bash
$ kubectl apply -f contrib/examples/walkthrough/test-broker/test-servicebroker.yaml
```

## Uninstalling the Chart

To uninstall/delete the `test-broker` deployment:

```bash
$ helm delete test-broker
```

The command removes all the Kubernetes components associated with the chart and
deletes the release.

## Configuration

The following tables lists the configurable parameters of the Test
Service Broker

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image` | Image to use | `quay.io/kubernetes-service-catalog/test-broker:v0.2.1` |
| `imagePullPolicy` | `imagePullPolicy` for the test-broker | `Always` |

Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be
provided while installing the chart. For example:

```bash
$ helm install charts/test-broker --name test-broker --namespace test-broker \
  --values values.yaml
```
