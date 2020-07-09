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

# TODO need to run another config command after use-context to update kubeconfig
echo "Switch context to target cluster and set manifest"
airshipctl config use-context target-cluster-admin@target-cluster
airshipctl config set-context target-cluster-admin@target-cluster --manifest dummy_manifest

echo "Waiting for all control pods to come up"
kubectl --kubeconfig $KUBECONFIG wait -n kube-system --for=condition=Ready pods --selector tier=control-plane --timeout=600s
kubectl --kubeconfig $KUBECONFIG --namespace metal3 get pods --selector tier=control-plane "--output=jsonpath={.items[*].metadata.name}"

# TODO remove taint
kubectl --kubeconfig $KUBECONFIG taint node node01 node-role.kubernetes.io/master-

echo "Deploy infra to cluster"
airshipctl phase apply initinfra --debug --wait-timeout 600s

kubectl --kubeconfig $KUBECONFIG get pods --all-namespaces
