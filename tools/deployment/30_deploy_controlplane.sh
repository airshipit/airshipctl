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

set -ex

EPHEMERAL_DOMAIN_NAME="air-ephemeral"
export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-cluster"}

# TODO (dukov) this is needed due to sushy tools inserts cdrom image to
# all vms. This can be removed once sushy tool is fixed
echo "Ensure all cdrom images are ejected."
for vm in $(sudo virsh list --all --name |grep -v ${EPHEMERAL_DOMAIN_NAME})
do
  sudo virsh domblklist $vm |
    awk 'NF==2 {print $1}' |
    grep -v Target |
    xargs -I{} sudo virsh change-media $vm {} --eject || :
done

echo "Create target k8s cluster resources"
airshipctl phase run controlplane-ephemeral --debug

echo "List all nodes in target cluster"
kubectl \
  --kubeconfig $KUBECONFIG \
  --context $KUBECONFIG_TARGET_CONTEXT \
  --request-timeout 10s \
  get node


echo "List all pods in target cluster"
kubectl \
  --kubeconfig  $KUBECONFIG \
  --context $KUBECONFIG_TARGET_CONTEXT \
  --request-timeout 10s \
  get pods \
  --all-namespaces
