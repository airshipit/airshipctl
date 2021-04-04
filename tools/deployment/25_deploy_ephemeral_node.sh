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

echo "Deploy ephemeral node using redfish with iso"
airshipctl phase run remotedirect-ephemeral --debug

echo "Wait for apiserver to become available"
# Scripts for this phase placed in manifests/function/phase-helpers/wait_node/
# To get ConfigMap for this phase, execute `airshipctl phase render --source config -k ConfigMap`
# and find ConfigMap with name kubectl-get-node
airshipctl phase run kubectl-wait-node-ephemeral --debug

echo "List all pods"
# Scripts for this phase placed in manifests/function/phase-helpers/get_pods/
# To get ConfigMap for this phase, execute `airshipctl phase render --source config -k ConfigMap`
# and find ConfigMap with name kubectl-get-pods
airshipctl phase run kubectl-get-pods-ephemeral --debug
