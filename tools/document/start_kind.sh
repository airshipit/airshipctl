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

# This starts up a kubernetes cluster which is sufficient for
# assisting with tasks like `kubectl apply --dry-run` style validation
set -e

: ${KIND:="/usr/local/bin/kind"}
: ${CLUSTER:="airship"} # NB: kind prepends "kind-"
: ${KUBECONFIG:="${HOME}/.airship/kubeconfig"}

${KIND} create cluster --name ${CLUSTER} \
    --kubeconfig ${KUBECONFIG}

echo "This cluster can be deleted via:"
echo "${KIND} delete cluster --name ${CLUSTER}"
