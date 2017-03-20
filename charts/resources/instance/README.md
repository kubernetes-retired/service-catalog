# Service Catalog Broker Resource

This chart contains a single instance resource intended for use by the service-catalog
system running in Third Party Resource mode.

For more information, [visit the Service Catalog project on github]
(https://github.com/kubernetes-incubator/service-catalog).

## Installing the Chart

To install the chart with the release name `instance-resource`:

```bash
$ helm install --name instance-resource --namespace test-ns charts/resources/broker
```

## Uninstalling the Chart

To uninstall/delete the `instance-resource` release:

```bash
$ helm delete instance-resource
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of this chart.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `name` | The URL of the broker server. | `test-instance` |
| `serviceClassName` | The name of the service class this instance should provision. | `test-service-class` |
| `planName` | The name of the plan this instance should provision. | `test-plan` |

Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be
provided while installing the chart. For example:

```bash
$ helm install --name instance-resource --namespace test-ns  --values values.yaml charts/resources/instances
```
