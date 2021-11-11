#!/bin/bash

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
# This starts up a kubernetes cluster which is sufficient for
# assisting with tasks like `kubectl apply --dry-run` style validation
# Usage
# example 1: create a kind cluster, with name as airship
#
#              ./tools/deployment/kind/start_kind.sh
#
# example 2: create a kind cluster, with a custom name
#
#              CLUSTER=ephemeral-cluster ./tools/deployment/kind/start_kind.sh
#
# example 3: create a kind cluster, using custom name and config
#
#              CLUSTER=ephemeral-cluster KIND_CONFIG=./tools/deployment/templates/kind-cluster-with-extramounts.yaml \
#              ./tools/deployment/kind/start_kind.sh
#
# example 4: create a kind cluster with name as airship, using custom config
#
#              KIND_CONFIG=./tools/deployment/templates/kind-cluster-with-extramounts.yaml \
#              ./tools/deployment/kind/start_kind.sh

: ${KIND:="/usr/local/bin/kind"}
: ${CLUSTER:="airship"} # NB: kind prepends "kind-"
: ${KUBECONFIG:="${HOME}/.airship/kubeconfig"}
: ${TIMEOUT:=3600}
: ${KIND_CONFIG:=""}

export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-cluster"}

echo "cluster name: $CLUSTER";

if [ -z "$KIND_CONFIG" ]
then
      ${KIND} create cluster --name $CLUSTER --wait 120s --kubeconfig ${KUBECONFIG}
else
      echo "Using kind configuration file: $KIND_CONFIG"
      ${KIND} create cluster --config $KIND_CONFIG --name $CLUSTER --wait 120s --kubeconfig ${KUBECONFIG}
fi

# Wait till Control Plane Node is ready
kubectl wait --for=condition=ready node --all --timeout=1000s --kubeconfig $KUBECONFIG -A

# Add context <cluster> e.g ephemeral-cluster.
# By default, kind creates context kind-<cluster_name>
kubectl config set-context ${CLUSTER} --cluster kind-${CLUSTER} --user kind-${CLUSTER} --kubeconfig $KUBECONFIG

# Add context for target-cluster
kubectl config set-context ${KUBECONFIG_TARGET_CONTEXT} --user target-cluster-admin --cluster ${KUBECONFIG_TARGET_CONTEXT} --kubeconfig  $KUBECONFIG

echo "This cluster can be deleted via:"
echo "${KIND} delete cluster --name ${CLUSTER}"
