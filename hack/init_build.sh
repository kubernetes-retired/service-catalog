#!/bin/bash
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

set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
. "${ROOT}/hack/utilities.sh" || { echo 'Cannot load bash utilities.'; exit 1; }

VERSIONS="/tmp/.versions"
GO_VERSION='1.7.3'
HELM_VERSION='v2.0.0'

function update-golang() {
  # Check version of golang
  local current_go_version="$(cat "${VERSIONS}/go" || echo "unknown")"

  if [[ "${current_go_version}" == "${GO_VERSION}" ]]; then
    echo "Golang is up-to-date: ${current_go_version}"
  else
    echo "Upgrading golang ${current_go_version} to ${GO_VERSION}"

    # Install new golang.
    local golang_url='https://storage.googleapis.com/golang'
    rm -rf /usr/local/go \
      && curl -sSL "${golang_url}/go${GO_VERSION}.linux-amd64.tar.gz" \
         | tar -C /usr/local -xz \
      || { echo "Cannot upgrade golang to ${GO_VERSION}"; return 1; }
  fi
  return 0
}


function update-helm() {
  # Check version of Helm
  local current_helm_version="$(cat "${VERSIONS}/helm" || echo "unknown")"

  if [[ "${current_helm_version}" == "${HELM_VERSION}" ]]; then
    echo "Helm is up-to-date: ${current_helm_version}"
  else
    echo "Upgrading Helm ${current_helm_version} to ${HELM_VERSION}"

    # Install new Helm.
    local helm_url='https://storage.googleapis.com/kubernetes-helm'
    curl -sSL "${helm_url}/helm-${HELM_VERSION}-linux-amd64.tar.gz" \
        | tar -C /usr/local/bin -xz --strip-components=1 linux-amd64/helm \
      || { echo "Cannot upgrade helm to ${HELM_VERSION}"; return 1; }
  fi
}


function main() {
  update-golang || error_exit 'Failed to update golang'
  update-helm   || error_exit 'Failed to update helm'
}

main
