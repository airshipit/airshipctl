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

echo "Generating secrets using airshipctl"
airshipctl phase run secret-generate

export AIRSHIP_CONFIG_MANIFEST_DIRECTORY=${AIRSHIP_CONFIG_MANIFEST_DIRECTORY:-"/tmp/airship"}
export AIRSHIP_CONFIG_PHASE_REPO_URL=${AIRSHIP_CONFIG_PHASE_REPO_URL:-"https://review.opendev.org/airship/airshipctl"}
export EXTERNAL_KUBECONFIG=${EXTERNAL_KUBECONFIG:-""}

echo "Generating ~/.airship/kubeconfig"
if [[ -z "$EXTERNAL_KUBECONFIG" ]]; then
   # TODO: use airshipctl cluster get-kubeconfig command when it's implemented
   KUSTOMIZE_PLUGIN_HOME=./ kustomize build --enable_alpha_plugins "${AIRSHIP_CONFIG_MANIFEST_DIRECTORY}/$(basename ${AIRSHIP_CONFIG_PHASE_REPO_URL})/manifests/site/test-site/kubeconfig/" | yq '.config' --yaml-output > ~/.airship/kubeconfig
fi
