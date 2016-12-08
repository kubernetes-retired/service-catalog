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

. "${ROOT}/hack/cluster_utilities.sh" || { echo 'Cannot load cluster utilities'; exit 1; }
. "${ROOT}/hack/utilities.sh" || { echo 'Cannot load Bash utilities'; exit 1; }

while [[ $# -ne 0 ]]; do
  case "$1" in
    --project)    PROJECT="$2" ; shift ;;
    --zone)       ZONE="$2" ; shift ;;
    *)            CLUSTERNAME="$1" ;;
  esac
  shift
done

[[ -n "${CLUSTERNAME:-}" ]] \
  || { echo 'Cluster name must be provided.'; exit 1; }

[[ -n "${ZONE:-}" ]] \
  || { echo 'Zone must be provided.'; exit 1; }

[[ -n "${PROJECT:-}" ]] \
  || { echo 'Project must be provided.'; exit 1; }

wipe_cluster \
  || error_exit 'Failed to shutdown Kubernetes resources on cluster.'

gcloud container clusters delete "${CLUSTERNAME}" --project="${PROJECT}" \
    --zone="${ZONE}" --quiet --async
