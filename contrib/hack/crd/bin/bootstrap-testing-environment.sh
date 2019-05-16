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
helm install --name catalog --namespace default  ${CURRENT_DIR}/../../../../charts/catalog/ --wait

echo "- Installing test-broker"
helm install --name ups-broker --namespace ups-broker  ${CURRENT_DIR}/../../../../charts/ups-broker --wait

echo "- Registering test-broker"
kubectl apply -f ${CURRENT_DIR}/../../../examples/walkthrough/ups-clusterservicebroker.yaml