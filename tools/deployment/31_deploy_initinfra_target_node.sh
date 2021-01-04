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
NODENAME="node01"
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-cluster"}

echo "Deploy calico using tigera operator"
airshipctl phase run initinfra-networking-target --debug

echo "Wait for Calico to be deployed using tigera"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait --all-namespaces --for=condition=Ready pods --all --timeout=600s

echo "Wait for Tigerastatus to be available"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait --for=condition=Available tigerastatus --all --timeout=600s -A

echo "Deploy infra to cluster"
airshipctl phase run initinfra-target --debug

echo "List all pods"
kubectl \
  --kubeconfig $KUBECONFIG \
  --context $KUBECONFIG_TARGET_CONTEXT \
   get pods \
  --all-namespaces
