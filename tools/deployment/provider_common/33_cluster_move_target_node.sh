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
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-cluster"}
export TARGET_KUBECONFIG="/tmp/target-cluster.kubeconfig"

echo "Waiting for machines to come up"
kubectl --kubeconfig ${KUBECONFIG} wait --for=condition=Ready machines --all --timeout 4000s

#add wait condition
end=$(($(date +%s) + $TIMEOUT))
echo "Waiting $TIMEOUT seconds for Machine to be Running."
while true; do
    if (kubectl --request-timeout 20s --kubeconfig $KUBECONFIG get machines -o json | jq '.items[0].status.phase' | grep -q "Running") ; then
        echo -e "\nMachine is Running"
        kubectl --kubeconfig $KUBECONFIG get machines
        break
    else
        now=$(date +%s)
        if [ $now -gt $end ]; then
            echo -e "\nMachine is not in Ruuning phase, Move can't be performed."
            exit 1
        fi
        echo -n .
        sleep 15
    fi
done

echo "Move Cluster Object to Target Cluster"
KUBECONFIG=$KUBECONFIG:$TARGET_KUBECONFIG kubectl config view --merge --flatten > "/tmp/merged_target_ephemeral.kubeconfig"
airshipctl phase run clusterctl-move --kubeconfig "/tmp/merged_target_ephemeral.kubeconfig"

echo "Waiting for pods to be ready"
kubectl --kubeconfig $TARGET_KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait --all-namespaces --for=condition=Ready pods --all --timeout=3000s
kubectl --kubeconfig $TARGET_KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get pods --all-namespaces

#Wait till crds are created
end=$(($(date +%s) + $TIMEOUT))
echo "Waiting $TIMEOUT seconds for crds to be created."
while true; do
    if (kubectl --request-timeout 20s --kubeconfig $TARGET_KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get cluster target-cluster -o json | jq '.status.controlPlaneReady' | grep -q true) ; then
        echo -e "\nGet CRD status"
        kubectl --kubeconfig $TARGET_KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get machines
        kubectl --kubeconfig $TARGET_KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get clusters
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
