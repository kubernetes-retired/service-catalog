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

BIN=registry
PKG=github.com/kubernetes-incubator/service-catalog/registry
DOCKER=$(BIN)

include ../hack/Makefile.mk

docker:: ./data/charts/definitions.json
	$(ECHO) mkdir -p $(BINDIR)/linux_$(ARCH)
	$(ECHO) cp ./data/charts/definitions.json $(BINDIR)/linux_$(ARCH)/definitions.json

clean::
	$(ECHO) rm -f $(BINDIR)/linux_$(ARCH)/definitions.json

include ../hack/Common.mk
