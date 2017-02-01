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
set -x

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
export PATH=${ROOT}/contrib/hack:${PATH}

# Make our kubectl image, if not already there
make-kubectl.sh

# Clean up old containers if still around
docker rm -f etcd apiserver > /dev/null 2>&1 || true

# Start etcd, our DB
docker run -ti --name etcd -d --net host quay.io/coreos/etcd > /dev/null

# And now our API Server
docker run -ti --name apiserver -d --net host \
	-v ${ROOT}:/go/src/github.com/kubernetes-incubator/service-catalog \
	-v ${ROOT}/.var/run/kubernetes-service-catalog:/var/run/kubernetes-service-catalog \
	-v ${ROOT}/.kube:/root/.kube \
	scbuildimage \
	bin/apiserver -v 10 --etcd-servers http://localhost:2379 > /dev/null

# Wait for apiserver to be up and running
while ! curl -k http://localhost:6443 > /dev/null 2>&1 ; do
	sleep 1
done

# Setup our credentials
kubectl config set-credentials service-catalog-creds --username=admin --password=admin
kubectl config set-cluster service-catalog-cluster --server=https://localhost:6443 --certificate-authority=/var/run/kubernetes-service-catalog/apiserver.crt
kubectl config set-context service-catalog-ctx --cluster=service-catalog-cluster --user=service-catalog-creds
kubectl config use-context service-catalog-ctx
