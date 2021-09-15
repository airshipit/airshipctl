#!/bin/bash

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

set -ex

# Create the "canary" file, indicating that the container is healthy
mkdir -p /tmp/healthy
touch /tmp/healthy/infra-builder

success=false
function cleanup() {
  if [[ "$success" == "false" ]]; then
    rm /tmp/healthy/infra-builder
  fi
}

function check_libvirt_readiness() {
  timeout=300

  #add wait condition
  end=$(($(date +%s) + $timeout))
  echo "Waiting $timeout seconds for libvirt to be ready."
  while true; do
    if ( virsh version | grep 'library' ); then
      echo "libvrit is now ready"
      break
    else
      echo "libvirt is not ready"
    fi
    now=$(date +%s)
    if [ $now -gt $end ]; then
      echo -e "\n Libvirt failed to become ready within a reasonable timeframe."
      exit 1
    fi
    sleep 10
  done
}

trap cleanup EXIT

check_libvirt_readiness

ansible-playbook -v /opt/ansible/playbooks/build-infra.yaml \
  -e local_src_dir="$(pwd)"

success=true
/signal_complete infra-builder

# Keep the container running for debugging/monitoring purposes
sleep infinity
