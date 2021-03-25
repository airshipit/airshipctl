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
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-cluster"}

echo "Stop/Delete ephemeral node"
kind delete cluster --name "ephemeral-cluster"

echo "Deploy worker node"
airshipctl phase run  workers-target --debug

#Wait till node is created
kubectl wait --for=condition=ready node --all --timeout=1000s --context $KUBECONFIG_TARGET_CONTEXT --kubeconfig $KUBECONFIG -A
