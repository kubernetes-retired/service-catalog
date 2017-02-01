#!/bin/bash
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

set -o errexit
set -o nounset
set -o pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
export PATH=${ROOT}/contrib/hack:${PATH}

start-server.sh

# create a few resources
kubectl create -f contrib/examples/apiserver/broker.yaml
kubectl create -f contrib/examples/apiserver/serviceclass.yaml
kubectl create -f contrib/examples/apiserver/instance.yaml
kubectl create -f contrib/examples/apiserver/binding.yaml

kubectl get broker test-broker -o yaml
kubectl get serviceclass test-serviceclass -o yaml
kubectl get instance test-instance -o yaml
kubectl get binding test-binding -o yaml

kubectl delete -f contrib/examples/apiserver/broker.yaml
kubectl delete -f contrib/examples/apiserver/serviceclass.yaml
kubectl delete -f contrib/examples/apiserver/instance.yaml
kubectl delete -f contrib/examples/apiserver/binding.yaml

stop-server.sh
