#!/usr/bin/env bash

# Copyright 2018 The Kubernetes Authors.
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

set -xeuo pipefail

REPO_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
DOCSITE=$REPO_ROOT/docsite
DEST=$DOCSITE/_site

clean() {
  # Start fresh so that removed files are picked up
  if [[ -d $DEST ]]; then
    rm -r $DEST
  fi
  mkdir -p $DEST
}

# Serve a live preview of the site
preview() {
  clean

  docker run -it --rm \
    -v $DOCSITE:/srv/jekyll \
    -v $REPO_ROOT/docs:/srv/docs \
    -v $DOCSITE/.bundler:/usr/local/bundle \
    -p 4000:4000 jekyll/jekyll jekyll serve
}

# Generate the static site's content
generate() {
  clean

  echo "Generating site..."
  docker run -it --rm \
    -v $DOCSITE:/srv/jekyll \
    -v $REPO_ROOT/docs:/srv/docs \
    -v $DOCSITE/.bundler:/usr/local/bundle \
    jekyll/jekyll /bin/bash -c "bundle install; jekyll build"
}

"$@"
