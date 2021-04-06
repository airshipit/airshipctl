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

export IMAGE_DIR=${IMAGE_DIR:-"/srv/images"}
export SERVE_PORT=${SERVE_PORT:-"8099"}
export AIRSHIPCTL_WS=${AIRSHIPCTL_WS:-$PWD}
export USER_NAME=${USER:-"ubuntu"}
export USE_PROXY=${USE_PROXY:-"false"}
export HTTPS_PROXY=${HTTPS_PROXY:-${https_proxy}}
export HTTP_PROXY=${HTTP_PROXY:-${http_proxy}}
export NO_PROXY=${NO_PROXY:-${no_proxy}}
export AIRSHIP_CONFIG_ISO_GEN_TARGET_PATH=${IMAGE_DIR}
export AIRSHIP_CONFIG_ISO_BUILDER_DOCKER_IMAGE=${BUILDER_IMAGE:-"quay.io/airshipit/image-builder:053c992218601cc49fd4a595ee4873380b132408-ubuntu_focal"}
export REMOTE_TYPE=redfish
export REMOTE_INSECURE=true
export REMOTE_PROXY=false
export AIRSHIP_CONFIG_ISO_SERVE_HOST=${HOST:-"localhost"}
export AIRSHIP_CONFIG_ISO_PORT=${SERVE_PORT}
export AIRSHIP_CONFIG_ISO_NAME=${ISO_NAME:-"ephemeral.iso"}
export SITE=${SITE:-"test-site"}
export AIRSHIP_CONFIG_METADATA_PATH=${AIRSHIP_CONFIG_METADATA_PATH:-"manifests/site/${SITE}/metadata.yaml"}
export SYSTEM_ACTION_RETRIES=30
export SYSTEM_REBOOT_DELAY=30
export AIRSHIP_CONFIG_PHASE_REPO_BRANCH=${BRANCH:-"master"}
# the git repo url or local file system path to a cloned repo, e.g., /home/stack/airshipctl
export AIRSHIP_CONFIG_PHASE_REPO_URL=${AIRSHIP_CONFIG_PHASE_REPO_URL:-"https://review.opendev.org/airship/airshipctl"}
export AIRSHIP_CONFIG_PHASE_REPO_NAME=${AIRSHIP_CONFIG_PHASE_REPO_NAME:-"airshipctl"}
export AIRSHIP_CONFIG_MANIFEST_DIRECTORY=${AIRSHIP_CONFIG_MANIFEST_DIRECTORY:-"/tmp/airship"}
export EXTERNAL_KUBECONFIG=${EXTERNAL_KUBECONFIG:-""}

# Remove the contents of the .airship folder, preserving the kustomize plugin directory
rm -rf $HOME/.airship/config
mkdir -p $HOME/.airship

echo "Generate ~/.airship/config and ~/.airship/kubeconfig"
envsubst <"${AIRSHIPCTL_WS}/tools/deployment/templates/airshipconfig_template" > ~/.airship/config

if [[ -z "$EXTERNAL_KUBECONFIG" ]]; then
   # make airshipctl happy
   touch ~/.airship/kubeconfig
fi
