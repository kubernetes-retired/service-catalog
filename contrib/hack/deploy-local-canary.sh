#!/usr/bin/env bash
# Copyright 2017 The Kubernetes Authors.
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

set -o nounset
set -o errexit

export REGISTRY=${REGISTRY:=$USER/}
make service-catalog-image service-catalog-push
helm upgrade --install catalog --namespace catalog charts/catalog \
    --recreate-pods --force \
    --set image=${REGISTRY}service-catalog:canary \
    --set imagePullPolicy=Always \
    --set deploymentStrategy=Recreate \
    --set apiserver.storage.etcd.persistence.enabled=true \
    --set rbacEnable=true \
    --set namespacedServiceBrokerDisabled=false \
    --set servicePlanDefaultsEnabled=true \
    --wait
