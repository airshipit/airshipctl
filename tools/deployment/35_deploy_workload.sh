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

#Default wait timeout is 600 seconds
export TIMEOUT=${TIMEOUT:-600s}
export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-cluster"}
export TARGET_IP=${TARGET_IP:-"10.23.25.102"}
export TARGET_PORT=${TARGET_PORT:-"30000"}

echo "Deploy workload"
airshipctl phase run workload-target --debug

echo "Ensure we can reach ingress controller default backend"
if [ "404" != "$(curl --head --write-out '%{http_code}' --silent --output /dev/null $TARGET_IP:$TARGET_PORT/should-404)" ]; then
    echo -e "\nFailed to reach ingress controller default backend."
    exit 1
fi
