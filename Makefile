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

include ./hack/Makefile.mk

# Directories that the make will recurse into.
DIRS := \
  contrib/registry \
  contrib/broker/k8s \
  contrib/broker/user_provided \
  contrib/broker/server \
  pkg \
  pkg/apis/servicecatalog \
  pkg/apis/servicecatalog/v1alpha1 \
  pkg/controller/catalog

ALL := all build build-linux build-darwin clean docker push test lint coverage
SUB := $(addsuffix .sub, $(ALL))
.PHONY: $(ALL) $(SUB) format

# Recursive targets.
build: build.sub
build-linux: build-linux.sub
build-darwin: build-darwin.sub
docker: docker.sub
lint: lint.sub
push: push.sub
test: test.sub

clean: clean.sub
	rm -f .dockerInit
	rm -f .scBuildImage
	docker rmi -f scbuildimage > /dev/null 2>&1 || true

# Use this target when you want to build everything using docker containers.
# Good for cases when you don't have the tools installed (like glide or go).
docker-all: .scBuildImage
	docker run --rm -ti \
	  -v $(PWD):/go/src/github.com/kubernetes-incubator/service-catalog \
	  scbuildimage \
	  make .dockerInit all

# .dockerInit tells us if our vendor stuff if out of date or not.
# And if so we'll run init under our docker build.  For non-docker builds
# it is assumed you'll run "make init" manually.
.dockerInit: glide.yaml
	make init
	echo > .dockerInit

# .scBuildImage is used to know when the docker image ("scbuildimage") is out
# of date with the Dockerfile.
.scBuildImage: hack/Dockerfile
	docker build -t scbuildimage - < hack/Dockerfile
	echo > .scBuildImage

# Build the same target recursively in all directories.
$(SUB): %.sub:
	$(ECHO) for dir in $(DIRS); do $(MAKE) --no-print-directory -C "$${dir}" $* || exit $$? ; done

init:
	$(ECHO) glide install --strip-vendor

format:
	$(ECHO) gofmt -w -s $(addprefix ./,$(DIRS))

coverage:
	$(ECHO) $(ROOT)/hack/coverage.sh --html "$(COVERAGE)" $(addprefix ./,$(DIRS))

.PHONY: apiserver
apiserver:
	go install -v github.com/kubernetes-incubator/service-catalog/cmd/service-catalog
	go build -v -o apiserver cmd/service-catalog/server.go

verify:

# Runs all the presubmission verifications.
#
# Args:
#   BRANCH: Branch to be passed to verify-godeps.sh script.
#
# Example:
#   make verify
#   make verify BRANCH=branch_x
.PHONY: verify
verify:
	KUBE_VERIFY_GIT_BRANCH=$(BRANCH) hack/verify-all.sh -v

