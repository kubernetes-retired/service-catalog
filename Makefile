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

all: build test verify

# Define some constants
#######################
ROOT          = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BINDIR       ?= bin
COVERAGE     ?= $(CURDIR)/coverage.html
SC_PKG        = github.com/kubernetes-incubator/service-catalog
TOP_SRC_DIRS  = cmd contrib pkg util
SRC_DIRS      = $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*.go \
                  -exec dirname {} \\; | sort | uniq")
TEST_DIRS     = $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
                  -exec dirname {} \\; | sort | uniq")
GO_VERSION    = 1.7.3
GO_BUILD      = go build -i -v
BASE_PATH     = $(ROOT:/src/github.com/kubernetes-incubator/service-catalog/=)
export GOPATH = $(BASE_PATH):$(ROOT)/vendor
DOCKER_CMD    = docker run --rm -ti -v $(PWD):/go/src/$(SC_PKG) \
                 -e GOOS=$$SC_GOOS -e GOARCH=$$SC_GOARCH \
                 scbuildimage

ifneq ($(origin DOCKER),undefined)
  # If DOCKER is defined then make it the full docker cmd line we want to use
  DOCKER=$(DOCKER_CMD)
  # Setting scBuildImageTarget will force the Docker image to be built
  # in the .init rule
  scBuildImageTarget=.scBuildImage
endif

# This section builds the output binaries.
# Some will have dedicated targets to make it easier to type, for example
# "apiserver" instead of "bin/apiserver".
#########################################################################
build: .init .generate_files \
       $(BINDIR)/controller $(BINDIR)/registry $(BINDIR)/k8s-broker \
       $(BINDIR)/service-catalog $(BINDIR)/user-broker \
       $(BINDIR)/apiserver

controller: $(BINDIR)/controller
$(BINDIR)/controller: pkg/controller/catalog $(shell find pkg/controller/catalog -type f)
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/pkg/controller/catalog

registry: $(BINDIR)/registry
$(BINDIR)/registry: contrib/registry $(shell find contrib/registry -type f)
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/contrib/registry

$(BINDIR)/k8s-broker: contrib/broker/k8s $(shell find contrib/broker/k8s -type f)
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/contrib/broker/k8s

$(BINDIR)/service-catalog: cmd/service-catalog $(shell find cmd/service-catalog -type f)
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/cmd/service-catalog

$(BINDIR)/user-broker: contrib/broker/k8s $(shell find contrib/broker/k8s -type f)
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/contrib/broker/k8s

apiserver: $(BINDIR)/apiserver
$(BINDIR)/apiserver: cmd/service-catalog $(shell find pkg/apis/servicecatalog -type f)
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/cmd/service-catalog

# This section contains the code generation stuff
#################################################
.generate_exes: $(BINDIR)/defaulter-gen $(BINDIR)/deepcopy-gen
	touch $@

$(BINDIR)/defaulter-gen: cmd/libs/go2idl/defaulter-gen
	$(DOCKER) go build -o $@ $(SC_PKG)/$^

$(BINDIR)/deepcopy-gen: cmd/libs/go2idl/deepcopy-gen
	$(DOCKER) go build -o $@ $(SC_PKG)/$^

.generate_files: .generate_exes
	$(DOCKER) $(BINDIR)/defaulter-gen --v 1 --logtostderr \
	  -i $(SC_PKG)/pkg/apis/servicecatalog,$(SC_PKG)/pkg/apis/servicecatalog/v1alpha1 \
	  --extra-peer-dirs $(SC_PKG)/pkg/apis/servicecatalog,$(SC_PKG)/pkg/apis/servicecatalog/v1alpha1 \
	  -O zz_generated.defaults
	$(DOCKER) $(BINDIR)/deepcopy-gen --v 1 --logtostderr \
	  -i $(SC_PKG)/pkg/apis/servicecatalog,$(SC_PKG)/pkg/apis/servicecatalog/v1alpha1 \
	  --bounding-dirs github.com/kubernetes-incubator/service-catalog \
	  -O zz_generated.deepcopy
	  touch $@

# Some prereq stuff
###################
.init: $(scBuildImageTarget) glide.yaml
	$(DOCKER) glide install --strip-vendor
	touch $@

.scBuildImage: hack/Dockerfile
	sed "s/GO_VERSION/$(GO_VERSION)/g" < hack/Dockerfile | \
	  docker build -t scbuildimage -
	touch $@

# Util targets
##############
verify: .init .generate_files
	@echo Running gofmt:
	@$(DOCKER) gofmt -l -s $(TOP_SRC_DIRS) > .out 2>&1 || true
	@bash -c '[ "`cat .out`" == "" ] || \
	  (echo -e "\n*** Please 'gofmt' the following:" ; cat .out ; echo ; false)'
	@rm .out
	@echo Running golint and go vet:
	# Exclude the generated (zz) files for now
	@# The following command echo's the "for" loop to stdout so that it can
	@# be piped to the "sh" cmd running in the container. This allows the
	@# "for" to be executed in the container and not on the host. Which means
	@# we have just one container for everything and not one container per
	@# file.  The $(subst) removes the "-t" flag from the Docker cmd.
	@echo for i in \`find $(TOP_SRC_DIRS) -name \*.go \| grep -v zz\`\; do \
	  golint --set_exit_status \$$i \&\& \
	  go vet \$$i \; \
	done | $(subst -ti,-i,$(DOCKER)) sh
	@echo Running repo-infra verify scripts
	$(DOCKER) vendor/github.com/kubernetes/repo-infra/verify/verify-boilerplate.sh --rootdir=. | grep -v zz_generated > .out 2>&1 || true
	@bash -c '[ "`cat .out`" == "" ] || (cat .out ; false)'
	@rm .out

format: .init
	$(DOCKER) gofmt -w -s $(TOP_SRC_DIRS)

coverage: .init
	$(DOCKER) hack/coverage.sh --html "$(COVERAGE)" $(addprefix ./,$(TEST_DIRS))

test:
	@echo Running tests:
	@for i in $(addprefix $(SC_PKG)/,$(TEST_DIRS)); do \
	  $(DOCKER) go test $$i || exit $$? ; \
	done

build-darwin:
	SC_GOOS=darwin SC_GOARCH=amd64 BINDIR=bin/darwin_amd64 DOCKER=1 \
	  $(MAKE) build

build-linux:
	SC_GOOS=linux SC_GOARCH=amd64 BINDIR=bin/darwin_amd64 DOCKER=1 \
	  $(MAKE) build

clean:
	rm -rf $(BINDIR)
	rm -f .init .scBuildImage .generate_files .generate_exes
	rm -f $(COVERAGE)
	find $(TOP_SRC_DIRS) -name zz_generated* -exec rm {} \;
	docker rmi -f scbuildimage > /dev/null 2>&1 || true
