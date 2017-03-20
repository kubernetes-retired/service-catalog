# Service Catalog Broker Resource

This chart contains a single binding resource intended for use by the service-catalog
system running in Third Party Resource mode.

For more information, [visit the Service Catalog project on github]
(https://github.com/kubernetes-incubator/service-catalog).

## Installing the Chart

To install the chart with the release name `binding-resource`:

```bash
$ helm install --name binding-resource --namespace test-ns charts/resources/binding
```

## Uninstalling the Chart

To uninstall/delete the `binding-resource` release:

```bash
$ helm delete binding-resource
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of this chart.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `name` | The name of the binding. | `test-binding` |
| `instanceRef.name` | The name of the instance this binding should bind to. | `test-instance` |
| `secretName` | The name of the secret to which the service-catalog should write bind credentials. | `test-secret` |

Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be
provided while installing the chart. For example:

```bash
$ helm install --name binding-resource --namespace test-ns  --values values.yaml charts/resources/binding
```
