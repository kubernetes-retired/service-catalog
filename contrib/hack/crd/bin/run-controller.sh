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

readonly ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

go run ${ROOT_PATH}/../../../../cmd/service-catalog/main.go controller-manager \
--secure-port="8444" \
--cluster-id-configmap-namespace="default" \
--leader-elect="false" \
-v="6" \
--resync-interval="5m" \
--broker-relist-interval="24h" \
--operation-polling-maximum-backoff-duration="20m" \
--k8s-kubeconfig="${KUBECONFIG}" \
--service-catalog-kubeconfig="${KUBECONFIG}" \
--cert-dir="${ROOT_PATH}/../../../../tmp/" \
--feature-gates="OriginatingIdentity=true" \
--feature-gates="ServicePlanDefaults=false"