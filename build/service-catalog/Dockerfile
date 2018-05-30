# Copyright 2017 The Kubernetes Authors.
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

# Build a Debian-based image to copy root CA certificates from it
# to a scratch-based image using Docker multi-stage builds
FROM BASEIMAGE as base

RUN export DEBIAN_FRONTEND=noninteractive && \
    apt-get update && \
    apt-get install ca-certificates -y && \
    rm -rf /var/lib/apt/lists/*

# Build actual scratch-based Service Catalog image with root CA certificates
FROM scratch

COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ADD tmp /tmp

ADD service-catalog opt/services/

ENTRYPOINT ["/opt/services/service-catalog"]
