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

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=${CURRENT_DIR}/../../../

source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

export REGISTRY=${REGISTRY:-quay.io/kubernetes-service-catalog/}

docker login -u "${QUAY_USERNAME}" -p "${QUAY_PASSWORD}" quay.io

pushd ${REPO_ROOT_DIR}

if [[ "${TRAVIS_TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+[a-z]*(-(r|R)(c|C)[0-9]+)*$ ]]; then
    shout "Pushing images with tags '${TRAVIS_TAG}' and 'latest-v0.2'."
    TAG_VERSION="${TRAVIS_TAG}" VERSION="${TRAVIS_TAG}" MUTABLE_TAG="latest-v0.2" make release-push svcat-publish
elif [[ "${TRAVIS_BRANCH}" == "v0.2" ]]; then
    shout "Pushing images with default tags (git sha and 'canary-v0.2')."
    make push svcat-publish
else
    shout "Nothing to deploy"
fi

popd
