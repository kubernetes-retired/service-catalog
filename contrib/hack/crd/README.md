## Cookbook for CRDs POC 

Execute all commands from the cookbook in the `hack` directory.

### Bootstrap local environment for testing

1. In one shell execute:
```bash
./bin/bootstrap-testing-environment.sh
```

Under the hood this script is:
- creating minikube
- installing tiller
- installing Service Catalog
- installing Helm Broker
- installing Binding Usage Controller
- registering Helm Broker in Service Catalog with http://localhost:8081
- exposing the Helm Broker to your localhost on port 8081, so your controller in step 2 can access all broker endpoints 

2. When step one is finished then on the other shell execute:
```bash
./bin/run-controller.sh
```

**Now you are ready to go!**

When you execute `svcat get classes`, then you should see:
```bash
          NAME           NAMESPACE                 DESCRIPTION
+----------------------+-----------+------------------------------------------+
  azure-service-broker               Extends the Service Catalog with Azure
                                     services
  redis                              Redis by Helm Broker (Experimental)
  gcp-service-broker                 Extends the Service Catalog with Google
                                     Cloud Platform services
``` 

### Testing Scenario

Follow these steps:

1. Export the name of the Namespace.
```bash
export namespace="qa"
```
2. Create a Redis instance.
```bash
kubectl create -f assets/scenario/redis-instance-manual.yaml -n $namespace
```
3. Check if the Redis instance is already provisioned.
```bash
watch -n 1 "kubectl get serviceinstance/redis -n $namespace -o jsonpath='{ .status.conditions[0].reason }'"
```
4. Create Secrets for the Redis instance.
```bash
kubectl create -f assets/scenario/redis-instance-binding-manual.yaml -n $namespace
```
5. Create a deploy.
```bash
kubectl create -f assets/scenario/redis-client.yaml -n $namespace
```
6. Create a Binding Usage with **APP_** prefix.
```bash
kubectl create -f assets/scenario/service-binding-usage.yaml -n $namespace
```
7. Wait until the Pod is ready.
```bash
kubectl get po -l app=redis-client -n $namespace -o jsonpath='{ .items[*].status.conditions[?(@.type=="Ready")].status }'
```
8. Export the name of the Pod.
```bash
export POD_NAME=$(kubectl get po -l app=redis-client -n $namespace -o jsonpath='{ .items[*].metadata.name }')
```
9. Execute the `check-redis` script on the Pod.
```bash
kubectl exec ${POD_NAME} -n $namespace /check-redis.sh
```

The information and statistics about the Redis server appear.


### Documentation

- [Design of the Service Catalog](https://svc-cat.io/docs/design/)
- [Service Catalog Developer Guide](https://svc-cat.io/docs/devguide/)
- [Service Catalog Code & Documentation Standards](https://svc-cat.io/docs/code-standards/)


### Old way of running controller locally

#### Prerequisites

Kyma installed on your cluster but without the ServiceCatalog.

#### Steps

1. Install ServiceCatalog chart
```bash
helm install --name catalog --namespace kyma-system  charts/catalog/ --wait
```

2. Register Helm Broker
```bash
kubectl apply -f ./assets/helm-broker.yaml
```

3. Export the name of the HelmBroker Pod.
```bash
export HB_POD_NAME=$(kubectl get po -l app=helm-broker -n kyma-system -o jsonpath='{ .items[*].metadata.name }')
```

4. Expose helm-broker service
```bash
kubectl port-forward -n kyma-system pod/${HB_POD_NAME} 8081:8080
```

5. Scale down controller manager 

```bash
kubectl -n kyma-system scale deploy --replicas=0 catalog-catalog-controller-manager
```

6. Run the Service Catalog controller-manager
```bash
./bin/run-controller.sh
```
