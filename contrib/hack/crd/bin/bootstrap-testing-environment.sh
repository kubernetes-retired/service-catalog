#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -u
set -o errexit

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

echo "- Initialize Minikube"
bash ${CURRENT_DIR}/minikube.sh

echo "- Installing Tiller..."
kubectl apply -f ${CURRENT_DIR}/../assets/tiller.yaml

bash ${CURRENT_DIR}/is-ready.sh kube-system name tiller

echo "- Installing SC"
helm install --name catalog --namespace kyma-system  ${CURRENT_DIR}/../../../../charts/catalog/ --wait

echo "- Installing Pod Preset Helm Chart"
helm install ${CURRENT_DIR}/../assets/pod-preset-chart.tgz  --name podpreset --namespace kyma-system --wait

echo "- Installing Helm Broker Helm Chart"
helm install ${CURRENT_DIR}/../assets/helm-broker-chart.tgz  --name helm-broker --namespace kyma-system --wait
echo "- Installing BUC Helm Chart"
helm install ${CURRENT_DIR}/../assets/buc-chart.tgz  --name buc --namespace kyma-system --wait

echo "- Register Helm Broker in Service Catalog"
kubectl apply -f  ${CURRENT_DIR}/../assets/helm-broker.yaml

echo "- Scale down controller manager"
kubectl -n kyma-system scale deploy --replicas=0 catalog-catalog-controller-manager

echo "- Expose Helm Broker to localhost on port 8081"
export HB_POD_NAME=$(kubectl get po -l app=helm-broker -n kyma-system -o jsonpath='{ .items[*].metadata.name }')
kubectl port-forward -n kyma-system pod/${HB_POD_NAME} 8081:8080
