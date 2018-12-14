# Test Service Broker

Test Service Broker is an example
[Open Service Broker](https://www.openservicebrokerapi.org/)
for manually testing & demonstrating the Kubernetes
Service Catalog.

For more information,
[visit the Service Catalog project on github](https://github.com/kubernetes-incubator/service-catalog).

## Installing the Chart

To install the chart with the release name `test-broker`:

```bash
$ helm install charts/test-broker --name test-broker --namespace test-broker
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
| `image` | Image to use | `quay.io/kubernetes-service-catalog/test-broker:v0.1.38` |
| `imagePullPolicy` | `imagePullPolicy` for the test-broker | `Always` |

Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be
provided while installing the chart. For example:

```bash
$ helm install charts/test-broker --name test-broker --namespace test-broker \
  --values values.yaml
```
