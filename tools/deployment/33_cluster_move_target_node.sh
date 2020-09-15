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
export KUBECONFIG_EPHEMERAL_CONTEXT=${KUBECONFIG_EPHEMERAL_CONTEXT:-"ephemeral-context"}
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-context"}

echo "Switch context to ephemeral cluster and set manifest"
airshipctl config use-context ephemeral-context
airshipctl config set-context ephemeral-context --manifest dummy_manifest

echo "Check Cluster Status"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT get cluster target-cluster -o json | jq '.status.controlPlaneReady'

echo "Annotate BMH for target node"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT annotate bmh node01 baremetalhost.metal3.io/paused=true

echo "Move Cluster Object to Target Cluster"
clusterctl --v 20 move --kubeconfig $KUBECONFIG --kubeconfig-context $KUBECONFIG_EPHEMERAL_CONTEXT --to-kubeconfig $KUBECONFIG --to-kubeconfig-context $KUBECONFIG_TARGET_CONTEXT

echo "Switch context to target cluster and set manifest"
airshipctl config use-context target-context
airshipctl config set-context target-context --manifest dummy_manifest

echo "Waiting for pods to be ready"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait --all-namespaces --for=condition=Ready pods --all --timeout=3000s
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get pods --all-namespaces

#Wait till crds are created
end=$(($(date +%s) + $TIMEOUT))
echo "Waiting $TIMEOUT seconds for crds to be created."
while true; do
    if (kubectl --request-timeout 20s --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get cluster target-cluster -o json | jq '.status.controlPlaneReady' | grep -q true) ; then
        echo -e "\nGet CRD status"
        kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get bmh
        kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get machines
        kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get clusters
        break
    else
        now=$(date +%s)
        if [ $now -gt $end ]; then
            echo -e "\nCluster move failed and CRDs was not ready before TIMEOUT."
            exit 1
        fi
        echo -n .
        sleep 15
    fi
done
