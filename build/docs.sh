#!/usr/bin/env bash
set -xeuo pipefail

REPO_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
DOCSITE=$REPO_ROOT/docsite
DEST=$DOCSITE/_site

# Run the jekyll binary in a container, allowing for live-edits of the site's content
jekyll () {
  docker run -it --rm \
    -v $DOCSITE:/srv/jekyll \
    -v $REPO_ROOT/docs:/srv/docs \
    -v $DOCSITE/.bundler:/usr/local/bundle \
    -p 4000:4000 jekyll/jekyll jekyll $@
}

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

  jekyll serve
}

# Generate the static site's content
generate() {
  clean

  echo "Generating site..."
  jekyll build
}

"$@"
