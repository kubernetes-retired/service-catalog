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

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly TMP_DIR=$(mktemp -d)

source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }
source "${CURRENT_DIR}/lib/deps_ver.sh" || { echo 'Cannot load dependencies versions.'; exit 1; }

readonly SVC_CATALOG_BUCKET=${SVC_CATALOG_BUCKET:-svc-catalog-charts}
readonly SVC_CATALOG_REPO_URL=https://${SVC_CATALOG_BUCKET}.storage.googleapis.com/

shout "- Setup Helm"
export INSTALL_DIR=${TMP_DIR} HELM_VERSION=${STABLE_HELM_VERSION}
install::local::helm
helm init --client-only

shout "- Install and configure gcloud"
sudo apt-get install -y python
export CLOUD_SDK_VERSION=204.0.0
curl -LO "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-$CLOUD_SDK_VERSION-linux-x86_64.tar.gz"
tar xzf "google-cloud-sdk-$CLOUD_SDK_VERSION-linux-x86_64.tar.gz"
rm "google-cloud-sdk-$CLOUD_SDK_VERSION-linux-x86_64.tar.gz"
rm -rf /google-cloud-sdk/.install/.backup
export PATH="$(pwd)/google-cloud-sdk/bin:$PATH"
gcloud version
gcloud config set core/disable_usage_reporting true
gcloud config set component_manager/disable_update_check true

shout "- Authenticate to Google Cloud Storage"
gcloud auth activate-service-account --key-file contrib/hack/ci/assets/gcloud-key-file.json

shout "- Create the repository"
SVC_CATALOG_REPO_DIR=svc-catalog-repo
mkdir -p ${SVC_CATALOG_REPO_DIR}
pushd ${SVC_CATALOG_REPO_DIR}

  gsutil cp gs://${SVC_CATALOG_BUCKET}/index.yaml .
  helm dep build ../charts/catalog-v0.2
  helm package ../charts/catalog-v0.2
  helm repo index --url ${SVC_CATALOG_REPO_URL} --merge ./index.yaml .
  gsutil -m rsync ./ gs://${SVC_CATALOG_BUCKET}/

popd

ls -l ${SVC_CATALOG_REPO_DIR}
