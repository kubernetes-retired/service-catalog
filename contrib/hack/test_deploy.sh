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

. "${ROOT}/contrib/hack/utilities.sh" || { echo 'Cannot load Bash utilities'; exit 1; }

${ROOT}/contrib/hack/deploy.sh "${@}" \
  || error_exit 'Deployment to Kubernetes cluster failed.'

IP="$(kubectl get services | xargs echo -n | sed 's/.*booksfe [0-9.]* \([0-9.]*\).*/\1/')"

${ROOT}/contrib/hack/bookstore_client.py --host="${IP}:8080" --api_key=123 \
    --verify --verbose=true --count=1 \
  || error_exit 'Tests failed.'
