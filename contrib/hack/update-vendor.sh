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

# Copied and adjsuted version of https://github.com/kubernetes/kubernetes/blob/v1.18.2/hack/update-vendor.sh
# all stuff regarding stage repos were removed

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
# Ensure sort order doesn't depend on locale
export LANG=C
export LC_ALL=C
# Detect problematic GOPROXY settings that prevent lookup of dependencies
if [[ "${GOPROXY:-}" == "off" ]]; then
  echo "Cannot run ./contrib/hack/update-vendor.sh with \$GOPROXY=off"
  exit 1
fi

golang::verify_go_version
require-jq

function group_replace_directives() {
  local local_tmp_dir
  local_tmp_dir=$(mktemp -d "${TMP_DIR}/group_replace.XXXX")
  local go_mod_replace="${local_tmp_dir}/go.mod.replace.tmp"
  local go_mod_noreplace="${local_tmp_dir}/go.mod.noreplace.tmp"
  # separate replace and non-replace directives
  awk "
     # print lines between 'replace (' ... ')' lines
     /^replace [(]/      { inreplace=1; next                   }
     inreplace && /^[)]/ { inreplace=0; next                   }
     inreplace           { print > \"${go_mod_replace}\"; next }

     # print ungrouped replace directives with the replace directive trimmed
     /^replace [^(]/ { sub(/^replace /,\"\"); print > \"${go_mod_replace}\"; next }

     # otherwise print to the noreplace file
     { print > \"${go_mod_noreplace}\" }
  " < go.mod
  {
    cat "${go_mod_noreplace}";
    echo "replace (";
    cat "${go_mod_replace}";
    echo ")";
  } > go.mod

  go mod edit -fmt
}

function add_generated_comments() {
  local local_tmp_dir
  local_tmp_dir=$(mktemp -d "${TMP_DIR}/add_generated_comments.XXXX")
  local go_mod_nocomments="${local_tmp_dir}/go.mod.nocomments.tmp"

  # drop comments before the module directive
  awk "
     BEGIN           { dropcomments=1 }
     /^module /      { dropcomments=0 }
     dropcomments && /^\/\// { next }
     { print }
  " < go.mod > "${go_mod_nocomments}"

  # Add the specified comments
  local comments="${1}"
  {
    echo "${comments}"
    echo ""
    cat "${go_mod_nocomments}"
   } > go.mod

  # Format
  go mod edit -fmt
}

shout "Phase 1: tidying and grouping replace directives"
# resolves/expands references in the root go.mod (if needed)
go mod tidy
# group replace directives
group_replace_directives

shout "Phase 2: adding generated comments"
add_generated_comments "
// This is a generated file. Do not edit directly.
// Run ./contrib/hack/pin-dependency.sh to change pinned dependency versions.
// Run ./contrib/hack/update-vendor.sh to update go.mod files and the vendor directory.
"

shout "Phase 3: rebuild vendor directory"
go mod vendor

# sort recorded packages for a given vendored dependency in modules.txt.
# `go mod vendor` outputs in imported order, which means slight go changes (or different platforms) can result in a differently ordered modules.txt.
# scan                 | prefix comment lines with the module name       | sort field 1  | strip leading text on comment lines
awk '{if($1=="#") print $2 " " $0; else print}' < vendor/modules.txt | sort -k1,1 -s | sed 's/.*#/#/' > "${TMP_DIR}/modules.txt.tmp"
mv "${TMP_DIR}/modules.txt.tmp" vendor/modules.txt
