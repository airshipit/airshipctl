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

# Example Usage
# KUBECONFIG=/tmp/$KUBECONFIG \
# ./tools/deployment/provider_common/32_cluster_init_target_node.sh

export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-cluster"}

# Get control plane node
CONTROL_PLANE_NODES=( $(kubectl --context $KUBECONFIG_TARGET_CONTEXT --kubeconfig $KUBECONFIG get --no-headers=true nodes \
| grep cluster-control-plane | awk '{print $1}') )

# Remove noschedule taint to prevent cluster init from timing out
for i in "${CONTROL_PLANE_NODES}"; do
  echo untainting node $i
  kubectl taint node $i node-role.kubernetes.io/master- --context $KUBECONFIG_TARGET_CONTEXT --kubeconfig $KUBECONFIG --request-timeout 10s
done

echo "Deploy CAPI components to target cluster"
airshipctl phase run clusterctl-init-target --debug

echo "Waiting for pods to be ready"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait --all-namespaces --for=condition=Ready pods --all --timeout=600s
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get pods --all-namespaces
