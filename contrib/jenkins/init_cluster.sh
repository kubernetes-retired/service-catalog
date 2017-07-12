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

set -o nounset
set -o errexit

while [[ $# -ne 0 ]]; do
  case "$1" in
    --project)     PROJECT="$2"; shift ;;
    --zone)        ZONE="$2"; shift ;;
    --credentials) CREDENTIALS="$2"; shift ;;
    *)             CLUSTERNAME="$1" ;;
  esac
  shift
done

[[ -n "${CLUSTERNAME:-}" ]] \
  || { echo 'Cluster name must be provided.'; exit 1; }

[[ -n "${ZONE:-}" ]] \
  || { echo 'Zone must be provided.'; exit 1; }

[[ -n "${PROJECT:-}" ]]  \
  || { echo 'Project must be provided.'; exit 1; }

[[ -n "${CREDENTIALS:-}" ]] \
  || { echo '--credentials is a required parameter' ; exit 1; }

export GOOGLE_APPLICATION_CREDENTIALS="${CREDENTIALS}"

gcloud auth activate-service-account \
    --key-file="${GOOGLE_APPLICATION_CREDENTIALS}" \
  || { echo "Cannot activate GCloud service account from ${CREDENTIALS}"; exit 1; }

echo "Creating cluster ${CLUSTERNAME}"

gcloud container clusters create "${CLUSTERNAME}" --project="${PROJECT}" --zone="${ZONE}" \
  || { echo 'Cannot create cluster.'; exit 1; }

echo "Using cluster ${CLUSTERNAME}."

gcloud container clusters get-credentials "${CLUSTERNAME}" --project="${PROJECT}" --zone="${ZONE}" \
  || { echo 'Cannot get credentials for cluster.'; exit 1; }

helm init \
  || { echo 'Cannot initialize Helm.'; exit 1; }
