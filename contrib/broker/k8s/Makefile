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

BIN=k8s-broker
PKG=github.com/kubernetes-incubator/service-catalog/broker/k8s
DOCKER=$(BIN)

include ../../script/Makefile.mk

HELM_VERSION := v2.0.0-rc.2

docker::
	$(ECHO) mkdir -p $(BINDIR)/linux_$(ARCH)
	$(ECHO) cp $(ROOT)/deps/helm-$(HELM_VERSION)-linux-amd64.tar.gz $(BINDIR)/linux_$(ARCH)/

clean::
	$(ECHO) rm -f $(BINDIR)/linux_$(ARCH)/helm-$(HELM_VERSION)-linux-amd64.tar.gz

include ../../script/Common.mk
