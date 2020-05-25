#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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

# This script checks version dependencies of modules. It checks whether all
# pinned versions of checked dependencies match their preferred version or not.
# Usage: `./contrib/hack/verify-modules.sh`.

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly TMP_DIR=$(mktemp -d)

source "${CURRENT_DIR}/ci/lib/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

# Explicitly opt into go modules, even though we're inside a GOPATH directory
export GO111MODULE=on
# Explicitly clear GOFLAGS, since GOFLAGS=-mod=vendor breaks dependency resolution while rebuilding vendor
export GOFLAGS=
# Detect problematic GOPROXY settings that prevent lookup of dependencies
if [[ "${GOPROXY:-}" == "off" ]]; then
  echo "Cannot run with \$GOPROXY=off"
  exit 1
fi

golang::verify_go_version
require-jq

outdated=$(go list -m -json all | jq -r "
  select(.Replace.Version != null) |
  select(.Version != .Replace.Version) |
  select(.Path) |
  \"\(.Path)
    pinned:    \(.Replace.Version)
    preferred: \(.Version)
    ./contrib/hack/pin-dependency.sh \(.Path) \(.Version)\"
")
if [[ -n "${outdated}" ]]; then
  echo "These modules are pinned to versions different than the minimal preferred version."
  echo "That means that without replace directives, a different version would be selected,"
  echo "which breaks consumers of our published modules."
  echo "1. Use ./contrib/hack/pin-dependency.sh to switch to the preferred version for each module"
  echo "2. Run ./contrib/hack/update-vendor.sh to rebuild the vendor directory"
  echo "3. Run ./contrib/hack/verify-modules.sh to verify no additional changes are required"
  echo ""
  echo "${outdated}"
fi

unused=$(comm -23 \
  <(go mod edit -json | jq -r '.Replace[] | select(.New.Version != null) | .Old.Path' | sort) \
  <(go list -m -json all | jq -r .Path | sort))
if [[ -n "${unused}" ]]; then
  echo ""
  echo "Use the given commands to remove pinned module versions that aren't actually used:"
  echo "${unused}" | xargs -L 1 echo 'GO111MODULE=on go mod edit -dropreplace'
fi

"${CURRENT_DIR}/update-vendor.sh"

modifedModules=$(git status --porcelain go.mod go.sum vendor)
if [ -n "$modifedModules" ]; then
    echo "go mod tidy modified go.mod and/or go.sum"
fi

if [[ -n "${unused}${modifedModules}${outdated}" ]]; then
  exit 1
fi

echo "All pinned versions of checked dependencies match their preferred version."
exit 0