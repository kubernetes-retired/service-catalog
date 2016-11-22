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

.PHONY: lint

$(BINDIR)/$(GOLINT):
	$(GO) get -u github.com/golang/lint/golint

ifneq (,$(BIN))

build:
	$(ECHO) echo "Building $(PKG)"
	$(ECHO) $(GO) build -o "$(BINDIR)/$(BIN)" -ldflags "-X github.com/kubernetes-incubator/service-catalog/pkg.VERSION=$(VERSION)" "$(PKG)"

build-linux build-darwin: build-%:
	$(call platform_compile,$*)

clean::
	$(call delete_binaries,$(BIN))

else

build:
	@echo > /dev/null

build-linux build-darwin: build-%:
	@echo > /dev/null

clean::
	@echo > /dev/null

endif

ifneq (,$(DOCKER))

docker:: build-linux Dockerfile
	$(call docker_build,$(BIN))

push:
	$(ECHO) $(call docker_push,$(BIN))

else

push:
	@echo > /dev/null

docker:
	@echo > /dev/null

endif

test:
	$(ECHO) echo "Testing $(PKG)"
	$(ECHO) $(GO) test "$(PKG)/..."

lint: $(BINDIR)/$(GOLINT)
	$(ECHO) $(BINDIR)/$(GOLINT) --set_exit_status "$(PKG)/..."
	$(ECHO) $(GO) vet "$(PKG)/..."

coverage:
	$(ECHO) $(ROOT)/hack/coverage.sh --html "$(CURDIR)/$(BIN)-coverage.html" "$(PKG)"

%:
	$(ECHO) echo "Not building $* in $(CURDIR)"

