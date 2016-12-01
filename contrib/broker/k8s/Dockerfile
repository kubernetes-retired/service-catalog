# Copyright 2016 The Kubernetes Authors.
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

FROM debian

ENV HELM_VERSION=v2.0.0

RUN export DEBIAN_FRONTEND=noninteractive \
 && apt-get update \
 && apt-get install -y \
    ca-certificates \
    curl \
 && curl -Ls -o /tmp/helm-${HELM_VERSION}-linux-amd64.tar.gz \
    http://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz \
 && mkdir -p /opt/services \
 && tar -C /opt/services -x -z -f /tmp/helm-${HELM_VERSION}-linux-amd64.tar.gz \
    --strip-components=1 linux-amd64/helm \
 && rm -f /tmp/helm-${HELM_VERSION}-linux-amd64.tar.gz \
 && apt-get purge -y --auto-remove curl \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/opt/services/k8s-broker", "--helm_binary", "/opt/services/helm"]
