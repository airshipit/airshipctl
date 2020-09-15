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
export KUBECONFIG_EPHEMERAL_CONTEXT=${KUBECONFIG_EPHEMERAL_CONTEXT:-"ephemeral-cluster"}

echo "Deploy metal3.io components to ephemeral node"
airshipctl phase run initinfra-ephemeral --debug

echo "Getting metal3 pods as debug information"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT --namespace metal3 get pods

echo "Deploy cluster-api components to ephemeral node"
airshipctl phase run clusterctl-init-ephemeral --debug

echo "Waiting for clusterapi pods to come up"
kubectl --kubeconfig $KUBECONFIG  --context $KUBECONFIG_EPHEMERAL_CONTEXT wait --for=condition=available deploy --all --timeout=1000s -A
kubectl --kubeconfig $KUBECONFIG  --context $KUBECONFIG_EPHEMERAL_CONTEXT get pods --all-namespaces
