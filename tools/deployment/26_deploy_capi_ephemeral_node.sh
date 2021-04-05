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
export PROVIDER=${PROVIDER:-"metal3"}


if [ "$PROVIDER" = "metal3" ]; then

    echo "Deploy calico using tigera operator"
    airshipctl phase run initinfra-networking-ephemeral --debug

    # "Wait for Calico to be deployed using tigera"
    # Scripts for this phase placed in manifests/function/phase-helpers/wait_tigera/
    # To get ConfigMap for this phase, execute `airshipctl phase render --source config -k ConfigMap`
    # and find ConfigMap with name kubectl-wait_tigera
    airshipctl phase run kubectl-wait-tigera-ephemeral --debug

    echo "Deploy metal3.io components to ephemeral node"
    airshipctl phase run initinfra-ephemeral --debug

    echo "Getting pods as debug information"
    # Scripts for this phase placed in manifests/function/phase-helpers/get_pods/
    # To get ConfigMap for this phase, execute `airshipctl phase render --source config -k ConfigMap`
    # and find ConfigMap with name kubectl-get-pods
    airshipctl phase run kubectl-get-pods-ephemeral --debug

    echo "Deploy cluster-api components to ephemeral node"
    airshipctl phase run clusterctl-init-ephemeral --debug
else
    echo "Deploy cluster-api components to ephemeral node"
    airshipctl phase run clusterctl-init-ephemeral --debug
    # Scripts for this phase placed in manifests/function/phase-helpers/get_pods/
    # To get ConfigMap for this phase, execute `airshipctl phase render --source config -k ConfigMap`
    # and find ConfigMap with name kubectl-get-pods
    airshipctl phase run kubectl-get-pods-ephemeral --debug
fi

echo "Waiting for clusterapi pods to come up"
# Scripts for this phase placed in manifests/function/phase-helpers/wait_deploy/
# To get ConfigMap for this phase, execute `airshipctl phase render --source config -k ConfigMap`
# and find ConfigMap with name kubectl-wait-deploy
airshipctl phase run kubectl-wait-deploy-ephemeral --debug

# Scripts for this phase placed in manifests/function/phase-helpers/get_pods/
# To get ConfigMap for this phase, execute `airshipctl phase render --source config -k ConfigMap`
# and find ConfigMap with name kubectl-get-pods
airshipctl phase run kubectl-get-pods-ephemeral --debug
