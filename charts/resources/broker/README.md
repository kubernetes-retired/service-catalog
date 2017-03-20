# Service Catalog Broker Resource

This chart contains a single broker resource intended for use by the service-catalog
system running in Third Party Resource mode.

For more information, [visit the Service Catalog project on github]
(https://github.com/kubernetes-incubator/service-catalog).

## Installing the Chart

To install the chart with the release name `broker-resource`:

```bash
$ helm install --name broker-resource --namespace test-ns charts/resources/broker
```

## Uninstalling the Chart

To uninstall/delete the `broker-resource` release:

```bash
$ helm delete broker-resource
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of this chart.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `brokerURL` | The URL of the broker server | `http://doesnt-exist.com` |
| `name` | The name of the broker server | `test-broker` |
| `auth.secretName` | The name of the secret that contains basic authentication credentials. This secret should have a `username` and `password` field | `broker-secret` |
| `auth.username` | The username that the broker should use to authenticate clients | `broker-username` |
| `auth.password` | The password that the broker should use to authenticate clients | `broker-password` |

Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be
provided while installing the chart. For example:

```bash
$ helm install --name broker-resource --namespace test-ns  --values values.yaml charts/resources/broker
```
