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


ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
. "${ROOT}/script/utilities.sh" || { echo 'Cannot load Bash utilities'; exit 1; }

BUCKET=gs://helm-sb-test
TEMP="$(mktemp -d)" \
  || error_exit 'Cannot create a temporary directory'

pushd "${TEMP}"

for chart in $(find "${ROOT}/registry/data/charts" \( ! -regex '\.\.*$' \) -type d -mindepth 1 -maxdepth 1); do
  chart_name=$(basename ${chart})

  if [[ ! -e "${chart}/Chart.yaml" ]]; then
    echo "Skipping ${chart_name} because Chart.yaml is not found."
    continue
  fi

  version=$(perl -n -e'/version: ([0-9.]+)/ && print $1' < "${chart}/Chart.yaml")
  if [[ $? -ne 0 ]]; then
    echo "Skipping ${chart_name} because cannot determine version."
    continue
  fi

  if ! helm package "${chart}"; then
    echo "Skipping ${chart_name} because cannot helm package."
    continue
  fi

  TGZ="${chart##*/}-${version}.tgz"

  if [[ ! -e "./${TGZ}" ]]; then
    echo "Skipping ${chart_name} because cannot find Helm package."
    continue
  fi

  if ! gsutil ls "${BUCKET}/${TGZ}" &>/dev/null ; then
    echo "Pushing ${chart_name}:${version}..."

    gsutil cp "./${TGZ}" "${BUCKET}/" \
      || echo "Failed to copy ${TGZ}"

    if [[ -e "${chart}/schema.yaml" ]]; then
      gsutil cp "${chart}/schema.yaml" "${BUCKET}/${TGZ}.schema" \
        || echo "Failed to copy ${chart}/schema.yaml"
    else
      echo "Skipping upload of ${chart}/schema.yaml - file doesn't exist."
    fi
  else
    echo "Skipping ${chart_name}:${version} because it already exists."
  fi
done

popd
rm -rf ${TEMP}

