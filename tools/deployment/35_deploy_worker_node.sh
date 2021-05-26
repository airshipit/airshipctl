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

set -e

# all vms. This can be removed once sushy tool is fixed
# Scripts for this phase placed in manifests/function/phase-helpers/virsh-destroy-vms/
# To get ConfigMap for this phase, execute `airshipctl phase render --source config -k ConfigMap`
# and find ConfigMap with name virsh-destroy-vms
airshipctl phase run virsh-destroy-vms --debug

echo "Deploy worker node"
airshipctl phase run workers-target --debug

# Waiting for node to be provisioned."
# Scripts for this phase placed in manifests/function/phase-helpers/wait_label_node/
# To get ConfigMap for this phase, execute `airshipctl phase render --source config -k ConfigMap`
# and find ConfigMap with name kubectl-wait-label-node
airshipctl phase run kubectl-wait-label-node-target --debug
