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
  controller \
  pkg \
  pkg/apis/servicecatalog

ALL := all build build-linux build-darwin clean docker push test lint coverage
SUB := $(addsuffix .sub, $(ALL))
.PHONY: $(ALL) $(SUB) format

# Recursive targets.
build: build.sub
build-linux: build-linux.sub
build-darwin: build-darwin.sub
clean: clean.sub
docker: docker.sub
lint: lint.sub
push: push.sub
test: test.sub

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
