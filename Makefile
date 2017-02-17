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

# Some env vars that devs might find useful:
#  GOFLAGS      : extra "go build" flags to use - e.g. -v   (for verbose)
#  NO_DOCKER=1  : execute each step natively, not in a Docker container
#  TEST_DIRS=   : only run the unit tests from the specified dirs
#  UNIT_TESTS=  : only run the unit tests matching the specified regexp

# Define some constants
#######################
ROOT           = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BINDIR        ?= bin
BUILD_DIR     ?= build
COVERAGE      ?= $(CURDIR)/coverage.html
SC_PKG         = github.com/kubernetes-incubator/service-catalog
TOP_SRC_DIRS   = cmd contrib pkg
SRC_DIRS       = $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*.go \
                   -exec dirname {} \\; | sort | uniq")
TEST_DIRS     ?= $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
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

PLATFORM?=linux
ARCH?=amd64

GO_BUILD       = env GOOS=$(PLATFORM) GOARCH=$(ARCH) go build -i $(GOFLAGS) \
                   -ldflags "-X $(SC_PKG)/pkg.VERSION=$(VERSION)"
BASE_PATH      = $(ROOT:/src/github.com/kubernetes-incubator/service-catalog/=)
export GOPATH  = $(BASE_PATH):$(ROOT)/vendor

# precheck to avoid kubernetes-incubator/service-catalog#361
$(if $(realpath vendor/k8s.io/kubernetes/vendor), \
	$(error the vendor directory exists in the kubernetes \
		vendored source and must be flattened. \
		run 'glide i -v'))

ifdef UNIT_TESTS
	UNIT_TEST_FLAGS=-run $(UNIT_TESTS) -v
endif

ifdef NO_DOCKER
	DOCKER_CMD =
	scBuildImageTarget =
else
	# Mount .pkg as pkg so that we save our cached "go build" output files
	DOCKER_CMD = docker run --rm -ti -v $(PWD):/go/src/$(SC_PKG) \
	  -v $(PWD)/.pkg:/go/pkg scbuildimage
	scBuildImageTarget = .scBuildImage
endif

# This section builds the output binaries.
# Some will have dedicated targets to make it easier to type, for example
# "apiserver" instead of "bin/apiserver".
#########################################################################
build: .init .generate_files \
       $(BINDIR)/controller-manager $(BINDIR)/controller $(BINDIR)/apiserver \
       $(BINDIR)/registry $(BINDIR)/k8s-broker $(BINDIR)/user-broker

controller: $(BINDIR)/controller
$(BINDIR)/controller: .init cmd/controller \
	  $(shell find cmd/controller -type f)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/cmd/controller

registry: $(BINDIR)/registry
$(BINDIR)/registry: .init contrib/cmd/registry \
	  $(shell find contrib/cmd/registry -type f)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/contrib/cmd/registry

k8s-broker: $(BINDIR)/k8s-broker
$(BINDIR)/k8s-broker: .init contrib/cmd/k8s-broker \
	  $(shell find contrib/cmd/k8s-broker -type f)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/contrib/cmd/k8s-broker

user-broker: $(BINDIR)/user-broker
$(BINDIR)/user-broker: .init contrib/cmd/user-broker \
	  $(shell find contrib/cmd/user-broker -type f)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/contrib/cmd/user-broker

# We'll rebuild apiserver if any go file has changed (ie. NEWEST_GO_FILE)
apiserver: $(BINDIR)/apiserver
$(BINDIR)/apiserver: .init .generate_files cmd/apiserver $(NEWEST_GO_FILE)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/cmd/apiserver

controller-manager: $(BINDIR)/controller-manager
$(BINDIR)/controller-manager: .init .generate_files cmd/controller-manager $(NEWEST_GO_FILE)
	$(DOCKER_CMD) $(GO_BUILD) -o $@ $(SC_PKG)/cmd/controller-manager

# This section contains the code generation stuff
#################################################
.generate_exes: $(BINDIR)/defaulter-gen \
                $(BINDIR)/deepcopy-gen \
                $(BINDIR)/conversion-gen \
                $(BINDIR)/client-gen \
                $(BINDIR)/lister-gen \
                $(BINDIR)/informer-gen \
                $(BINDIR)/openapi-gen
	touch $@

$(BINDIR)/defaulter-gen: .init
	$(DOCKER_CMD) go build -o $@ $(SC_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen

$(BINDIR)/deepcopy-gen: .init
	$(DOCKER_CMD) go build -o $@ $(SC_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen

$(BINDIR)/conversion-gen: .init
	$(DOCKER_CMD) go build -o $@ $(SC_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen

$(BINDIR)/client-gen: .init
	$(DOCKER_CMD) go build -o $@ $(SC_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/client-gen

$(BINDIR)/lister-gen: .init
	$(DOCKER_CMD) go build -o $@ $(SC_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/lister-gen

$(BINDIR)/informer-gen: .init
	$(DOCKER_CMD) go build -o $@ $(SC_PKG)/vendor/k8s.io/kubernetes/cmd/libs/go2idl/informer-gen

$(BINDIR)/openapi-gen: vendor/k8s.io/kubernetes/cmd/libs/go2idl/openapi-gen
	$(DOCKER_CMD) go build -o $@ $(SC_PKG)/$^

# Regenerate all files if the gen exes changed or any "types.go" files changed
.generate_files: .init .generate_exes $(TYPES_FILES)
	# Generate defaults
	$(DOCKER_CMD) $(BINDIR)/defaulter-gen \
		--v 1 --logtostderr \
		--go-header-file "vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(SC_PKG)/pkg/apis/servicecatalog" \
		--input-dirs "$(SC_PKG)/pkg/apis/servicecatalog/v1alpha1" \
	  	--extra-peer-dirs "$(SC_PKG)/pkg/apis/servicecatalog" \
		--extra-peer-dirs "$(SC_PKG)/pkg/apis/servicecatalog/v1alpha1" \
		--output-file-base "zz_generated.defaults"
	# Generate deep copies
	$(DOCKER_CMD) $(BINDIR)/deepcopy-gen \
		--v 1 --logtostderr \
		--go-header-file "vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(SC_PKG)/pkg/apis/servicecatalog" \
		--input-dirs "$(SC_PKG)/pkg/apis/servicecatalog/v1alpha1" \
		--bounding-dirs "github.com/kubernetes-incubator/service-catalog" \
		--output-file-base zz_generated.deepcopy
	# Generate conversions
	$(DOCKER_CMD) $(BINDIR)/conversion-gen \
		--v 1 --logtostderr \
		--go-header-file "vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(SC_PKG)/pkg/apis/servicecatalog" \
		--input-dirs "$(SC_PKG)/pkg/apis/servicecatalog/v1alpha1" \
		--output-file-base zz_generated.conversion
	# the previous three directories will be changed from kubernetes to apimachinery in the future
	# Generate the internal clientset (pkg/client/clientset_generated/internalclientset)
	$(DOCKER_CMD) $(BINDIR)/client-gen \
		--input-base "github.com/kubernetes-incubator/service-catalog/pkg/apis/" \
		--input servicecatalog/ \
		--clientset-path "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/" \
		--clientset-name internalclientset \
		--go-header-file "vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt"
	# Generate the versioned clientset (pkg/client/clientset_generated/clientset)
	$(DOCKER_CMD) $(BINDIR)/client-gen --input-base github.com/kubernetes-incubator/service-catalog/pkg/apis/ \
		--input servicecatalog/v1alpha1 \
		--clientset-path github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/ \
		--clientset-name clientset \
		--go-header-file vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt
	# generate lister
	$(DOCKER_CMD) $(BINDIR)/lister-gen \
		--input-dirs="github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog" \
		--input-dirs="github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1" \
		--output-package "github.com/kubernetes-incubator/service-catalog/pkg/client/listers" \
		--go-header-file vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt
	# generate informer
	$(DOCKER_CMD) $(BINDIR)/informer-gen \
		--go-header-file "vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog" \
		--input-dirs "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1" \
		--internal-clientset-package "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/internalclientset" \
		--versioned-clientset-package "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset" \
		--listers-package "github.com/kubernetes-incubator/service-catalog/pkg/client/listers" \
		--output-package "github.com/kubernetes-incubator/service-catalog/pkg/client/informers"
	# generate openapi
	$(DOCKER_CMD) $(BINDIR)/openapi-gen \
		--v 1 --logtostderr \
		--go-header-file "vendor/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1,k8s.io/kubernetes/pkg/api/v1,k8s.io/kubernetes/pkg/apis/meta/v1" \
		--output-package "github.com/kubernetes-incubator/service-catalog/pkg/openapi"
	# generate codec
	$(DOCKER_CMD) $(BUILD_DIR)/update-codecgen.sh
	touch $@

# Some prereq stuff
###################
.init: $(scBuildImageTarget) glide.yaml
	$(DOCKER_CMD) glide install --strip-vendor
	touch $@

.scBuildImage: build/build-image/Dockerfile
	sed "s/GO_VERSION/$(GO_VERSION)/g" < build/build-image/Dockerfile | \
	  docker build -t scbuildimage -
	touch $@

# Util targets
##############
verify: .init .generate_files
	@echo Running gofmt:
	@$(DOCKER_CMD) gofmt -l -s $(TOP_SRC_DIRS) > .out 2>&1 || true
	@bash -c '[ "`cat .out`" == "" ] || \
	  (echo -e "\n*** Please 'gofmt' the following:" ; cat .out ; echo ; false)'
	@rm .out
	@#
	@echo Running golint and go vet:
	@# Exclude the generated (zz) files for now, as well as defaults.go (it
	@# observes conventions from upstream that will not pass lint checks).
	@$(DOCKER_CMD) sh -c \
	  'for i in $$(find $(TOP_SRC_DIRS) -name *.go \
	    | grep -v generated \
	    | grep -v v1alpha1/defaults.go); \
	  do \
	    golint --set_exit_status $$i; \
	    go vet $$i; \
	  done'
	@#
	@echo Running repo-infra verify scripts
	@$(DOCKER_CMD) vendor/github.com/kubernetes/repo-infra/verify/verify-boilerplate.sh --rootdir=. | grep -v generated > .out 2>&1 || true
	@bash -c '[ "`cat .out`" == "" ] || (cat .out ; false)'
	@rm .out
	@#
	@echo Running href checker:
	@$(DOCKER_CMD) build/verify-links.sh
	@echo Running errexit checker:
	@$(DOCKER_CMD) build/verify-errexit.sh

format: .init
	$(DOCKER_CMD) gofmt -w -s $(TOP_SRC_DIRS)

coverage: .init
	$(DOCKER_CMD) contrib/hack/coverage.sh --html "$(COVERAGE)" \
	  $(addprefix ./,$(TEST_DIRS))

test: .init test-unit test-integration

test-unit: .init
	@echo Running tests:
	$(DOCKER_CMD) go test $(UNIT_TEST_FLAGS) \
	  $(addprefix $(SC_PKG)/,$(TEST_DIRS))

test-integration: .init
	contrib/hack/setup-kubectl.sh
	contrib/hack/test-apiserver.sh

clean:
	rm -rf $(BINDIR)
	rm -f .init .scBuildImage .generate_files .generate_exes
	rm -f $(COVERAGE)
	find $(TOP_SRC_DIRS) -name zz_generated* -exec rm {} \;
	docker rmi -f scbuildimage > /dev/null 2>&1 || true

# Building Docker Images for our executables
############################################
images: registry-image k8s-broker-image user-broker-image controller-image \
    apiserver-image

registry-image: contrib/build/registry/Dockerfile $(BINDIR)/registry
	mkdir -p contrib/build/registry/tmp
	cp $(BINDIR)/registry contrib/build/registry/tmp
	cp contrib/pkg/registry/data/charts/*.json contrib/build/registry/tmp
	docker build -t registry:$(VERSION) contrib/build/registry
	rm -rf contrib/build/registry/tmp

k8s-broker-image: contrib/build/k8s-broker/Dockerfile $(BINDIR)/k8s-broker
	mkdir -p contrib/build/k8s-broker/tmp
	cp $(BINDIR)/k8s-broker contrib/build/k8s-broker/tmp
	docker build -t k8s-broker:$(VERSION) contrib/build/k8s-broker
	rm -rf contrib/build/k8s-broker/tmp

user-broker-image: contrib/build/user-broker/Dockerfile $(BINDIR)/user-broker
	mkdir -p contrib/build/user-broker/tmp
	cp $(BINDIR)/user-broker contrib/build/user-broker/tmp
	docker build -t user-broker:$(VERSION) contrib/build/user-broker
	rm -rf contrib/build/user-broker/tmp

controller-image: build/controller/Dockerfile $(BINDIR)/controller
	mkdir -p build/controller/tmp
	cp $(BINDIR)/controller build/controller/tmp
	docker build -t controller:$(VERSION) build/controller
	rm -rf build/controller/tmp

apiserver-image: build/apiserver/Dockerfile $(BINDIR)/apiserver
	mkdir -p build/apiserver/tmp
	cp $(BINDIR)/apiserver build/apiserver/tmp
	docker build -t apiserver:$(VERSION) build/apiserver
	docker tag apiserver:$(VERSION) apiserver:latest
	rm -rf build/apiserver/tmp

controller-manager-image: build/controller-manager/Dockerfile $(BINDIR)/controller-manager
	mkdir -p build/controller-manager/tmp
	cp $(BINDIR)/controller-manager build/controller-manager/tmp
	docker build -t controller-manager:$(VERSION) build/controller-manager
	docker tag controller-manager:$(VERSION) controller-manager:latest
	rm -rf build/controller-manager/tmp

# Push our Docker Images to a registry
######################################
push: registry-push k8s-broker-push user-broker-push controller-push \
    apiserver-push

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

apiserver-push: apiserver-image
	[ ! -z "$(REGISTRY)" ] || (echo Set your REGISTRY env var first ; exit 1)
	docker tag apiserver:$(VERSION) $(REGISTRY)/apiserver:$(VERSION)
	docker push $(REGISTRY)/apiserver:$(VERSION)
	docker tag apiserver:$(VERSION) $(REGISTRY)/apiserver:latest
	docker push $(REGISTRY)/apiserver:latest
