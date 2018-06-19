#!/bin/bash
kubectl delete -f pkg/controller/podpreset/webhook/apod-rbac.yaml

kubectl delete -f pkg/controller/podpreset/webhook/apod.yaml
kubectl delete -f pkg/controller/podpreset/webhook/apod-preset.yaml
kubectl delete -f pkg/controller/podpreset/webhook/deployment.yaml

kubectl delete -f pkg/controller/podpreset/webhook/apod-deployment.yaml
kubectl delete -f pkg/controller/podpreset/webhook/apod-deployment-preset.yaml

kubectl delete -f install.yaml

# mini-walkthrough
kubectl delete -n test-ns servicebindings ups-binding
sleep 10
kubectl delete -n test-ns serviceinstances ups-instance
sleep 10
kubectl delete clusterservicebrokers ups-broker
sleep 10
helm delete --purge ups-broker
kubectl delete ns test-ns ups-broker

kubectl delete secret podpreset-service-tls
