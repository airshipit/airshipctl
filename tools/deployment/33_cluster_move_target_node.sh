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

echo "Switch context to ephemeral cluster and set manifest"
airshipctl config use-context dummy_cluster
airshipctl config set-context dummy_cluster --manifest dummy_manifest

echo "Check Cluster Status"
kubectl --kubeconfig $KUBECONFIG get cluster target-cluster -o json | jq '.status.controlPlaneReady'

echo "Annotate BMH for target node"
kubectl --kubeconfig $KUBECONFIG annotate bmh node01 baremetalhost.metal3.io/paused=true

echo "Move Cluster Object to Target Cluster"
clusterctl --v 20 move --kubeconfig $KUBECONFIG --kubeconfig-context dummy_cluster --to-kubeconfig $KUBECONFIG --to-kubeconfig-context  target-cluster-admin@target-cluster

echo "Switch context to target cluster and set manifest"
airshipctl config use-context target-cluster-admin@target-cluster
airshipctl config set-context target-cluster-admin@target-cluster --manifest dummy_manifest

echo "Waiting for pods to be ready"
kubectl --kubeconfig $KUBECONFIG wait --all-namespaces --for=condition=Ready pods --all --timeout=3000s
kubectl --kubeconfig $KUBECONFIG get pods --all-namespaces

#Wait till crds are created
end=$(($(date +%s) + $TIMEOUT))
echo "Waiting $TIMEOUT seconds for crds to be created."
while true; do
    if (kubectl --request-timeout 20s --kubeconfig $KUBECONFIG get cluster target-cluster -o json | jq '.status.controlPlaneReady' | grep -q true) ; then
        echo -e "\nGet CRD status"
        kubectl --kubeconfig $KUBECONFIG get bmh
        kubectl --kubeconfig $KUBECONFIG get machines
        kubectl --kubeconfig $KUBECONFIG get clusters
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
