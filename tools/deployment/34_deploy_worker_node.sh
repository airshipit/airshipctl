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

echo "Switch context to target cluster and set manifest"
airshipctl config use-context target-cluster-admin@target-cluster
airshipctl config set-context target-cluster-admin@target-cluster --manifest dummy_manifest

echo "Stop ephemeral node"
sudo virsh destroy air-ephemeral

echo "Deploy worker node"
airshipctl phase apply workers --debug

#Wait till node is created
end=$(($(date +%s) + $TIMEOUT))
echo "Waiting $TIMEOUT seconds for node to be provisioned."
while true; do
    if (kubectl --request-timeout 20s --kubeconfig $KUBECONFIG get node node03 | grep -qw Ready) ; then
        echo -e "\nGet node status"
        kubectl --kubeconfig $KUBECONFIG get node
        break
    else
        now=$(date +%s)
        if [ $now -gt $end ]; then
            echo -e "\nWorker node is not ready before TIMEOUT."
            exit 1
        fi
        echo -n .
        sleep 15
    fi
done
