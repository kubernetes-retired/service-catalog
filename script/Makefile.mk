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

# Shared Makefile definitions.
#

.PHONY: all build build-darwin build-linux docker push test

ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
# Strip /script/ suffix to get root of the Git client.
ROOT := $(ROOT:/script/=)
# Strip the /src/github.com/kubernetes-incubator/service-catalog suffix to get the GOPATH.
export GOPATH := $(ROOT:/src/github.com/kubernetes-incubator/service-catalog=)

export VERSION ?= $(shell git describe --always --abbrev=40 --dirty)
export HOST_OS ?= $(shell uname -s | tr A-Z a-z)

GO             := go
GOLINT         := golint
GO_VERSION     := 1.7.3
BINDIR         := $(GOPATH)/bin
PKG_ROOT       := github.com/kubernetes-incubator/service-catalog
ARCH           := amd64
COVERAGE       ?= $(CURDIR)/coverage.html

ifeq "$(V)" "1"
  $(info Makefile.mk included from $(CURDIR))
else
  ECHO := @
endif

#
# If REGISTRY is defined, prepare to push images.
#
ifeq ($(origin REGISTRY),undefined)
define docker_push
  $(info Not pushing $(PKG). Please set REGISTRY variable.)
  @false
endef
else
export REGISTRY
define docker_push
  $(ECHO) echo 'Pushing $(PKG)'
  $(ECHO) echo '  tagging '$(1):$(VERSION)' as $(REGISTRY)/$(1):$(VERSION)'
  $(ECHO) docker tag '$(1):$(VERSION)' '$(REGISTRY)/$(1):$(VERSION)'
  $(ECHO) docker push '$(REGISTRY)/$(1):$(VERSION)'
endef
endif

define docker_build
  $(ECHO) echo "Building Docker"
  $(ECHO) cp Dockerfile $(GOPATH)/bin/linux_$(ARCH)/dockerfile.tmp
  $(ECHO) docker build -t "$(1):$(VERSION)" \
        -f '$(GOPATH)/bin/linux_$(ARCH)/dockerfile.tmp' \
        '$(GOPATH)/bin/linux_$(ARCH)'
  $(ECHO) rm -rf '$(GOPATH)/bin/linux_$(ARCH)/dockerfile.tmp'
endef

define delete_binaries
  $(ECHO) rm -f "$(BINDIR)/$(BIN)"
  $(ECHO) rm -f "$(BINDIR)/linux_$(ARCH)/$(BIN)"
  $(ECHO) rm -f "$(BINDIR)/darwin_$(ARCH)/$(BIN)"
endef

ifeq ($(origin NO_DOCKER_COMPILE),undefined)

  # Use of docker-hosted compilation is allowed. Build
  # the platform-specific binaries inside Docker.
  define platform_compile
    $(info Building $(BIN) for $(1))
    $(ECHO) mkdir -p "$(GOPATH)/bin/$(1)_$(ARCH)"
    $(ECHO) docker run \
        --rm \
        --volume "$(GOPATH)":/go \
        --workdir "/go/src/github.com/kubernetes-incubator/service-catalog" \
        --env GOOS=$(1) \
        --env GOARCH=$(ARCH) \
        golang:$(GO_VERSION) \
        go build -o /go/bin/$(1)_$(ARCH)/$(BIN) -ldflags "-X github.com/kubernetes-incubator/service-catalog/pkg.VERSION=$(VERSION)" $(PKG)
  endef

else

  # Define rules which fail to compile. Then we dynamically
  # define one which works only when compiling for host platform.
  define platform_compile_linux
    $(error "Cannot compile for Linux on $(HOST_OS)")
  endef

  define platform_compile_darwin
    $(error "Cannot compile for Darwin on $(HOST_OS)")
  endef

  # Override the platform compile rule for HOST_OS, allowing
  # for native compilation.
  define platform_compile_$(HOST_OS)
    $(info Building $(BIN) for $(HOST_OS))
    $(ECHO) $(GO) build -o "$(BINDIR)/$(HOST_OS)_$(ARCH)/$(BIN)" "$(PKG)"
  endef

  # platform_compile delegates to one of the macros defined above.
  # If the platform matches host OS, Go compiler is used. Otherwise
  # we fail to compile because Go cross-compilation is disabled.
  define platform_compile
    $(call platform_compile_$(1))
  endef

endif

# The first target in the makefile is the default.
all: build test lint
