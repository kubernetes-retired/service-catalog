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

# Copied and adjusted version of https://github.com/kubernetes/kubernetes/blob/v1.18.2/hack/pin-dependency.sh

# This script switches to the preferred version for specified module.
# Usage: `./contrib/hack/pin-dependency.sh $MODULE $SHA-OR-TAG`.
# Example: `./contrib/hack/pin-dependency.sh github.com/docker/docker 501cb131a7b7`.

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

dep="${1:-}"
sha="${2:-}"
if [[ -z "${dep}" || -z "${sha}" ]]; then
  echo "Usage:"
  echo "  ./contrib/hack/pin-dependency.sh \$MODULE \$SHA-OR-TAG"
  echo ""
  echo "Example:"
  echo "  ./contrib/hack/pin-dependency.sh github.com/docker/docker 501cb131a7b7"
  echo ""
  exit 1
fi

cleanup() {
  rm -rf "${TMP_DIR}"
}
trap "cleanup" EXIT SIGINT
cleanup

# Add the require directive
echo "Running: go get ${dep}@${sha}"
go get -d "${dep}@${sha}"

# Find the resolved version
rev=$(go mod edit -json | jq -r ".Require[] | select(.Path == \"${dep}\") | .Version")

# No entry in go.mod, we must be using the natural version indirectly
if [[ -z "${rev}" ]]; then
  # backup the go.mod file, since go list modifies it
  cp go.mod "${TMP_DIR}/go.mod.bak"
  # find the revision
  rev=$(go list -m -json "${dep}" | jq -r .Version)
  # restore the go.mod file
  mv "${TMP_DIR}/go.mod.bak" go.mod
fi

# No entry found
if [[ -z "${rev}" ]]; then
  echo "Could not resolve ${sha}"
  exit 1
fi

echo "Resolved to ${dep}@${rev}"

# Add the replace directive
echo "Running: go mod edit -replace ${dep}=${dep}@${rev}"
go mod edit -replace "${dep}=${dep}@${rev}"

echo "Run ./contrib/hack/update-vendor.sh to rebuild the vendor directory"
