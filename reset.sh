#!/bin/bash -x
#kubectl delete -f pkg/controller/podpreset/webhook/apod-rbac.yaml

kubectl delete -f pkg/controller/podpreset/webhook/apod.yaml
kubectl delete -f pkg/controller/podpreset/webhook/apod-preset.yaml

kubectl delete -f pkg/controller/podpreset/webhook/apod-deployment.yaml
kubectl delete -f pkg/controller/podpreset/webhook/apod-deployment-preset.yaml

kubectl delete -f pkg/controller/podpreset/webhook/apod2-presetbinding.yaml
kubectl delete -f pkg/controller/podpreset/webhook/apod2-deployment.yaml

# mini-walkthrough cleanup
kubectl delete clusterservicebrokers ups-broker
helm delete --purge ups-broker
kubectl delete ns test-ns ups-broker
