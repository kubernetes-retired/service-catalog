# Health Check

HealthCheck is a tool that can be used to monitor the basic health of the Service Catalog deployment.  It utilizes the User Provided Service Broker to perform basic end to end tests such as creating a Service Instance and Binding and verifying the operations are successful and then tearing it down.  It collects Prometheus metrics which can be scraped for analysis and alerting (ie error rate, execution time to create an instance or binding, etc).

For more information,
[visit the Service Catalog project on github](https://github.com/kubernetes-incubator/service-catalog).

## Installing the Chart

To install the chart with the release name `healthcheck`:

```bash
$ helm install charts/healthcheck --name healthcheck --namespace healthcheck --set imagePullPolicy=Never --set image=healthcheck:canary
```

## Uninstalling the Chart

To uninstall/delete the `healthcheck` deployment:

```bash
$ helm delete healthcheck
```

The command removes all the Kubernetes components associated with the chart and
deletes the release.

## Configuration

The following tables lists the configurable parameters of the HealthCheck

TBD


Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be
provided while installing the chart. For example:

```bash
$ helm install charts/healthcheck --name healthcheck --namespace healthcheck \
  --values values.yaml
```
