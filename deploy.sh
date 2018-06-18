#!/bin/bash
# kubernetes and catalog expected to be running

echo "=== Building and deploying CRD"
./build-crd-image.sh
kubectl apply -f install.yaml

echo "=== Generating certificates for webhook"
#pushd pkg/controller/podpreset/webhook/pki && ./gen-certs.sh && popd
kubectl create secret tls podpreset-service-tls \
  --cert=pkg/controller/podpreset/webhook/pki/podpreset-service.pem \
  --key=pkg/controller/podpreset/webhook/pki/podpreset-service-key.pem
until [ $(kubectl get pods -n podpreset-crd-system -l api=podpreset-crd -o jsonpath="{.items[0].status.containerStatuses[0].ready}") = "true" ]; do
  sleep 2
done

echo "=== Apply RBAC for webhook"
pushd pkg/controller/podpreset/webhook
kubectl create -f apod-rbac.yaml

echo "=== Building and deploying webhook"
make build
kubectl create -f deployment.yaml
until [ $(kubectl get pods -l app=podpreset -o jsonpath="{.items[0].status.containerStatuses[0].ready}") = "true" ]; do
  sleep 2
done
popd

echo "=== Creating podpresets"
kubectl create -f pkg/controller/podpreset/webhook/apod-preset.yaml
kubectl create -f pkg/controller/podpreset/webhook/apod-deployment-preset.yaml
sleep 2

echo "=== Creating deployment with matching label selector"
kubectl create -f pkg/controller/podpreset/webhook/apod-deployment.yaml

echo "=== Creating pod with matching label selector"
kubectl create -f pkg/controller/podpreset/webhook/apod.yaml

echo "=== Retrieving logs from webhook"
kubectl logs -l app=podpreset

echo "=== Retrieving logs from controller"
kubectl logs -lapi=podpreset-crd -n podpreset-crd-system

###

echo "=== Creating podpreset binding"
kubectl create namespace test-ns
kubectl create -f pkg/controller/podpreset/webhook/apod2-presetbinding.yaml

echo "=== Creating deployment with matching label selector"
kubectl create -f pkg/controller/podpreset/webhook/apod2-deployment.yaml

echo "=== Executing mini-walkthrough"
helm install charts/ups-broker --name ups-broker --namespace ups-broker
kubectl create -f contrib/examples/walkthrough/ups-broker.yaml
kubectl create -f contrib/examples/walkthrough/ups-instance.yaml
kubectl create -f contrib/examples/walkthrough/ups-binding.yaml

echo "=== Retrieving logs from controller"
kubectl logs -lapi=podpreset-crd -n podpreset-crd-system
