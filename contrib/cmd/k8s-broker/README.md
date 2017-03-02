# Service Broker

This is an implementation of a service broker which runs in the Kubernetes
cluster and deploys Kubernetes native resources only using Helm as the reifier.

You need to have a [helm](https://github.com/kubernetes/helm) binary and tiller
running in the cluster and need to point this broker to it.

## Create a type for Helm chart

```shell

curl -X POST -d @- localhost:8001/services <<__EOF__
{
  "name": "nginx",
  "id": "4179E70A-4641-49D5-B395-A8ACB1419BCA",
  "description": "Helm chart for running nginx",
  "plans": [
    {
      "name": "nginx",
      "id": "696AD474-123F-474F-8FDB-C724C058CF03",
      "metadata": {
        "instanceType": "gs://helm-sb-test/nginx-0.1.0.tgz"
      }
    }
  ],
  "requires": [],
  "tags": [],
  "bindable":false
}
__EOF__

```


## Create an instance of it

```shell

curl -X PUT -d @- localhost:8002/v2/service_instances/E7075981-6A2A-4FFC-B91C-E056F6CC9671 <<__EOF__
{
  "service_id": "4179E70A-4641-49D5-B395-A8ACB1419BCA",
  "plan_id":"696AD474-123F-474F-8FDB-C724C058CF03"
}
__EOF__

```
