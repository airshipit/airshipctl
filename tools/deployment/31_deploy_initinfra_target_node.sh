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

export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export TIMEOUT=${TIMEOUT:-60}
NODENAME="node01"
export WAIT_TIMEOUT=${WAIT_TIMEOUT:-"2000s"}
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-context"}

# TODO need to run another config command after use-context to update kubeconfig
echo "Switch context to target cluster and set manifest"
airshipctl config use-context target-context
airshipctl config set-context target-context --manifest dummy_manifest

end=$(($(date +%s) + $TIMEOUT))
echo "Waiting $TIMEOUT seconds for $NODENAME to be created."
while true; do
  if (kubectl --request-timeout 10s --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get nodes | grep -q $NODENAME) ; then
    echo -e "\n$NODENAME found"
    break
  else
    now=$(date +%s)
    if [ $now -gt $end ]; then
      echo -e "\n$NODENAME was not ready before TIMEOUT."
      exit 1
    fi
    echo -n .
    sleep 10
  fi
done

# TODO remove taint
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT taint node $NODENAME node-role.kubernetes.io/master-

echo "Deploy infra to cluster"
airshipctl phase apply initinfra --debug --wait-timeout $WAIT_TIMEOUT

echo "List all pods"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get pods --all-namespaces
