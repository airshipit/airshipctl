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

export ISO_DIR=${ISO_DIR:-"/srv/iso"}
export SERVE_PORT=${SERVE_PORT:-"8099"}
export AIRSHIPCTL_WS=${AIRSHIPCTL_WS:-$PWD}
export USER_NAME=${USER:-"ubuntu"}
export USE_PROXY=${USE_PROXY:-"false"}
export HTTPS_PROXY=${HTTPS_PROXY:-${https_proxy}}
export HTTP_PROXY=${HTTP_PROXY:-${http_proxy}}
export NO_PROXY=${NO_PROXY:-${no_proxy}}
export AIRSHIP_CONFIG_ISO_GEN_TARGET_PATH=${ISO_DIR}
export AIRSHIP_CONFIG_ISO_BUILDER_DOCKER_IMAGE=${BUILDER_IMAGE:-"quay.io/airshipit/isogen:latest-ubuntu_focal"}
export REMOTE_TYPE=redfish
export REMOTE_INSECURE=true
export REMOTE_PROXY=false
export AIRSHIP_CONFIG_ISO_SERVE_HOST=${HOST:-"localhost"}
export AIRSHIP_CONFIG_ISO_PORT=${SERVE_PORT}
export AIRSHIP_CONFIG_ISO_NAME=${ISO_NAME:-"ubuntu-focal.iso"}
export AIRSHIP_CONFIG_METADATA_PATH=${AIRSHIP_CONFIG_METADATA_PATH:-"manifests/metadata.yaml"}
export SYSTEM_ACTION_RETRIES=30
export SYSTEM_REBOOT_DELAY=30
export AIRSHIP_CONFIG_PHASE_REPO_BRANCH=${BRANCH:-"master"}
# the git repo url or local file system path to a cloned repo, e.g., /home/stack/airshipctl
export AIRSHIP_CONFIG_PHASE_REPO_URL=${AIRSHIP_CONFIG_PHASE_REPO_URL:-"https://review.opendev.org/airship/airshipctl"}
export AIRSHIP_CONFIG_PHASE_REPO_NAME=${AIRSHIP_CONFIG_PHASE_REPO_NAME:-"airshipctl"}
export AIRSHIP_CONFIG_MANIFEST_DIRECTORY=${AIRSHIP_CONFIG_MANIFEST_DIRECTORY:-"/tmp/airship"}
export EPHEMERAL_CONFIG_CA_DATA=$(cat tools/deployment/certificates/ephemeral_config_ca_data| base64 -w0)
export EPHEMERAL_IP=${EPHEMERAL_IP:-"10.23.25.101"}
export EPHEMERAL_CONFIG_CLIENT_CERT_DATA=$(cat tools/deployment/certificates/ephemeral_config_client_cert_data| base64 -w0)
export EPHEMERAL_CONFIG_CLIENT_KEY_DATA=$(cat tools/deployment/certificates/ephemeral_config_client_key_data| base64 -w0)
export TARGET_IP=${TARGET_IP:-"10.23.25.102"}
export TARGET_CONFIG_CA_DATA=$(cat tools/deployment/certificates/target_config_ca_data| base64 -w0)
export TARGET_CONFIG_CLIENT_CERT_DATA=$(cat tools/deployment/certificates/target_config_client_cert_data| base64 -w0)
export TARGET_CONFIG_CLIENT_KEY_DATA=$(cat tools/deployment/certificates/target_config_client_key_data| base64 -w0)

# Remove the contents of the .airship folder, preserving the kustomize plugin directory
rm -rf $HOME/.airship/*config*
mkdir -p $HOME/.airship

echo "Generate ~/.airship/config and ~/.airship/kubeconfig"
envsubst <"${AIRSHIPCTL_WS}/tools/deployment/templates/airshipconfig_template" > ~/.airship/config
envsubst <"${AIRSHIPCTL_WS}/tools/deployment/templates/kubeconfig_template" > ~/.airship/kubeconfig
