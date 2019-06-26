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

export SC_NAMESPACE="catalog"
SC_CHART_NAME="catalog"
SC_CHART_VERSION="0.2.1"
export SC_APISERVER="${SC_CHART_NAME}-catalog-apiserver"
export SC_CONTROLLER="${SC_CHART_NAME}-catalog-controller-manager"

TB_CHART_NAME="test-broker"
TB_CHART_VERSION="0.2.1"
export TB_NAME="${TB_CHART_NAME}-test-broker"
export TB_NAMESPACE="test-broker"

export APP_KUBECONFIG_PATH="${HOME}/.kube/config"

echo "- Initialize Minikube"
bash ${CURRENT_DIR}/scripts/minikube.sh

echo "- Installing Tiller..."
helm init --wait

echo "- Installing ServiceCatalog"
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
helm install svc-cat/catalog --version ${SC_CHART_VERSION} --name ${SC_CHART_NAME} --namespace ${SC_NAMESPACE} --wait

echo "- Installing Test broker"
helm install svc-cat/test-broker --version ${TB_CHART_VERSION} --name ${TB_CHART_NAME} --namespace ${TB_NAMESPACE} --wait

echo "- Prepare test resources"
go run ${CURRENT_DIR}/examiner/main.go --action prepareData

echo "- Upgrade ServiceCatalog"
helm upgrade ${SC_CHART_NAME} ${CURRENT_DIR}/../../charts/catalog --namespace ${SC_NAMESPACE} --wait

echo "- Execute upgrade tests"
go run ${CURRENT_DIR}/examiner/main.go --action executeTests
