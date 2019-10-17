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

echo "DEPLOY_TYPE ${DEPLOY_TYPE}"

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly REPO_ROOT_DIR=${CURRENT_DIR}/../../../

source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

export REGISTRY=${REGISTRY:-quay.io/kubernetes-service-catalog/}

docker login -u "${QUAY_USERNAME}" -p "${QUAY_PASSWORD}" quay.io

usage() {
  echo "${0} [-a arch] [-h]"
  echo ""
  echo " -a, --arch <arch-name>: Architecture name that should be released, i.e. amd64"
  echo " -h, --help: Print help"
}

RELEASE_ARCH=""
if [[ "$#" -ne 0 ]]; then
  POSITIONAL=()
  while [[ $# -gt 0 ]]
    do
    key="$1"
    case ${key} in
      -b|--arch)
        RELEASE_ARCH="-${2}"
        shift # past argument
        shift # past value
        ;;
      -h|--help)
        usage
        exit
        ;;
      *)    # unknown option
        POSITIONAL+=("$1") # save it in an array for later
        shift # past argument
        ;;
    esac
  done
  set -- ${POSITIONAL[@]+"${POSITIONAL[@]}"}
fi

pushd ${REPO_ROOT_DIR}
echo ${RELEASE_ARCH}
if [[ "${DEPLOY_TYPE}" == "release" ]]; then
    shout "Pushing Service Catalog ${RELEASE_ARCH} images with tags '${TRAVIS_TAG}' and 'latest'."
    TAG_VERSION="${TRAVIS_TAG}" VERSION="${TRAVIS_TAG}" MUTABLE_TAG="latest" make release-push${RELEASE_ARCH}
elif [[ "${DEPLOY_TYPE}" == "master" ]]; then
    shout "Pushing Service Catalog ${RELEASE_ARCH} images with default tags (git sha and 'canary')."
    make push
else
    shout "Skip Service Catalog ${RELEASE_ARCH} deploy"
fi

popd
