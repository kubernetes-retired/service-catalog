all: build test verify

# Define some constants
#######################
ROOT          = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BINDIR       ?= bin
COVERAGE     ?= $(CURDIR)/coverage.html
SC_PKG        = github.com/kubernetes-incubator/service-catalog
TOP_SRC_DIRS  = cmd contrib pkg
SRC_DIRS      = $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*.go \
                  -exec dirname {} \\; | sort | uniq")
TEST_DIRS     = $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
                  -exec dirname {} \\; | sort | uniq")
SRC_PKGS      = $(addprefix $(SC_PKG)/,$(SRC_DIRS))
GO_VERSION    = 1.7.3
GO_BUILD      = go build
BASE_PATH     = $(ROOT:/src/github.com/kubernetes-incubator/service-catalog/=)
export GOPATH = $(BASE_PATH):$(ROOT)/vendor
DOCKER_CMD    = docker run --rm -ti -v $(PWD):/go/src/$(SC_PKG) \
                 -e GOOS=$$SC_GOOS -e GOARCH=$$SC_GOARCH \
                 scbuildimage

ifneq ($(origin DOCKER),undefined)
  # If DOCKER is defined then make it the full docker cmd line we want to use
  DOCKER=$(DOCKER_CMD)
endif

# This section builds the output binaries.
# Some will have dedicated targets to make it easier to type, for example
# "apiserver" instead of "bin/apiserver".
#########################################################################
build: init \
       $(BINDIR)/controller $(BINDIR)/registry $(BINDIR)/k8s-broker \
       $(BINDIR)/service-catalog $(BINDIR)/user-broker \
       $(BINDIR)/apiserver

controller: $(BINDIR)/controller
$(BINDIR)/controller: pkg/controller/catalog
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/$^

registry: $(BINDIR)/registry
$(BINDIR)/registry: contrib/registry
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/$^

$(BINDIR)/k8s-broker: contrib/broker/k8s
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/$^

$(BINDIR)/service-catalog: cmd/service-catalog
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/$^

$(BINDIR)/user-broker: contrib/broker/k8s
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/$^

apiserver: $(BINDIR)/apiserver
$(BINDIR)/apiserver: cmd/service-catalog
	$(DOCKER) $(GO_BUILD) -o $@ $(SC_PKG)/$^

# Some prereq stuff
###################
init: .scBuildImage .init

# .init is used to know when some Glide dependencies are out of date
.init: glide.yaml
	$(DOCKER) glide install --strip-vendor
	echo > $@

# .scBuildImage tells when the docker image ("scbuildimage") is out of date
.scBuildImage: hack/Dockerfile
	sed "s/GO_VERSION/$(GO_VERSION)/g" < hack/Dockerfile | \
	  docker build -t scbuildimage -
	echo > $@

# Util targets
##############

verify: init
	@echo Running gofmt:
	@$(DOCKER) gofmt -l -s $(TOP_SRC_DIRS) > .out 2>&1 || true
	@bash -c '[ "`cat .out`" == "" ] || \
	  (echo -e "\n*** Please 'gofmt' the following:" ; cat .out ; echo ; false)'
	@rm .out
	@echo Running golint and go vet:
	@for i in $(SRC_PKGS); do \
	  ($(DOCKER) golint --set_exit_status $$i/... && \
	   $(DOCKER) go vet $$i )|| exit $$? ; \
	done || false

format: init
	$(DOCKER) gofmt -w -s $(TOP_SRC_DIRS)

coverage: init
	$(DOCKER) hack/coverage.sh --html "$(COVERAGE)" $(addprefix ./,$(TEST_DIRS))

test:
	@echo Running tests:
	@for i in $(SRC_PKGS); do \
	  $(DOCKER) go test $$i || exit $$? ; \
	done

clean:
	rm -rf $(BINDIR)
	rm -f .init .scBuildImage
	rm -f $(COVERAGE)
	find $(TOP_SRC_DIRS) -name zz_generated* -exec rm {} \;
	docker rmi -f scbuildimage > /dev/null 2>&1 || true
