#!/bin/bash
kubectl delete -f pkg/controller/podpreset/webhook/apod-rbac.yaml

kubectl delete -f pkg/controller/podpreset/webhook/apod.yaml
kubectl delete -f pkg/controller/podpreset/webhook/apod-preset.yaml
kubectl delete -f pkg/controller/podpreset/webhook/deployment.yaml

kubectl delete -f pkg/controller/podpreset/webhook/apod-deployment.yaml
kubectl delete -f pkg/controller/podpreset/webhook/apod-deployment-preset.yaml

kubectl delete -f install.yaml

kubectl delete secret podpreset-service-tls

# mini-walkthrough
kubectl delete -n test-ns servicebindings ups-binding
kubectl delete -n test-ns serviceinstances ups-instance
kubectl delete clusterservicebrokers ups-broker
helm delete --purge ups-broker
kubectl delete ns test-ns ups-broker
