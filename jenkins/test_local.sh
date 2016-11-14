#!/bin/bash
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


set -ux

. "${ROOT}/script/run_utilities.sh" || { echo 'Cannot load run utilities.'; exit 1; }

#trap shutdown_processes EXIT
run_main \
  || error_exit 'Local initialization failed.'

# Test CLI is up
echo 'Testing Service Controller'
OUTPUT=$(${GOPATH}/bin/sc) || error_exit 'Service Controller not present.'

# Test creating both brokers
echo 'Testing creating brokers'
${GOPATH}/bin/sc brokers create k8s http://localhost:8002
${GOPATH}/bin/sc brokers create ups http://localhost:8003

if [[ $(${GOPATH}/bin/sc brokers list) != *k8s*ups* ]]; then
  error_exit 'Service brokers not created successfully.'
fi

# Test default definitions are present
echo 'Testing default service definitions'
OUTPUT=$(${GOPATH}/bin/sc inventory)
if [[ "${OUTPUT}" != *booksbe* ]] && \
    [[ "${OUTPUT}" != *booksfe* ]]; then
  error_exit 'Expected service definitions not present.'
fi

# Test creating service instances
echo 'Testing creating service instances'
$GOPATH/bin/sc service-instances create backend booksbe
$GOPATH/bin/sc service-instances create frontend booksfe

OUTPUT=$(${GOPATH}/bin/sc service-instances list)
if [[ "${OUTPUT}" != *backend*booksbe* ]] && \
    [[ "${OUTPUT}" != *frontend*booksfe* ]]; then
  error_exit 'Service instances not created successfully.'
fi

# Test creating service binding
echo 'Testing creating service bindings'
${GOPATH}/bin/sc service-bindings create database frontend backend

if [[ $(${GOPATH}/bin/sc service-bindings list) != *database* ]]; then
  error_exit 'Service binding not created successfully.'
fi

sleep 20 #Wait for services to come up

OUTPUT=$(kubectl get services)
if [[ "${OUTPUT}" != *booksfe* ]] || \
    [[ "${OUTPUT}" != *booksbe* ]]; then
  error_exit 'Service instances not created successfully in cluster.'
fi

# Wait for external IP address to be assigned
COUNT=0
INTERVAL=5
LIMIT=100
while [[ $(kubectl get services) == *pending* ]]; do
  sleep ${INTERVAL}
  (( COUNT+=${INTERVAL} ))
  if [[ ${COUNT} -gt ${LIMIT} ]]; then
    error_exit 'Frontend service took unexpected amount of time to get external IP.'
  fi
done

IP=$(echo $(kubectl get services) | sed 's/.*booksfe [0-9.]* \([0-9.]*\).*/\1/')

sleep 30

# TODO: Currently the local instantiation is broken. It will
# get assigned an IP, but no curls will work against it.

# Get list of shelves:
curl http://${IP}:8080/shelves \
  || error_exit 'Cannot get list of shelves.'

# Create a new shelf:
curl -H 'Content-Type: application/json' \
    -d '{ "theme": "Travel" }' \
    http://${IP}:8080/shelves \
  || error_exit 'Cannot create new shelf.'

# Create a book on the shelf:
curl -H 'Content-Type: application/json' \
    -d '{ "author": "Rick Steves", "title": "Travel as a Political
    Act" }' \
    http://${IP}:8080/shelves/3/books \
  || error_exit 'Cannot create book on shelf.'

# List the books on the travel shelf:
curl http://${IP}:8080/shelves/3/books \
  || error_exit 'Cannot list books on shelf.'

# Get the book:
curl http://${IP}:8080/shelves/3/books/3 \
  || error_exit 'Cannot get book.'

# Delete the book:
curl -X DELETE http://${IP}:8080/shelves/3/books/3 \
  || error_exit 'Cannot delete book.'

# Delete the shelf:
curl -X DELETE http://${IP}:8080/shelves/3 \
  || error_exit 'Cannot delete shelf.

echo "TODO: ADD MORE TESTS!"

