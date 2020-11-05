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

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.

readonly CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly TMP_DIR=$(mktemp -d)
readonly TEMP_CHART_DIR=svc-catalog-repo
readonly SVC_CATALOG_REPO_URL=https://github.com/kubernetes-sigs/service-catalog/blob/gh_pages/charts_archive/

declare COMMIT_MESSAGE="Publish the chart tarballs"
if [[ "$(git describe --tags)" =~ ^v[0-9]+(\.[0-9]+){1,2}$ ]]; then
  COMMIT_MESSAGE="${COMMIT_MESSAGE} for $(git describe --tags)"
fi

source "${CURRENT_DIR}/lib/utilities.sh" || {
  echo 'Cannot load CI utilities.'
  exit 1
}

source "${CURRENT_DIR}/lib/deps_ver.sh" || {
  echo 'Cannot load dependencies versions.'
  exit 1
}

setup_helm() {
  shout "- Setup Helm"

  export INSTALL_DIR=${TMP_DIR} HELM_VERSION=${STABLE_HELM_VERSION}
  install::local::helm
  helm init --client-only
}

create_tarballs() {
  shout "- Create the chart tarballs"

  mkdir -p ${TEMP_CHART_DIR}
  pushd ${TEMP_CHART_DIR}
  for dir in ../charts/*; do
    helm dep build "${dir}"
    helm package "${dir}"
  done
  for chart in *.tgz; do
    chart_name=${chart/-[0-9.]*tgz/}
    mkdir "$chart_name"
    mv "$chart" "$chart_name/"
  done
  popd
}

reconfigure_git() {
  git config remote.origin.fetch '+refs/heads/*:refs/remotes/origin/*'
  git fetch --all
  git config remote.origin.url "$(git config --get remote.origin.url | sed 's/https:\/\/github.com\//git@github.com:/')"
  git config --local core.sshCommand "/usr/bin/ssh -i ${CURRENT_DIR}/assets/chart_key"
}

publish() {
  shout "- Publish the new charts"

  reconfigure_git
  git checkout gh_pages
  cp index.yaml "${TEMP_CHART_DIR}"

  pushd "${TEMP_CHART_DIR}"
  helm repo index --url "${SVC_CATALOG_REPO_URL}" --merge ./index.yaml .
  sed 's/\.tgz$/.tgz?raw=true/' ./index.yaml >../index.yaml
  git add ../index.yaml
  for package in **/*.tgz; do
    dest_dir="../charts_archive/${package/\/[a-z0-9.-]*tgz/}"
    mkdir -p "$dest_dir"
    mv "$package" "../charts_archive/${dest_dir}/"
    git add "../charts_archive/${package}"
  done
  git commit -m "$COMMIT_MESSAGE"
  git push
  popd
}

main() {
  setup_helm
  create_tarballs
  publish
}

main
