#!/usr/bin/env bash

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

set -xe

#Default wait timeout is 3600 seconds
export TIMEOUT=${TIMEOUT:-3600}
export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export KUBECONFIG_EPHEMERAL_CONTEXT=${KUBECONFIG_EPHEMERAL_CONTEXT:-"ephemeral-cluster"}

echo "Deploy ephemeral node using redfish with iso"
airshipctl baremetal remotedirect --debug

echo "Wait for apiserver to become available"
N=0
MAX_RETRY=30
DELAY=60
until [ "$N" -ge ${MAX_RETRY} ]
do
  if timeout 20 kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT get node; then
      break
  fi

  N=$((N+1))
  echo "$N: Retrying to reach the apiserver"
  sleep ${DELAY}
done

if [ "$N" -ge ${MAX_RETRY} ]; then
  echo "Could not reach the apiserver"
  exit 1
fi

echo "List all pods"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT get pods --all-namespaces
