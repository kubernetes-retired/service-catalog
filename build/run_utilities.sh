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


ROOT="${ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"
. "${ROOT}/build/utilities.sh" || { echo 'Cannot load Bash utilities'; exit 1; }

BINDIR=${ROOT}/bin
GOPATH=${GOPATH:-${ROOT%/src/github.com/kubernetes-incubator/service-catalog}}
LOG_ROOT="${ROOT}/log"
PID_ROOT="${ROOT}/pid"
PKG_ROOT=github.com/kubernetes-incubator/service-catalog

function shutdown_processes() {
  echo "Shutting down local processes."
  for pid_file in $(ls ${PID_ROOT}); do
    pkill -F ${PID_ROOT}/${pid_file} \
      || error_exit "Cannot shut down ${pid_file}"

    echo "${pid_file} shut down"

    rm ${PID_ROOT}/${pid_file} \
      || error_exit "Cannot remove ${pid_file}"
  done
}

function run_main() {
  local k8s_registry_port=8101
  local k8s_broker_port=8001
  local user_provided_broker_port=8002
  local controller_port=10000
  local kubeconfig=${KUBECONFIG:-"${ROOT}/kubeconfig"}

  local ignore_controller='no'
  local shutdown='no'

  while [[ $# -gt 0 ]]; do
    case "$1" in
      -s|--shutdown)           shutdown='yes' ;;
      -k|--kubeconfig)         kubeconfig="${2:-}"; shift ;;

      -*) error_exit "Unrecognized command line flag $1" ;;

      # End of arguments
      *) break ;;
    esac
    shift
  done

  # Make sure KUBECONFIG file exists if supplied.
  [[ -z "${kubeconfig}" || -f "${kubeconfig}" ]] \
    || error_exit "Cannot find KUBECONFIG file '${kubeconfig}'."

  # Build
  make -C "${ROOT}" init build \
    || error_exit 'Failed to build.'


  # Shut down servers on exit if needed.
  if [[ "${shutdown}" == 'yes' ]]; then
    trap shutdown_processes EXIT
  fi

  # Run services.
  function deploy() {
    local title="$1"
    local log="$2"
    local pid_file="$3"
    local url="$4"
    local process="$5"

    if [[ -f "${pid_file}" ]]; then
      pkill -F "${pid_file}"
    else
      echo "${pid_file} not found"
    fi

    shift 4 # Drop title, url, log and pid name.

    echo "Deploying ${title}: \"${@}\""
    "${@}" &> "${log}" &

    local pid=$!

    echo " ... waiting for ${title}"
    retry -n 20 -s 1 -t 1 \
        curl --silent --fail "${url}" > /dev/null \
      || { echo "Cannot reach ${url}.\n\nLog file: "; cat "${log}"; return 1; }

    # Check if the process stayed running:
    ps ${pid} > /dev/null \
      || { echo "${title} died."; return 1; }

    echo ${pid} > "${pid_file}"
    echo " ... done"

    return 0
  }

  # Make sure logging directory exists.
  mkdir -p "${LOG_ROOT}"
  mkdir -p "${PID_ROOT}"

  # Kubernetes registry.

  deploy 'Kubernetes Registry' "${LOG_ROOT}/k8s_registry.txt" "${PID_ROOT}/k8s_registry.pid" \
      "http://localhost:${k8s_registry_port}/services" \
      "${BINDIR}/registry" \
      --port ${k8s_registry_port} \
      --definitions ${GOPATH}/src/${PKG_ROOT}/registry/data/charts/definitions.json \
    || error_exit 'Registry failed to start.'

  # Brokers

  # K8s
  deploy 'Kubernetes broker' "${LOG_ROOT}/k8s.txt" "${PID_ROOT}/k8s.pid" \
      "http://localhost:${k8s_broker_port}/v2/catalog" \
      "${BINDIR}/k8s-broker" \
      --port ${k8s_broker_port} \
      --registry_port ${k8s_registry_port} \
      --helm_binary '/usr/local/bin/helm' \
    || error_exit 'Kubernetes broker failed to start.'

  # User-provided
  deploy 'User broker' "${LOG_ROOT}/user_provided.txt" "${PID_ROOT}/user_provided.pid" \
      "http://localhost:${user_provided_broker_port}/v2/catalog" \
      "${BINDIR}/user-broker" \
      --port ${user_provided_broker_port} \
    || error_exit 'User broker failed to start.'

  # Controller.
  deploy 'Controller' "${LOG_ROOT}/controller.txt" "${PID_ROOT}/controller.pid" \
      "http://localhost:${controller_port}/v2/service_instances" \
      "${BINDIR}/controller" \
      --port ${controller_port} \
      --kubeconfig "${kubeconfig}" \
    || error_exit 'Controller failed to start.'

  # Request catalog from broker.

  printf '\nCatalog for Kubernetes broker:\n'
  curl --silent --fail localhost:${k8s_broker_port}/v2/catalog \
    | head -c 100 \
    || error_exit 'Failed to obtain Kubernetes broker catalog.'

  printf '\nCatalog for User Provided broker:\n'
  curl --silent --fail localhost:${user_provided_broker_port}/v2/catalog \
    | head -c 100 \
    || error_exit 'Failed to obtain User Provided broker catalog.'

  printf '\n'
}
