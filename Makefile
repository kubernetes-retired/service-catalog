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
ROOT           = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BINDIR        ?= bin
COVERAGE      ?= $(CURDIR)/coverage.html
SC_PKG         = github.com/kubernetes-incubator/service-catalog
TOP_SRC_DIRS   = cmd contrib pkg util
SRC_DIRS       = $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*.go \
                   -exec dirname {} \\; | sort | uniq")
TEST_DIRS      = $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
                   -exec dirname {} \\; | sort | uniq")
VERSION       ?= $(shell git describe --tags --always --abbrev=7 --dirty)
ifeq ($(shell uname -s),Darwin)
STAT           = stat -f '%c %N'
else
STAT           = stat -c '%Y %n'
endif
NEWEST_GO_FILE = $(shell find $(SRC_DIRS) -name \*.go -exec $(STAT) {} \; \
                   | sort -r | head -n 1 | sed "s/.* //")
TYPES_FILES    = $(shell find pkg/apis -name types.go)
GO_VERSION     = 1.7.3
GO_BUILD       = env GOOS=linux GOARCH=amd64 go build -i -v \
                   -ldflags "-X $(SC_PKG)/pkg.VERSION=$(VERSION)"
BASE_PATH      = $(ROOT:/src/github.com/kubernetes-incubator/service-catalog/=)
export GOPATH  = $(BASE_PATH):$(ROOT)/vendor

ifneq ($(origin DOCKER),undefined)
  # If DOCKER is defined then define the full docker cmd line we want to use
  DOCKER_FLAG  = DOCKER=1
  DOCKER_CMD   = docker run --rm -ti -v $(PWD):/go/src/$(SC_PKG) scbuildimage
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
$(BINDIR)/controller: .init pkg/controller/catalog \
	  $(shell find pkg/controller/catalog -type f)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/pkg/controller/catalog

registry: $(BINDIR)/registry
$(BINDIR)/registry: .init contrib/registry \
	  $(shell find contrib/registry -type f)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/contrib/registry

$(BINDIR)/k8s-broker: .init contrib/broker/k8s \
	  $(shell find contrib/broker/k8s -type f)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/contrib/broker/k8s

$(BINDIR)/service-catalog: .init cmd/service-catalog \
	  $(shell find cmd/service-catalog -type f)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/cmd/service-catalog

$(BINDIR)/user-broker: .init contrib/broker/k8s \
	  $(shell find contrib/broker/k8s -type f)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/contrib/broker/k8s

# We'll rebuild apiserver if any go file has changed (ie. NEWEST_GO_FILE)
apiserver: $(BINDIR)/apiserver
$(BINDIR)/apiserver: .init .generate_files cmd/service-catalog $(NEWEST_GO_FILE)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/cmd/service-catalog

# This section contains the code generation stuff
#################################################
.generate_exes: $(BINDIR)/defaulter-gen $(BINDIR)/deepcopy-gen
	touch $@

$(BINDIR)/defaulter-gen: .init cmd/libs/go2idl/defaulter-gen
	$(DOCKER_CMD) go build -o $@ $(SC_PKG)/cmd/libs/go2idl/defaulter-gen

$(BINDIR)/deepcopy-gen: .init cmd/libs/go2idl/deepcopy-gen
	$(DOCKER_CMD) go build -o $@ $(SC_PKG)/cmd/libs/go2idl/deepcopy-gen

# Regenerate all files if the gen exes changed or any "types.go" files changed
.generate_files: .init .generate_exes $(TYPES_FILES)
	$(DOCKER_CMD) $(BINDIR)/defaulter-gen --v 1 --logtostderr \
	  -i $(SC_PKG)/pkg/apis/servicecatalog,$(SC_PKG)/pkg/apis/servicecatalog/v1alpha1 \
	  --extra-peer-dirs $(SC_PKG)/pkg/apis/servicecatalog,$(SC_PKG)/pkg/apis/servicecatalog/v1alpha1 \
	  -O zz_generated.defaults
	$(DOCKER_CMD) $(BINDIR)/deepcopy-gen --v 1 --logtostderr \
	  -i $(SC_PKG)/pkg/apis/servicecatalog,$(SC_PKG)/pkg/apis/servicecatalog/v1alpha1 \
	  --bounding-dirs github.com/kubernetes-incubator/service-catalog \
	  -O zz_generated.deepcopy
	  touch $@

# Some prereq stuff
###################
.init: $(scBuildImageTarget) glide.yaml
	$(DOCKER_CMD) glide install --strip-vendor
	touch $@

.scBuildImage: hack/Dockerfile
	sed "s/GO_VERSION/$(GO_VERSION)/g" < hack/Dockerfile | \
	  docker build -t scbuildimage \
	    --build-arg UID=$(shell id -u) \
	    --build-arg GID=$(shell id -g) \
	    --build-arg USER=$(USER) \
	    -
	touch $@

# Util targets
##############
verify: .init .generate_files
	@echo Running gofmt:
	@$(DOCKER_CMD) gofmt -l -s $(TOP_SRC_DIRS) > .out 2>&1 || true
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
	  golint --set_exit_status \$$i \; \
	  go vet \$$i \; \
	done | $(subst -ti,-i,$(DOCKER_CMD)) sh -e
	@echo Running repo-infra verify scripts
	$(DOCKER_CMD) vendor/github.com/kubernetes/repo-infra/verify/verify-boilerplate.sh --rootdir=. | grep -v zz_generated > .out 2>&1 || true
	@bash -c '[ "`cat .out`" == "" ] || (cat .out ; false)'
	@rm .out

format: .init
	$(DOCKER_CMD) gofmt -w -s $(TOP_SRC_DIRS)

coverage: .init
	$(DOCKER_CMD) hack/coverage.sh --html "$(COVERAGE)" \
	  $(addprefix ./,$(TEST_DIRS))

test: .init
	@echo Running tests:
	@for i in $(addprefix $(SC_PKG)/,$(TEST_DIRS)); do \
	  $(DOCKER_CMD) go test $$i || exit $$? ; \
	done

clean:
	rm -rf $(BINDIR)
	rm -f .init .scBuildImage .generate_files .generate_exes
	rm -f $(COVERAGE)
	find $(TOP_SRC_DIRS) -name zz_generated* -exec rm {} \;
	docker rmi -f scbuildimage > /dev/null 2>&1 || true

# Building Docker Images for our executables
############################################
images: registry-image k8s-broker-image user-broker-image controller-image

registry-image: contrib/registry/Dockerfile $(BINDIR)/registry
	cp contrib/registry/Dockerfile $(BINDIR)
	cp contrib/registry/data/charts/*.json $(BINDIR)
	docker build -t registry:$(VERSION) $(BINDIR)
	rm -f $(BINDIR)/Dockerfile
	rm -f $(BINDIR)/*.json

k8s-broker-image: contrib/broker/k8s/Dockerfile $(BINDIR)/k8s-broker
	cp contrib/broker/k8s/Dockerfile $(BINDIR)
	docker build -t k8s-broker:$(VERSION) $(BINDIR)
	rm -f $(BINDIR)/Dockerfile

user-broker-image: contrib/broker/user_provided/Dockerfile $(BINDIR)/user-broker
	cp contrib/broker/user_provided/Dockerfile $(BINDIR)
	docker build -t user-broker:$(VERSION) $(BINDIR)
	rm -f $(BINDIR)/Dockerfile

controller-image: pkg/controller/catalog/Dockerfile $(BINDIR)/controller
	cp pkg/controller/catalog/Dockerfile $(BINDIR)
	docker build -t controller:$(VERSION) $(BINDIR)
	rm -f $(BINDIR)/Dockerfile

# Push our Docker Images to a registry
######################################
push: registry-push k8s-broker-push user-broker-push controller-push

registry-push: registry-image
	[ ! -z "$(REGISTRY)" ] || (echo Set your REGISTRY env var first ; exit 1)
	docker tag registry:$(VERSION) $(REGISTRY)/registry:$(VERSION)
	docker push $(REGISTRY)/registry:$(VERSION)

k8s-broker-push: k8s-broker-image
	[ ! -z "$(REGISTRY)" ] || (echo Set your REGISTRY env var first ; exit 1)
	docker tag k8s-broker:$(VERSION) $(REGISTRY)/k8s-broker:$(VERSION)
	docker push $(REGISTRY)/k8s-broker:$(VERSION)

user-broker-push: user-broker-image
	[ ! -z "$(REGISTRY)" ] || (echo Set your REGISTRY env var first ; exit 1)
	docker tag user-broker:$(VERSION) $(REGISTRY)/user-broker:$(VERSION)
	docker push $(REGISTRY)/user-broker:$(VERSION)

controller-push: controller-image
	[ ! -z "$(REGISTRY)" ] || (echo Set your REGISTRY env var first ; exit 1)
	docker tag controller:$(VERSION) $(REGISTRY)/controller:$(VERSION)
	docker push $(REGISTRY)/controller:$(VERSION)
