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
export HTTPS_PROXY=${HTTP_PROXY:-${http_proxy}}
export NO_PROXY=${NO_PROXY:-${no_proxy}}
export AIRSHIP_CONFIG_ISO_GEN_TARGET_PATH=${ISO_DIR}
export AIRSHIP_CONFIG_ISO_BUILDER_DOCKER_IMAGE=${BUILDER_IMAGE:-"quay.io/airshipit/isogen:latest-debian_stable"}
export REMOTE_TYPE=redfish
export REMOTE_INSECURE=true
export REMOTE_PROXY=false
export AIRSHIP_CONFIG_ISO_SERVE_HOST=${HOST:-"localhost"}
export AIRSHIP_CONFIG_ISO_PORT=${SERVE_PORT}
export AIRSHIP_CONFIG_ISO_NAME=${ISO_NAME:-"debian-custom.iso"}
export SYSTEM_ACTION_RETRIES=30
export SYSTEM_REBOOT_DELAY=30
export AIRSHIP_CONFIG_PRIMARY_REPO_BRANCH=${BRANCH:-"master"}
# the git repo url or local file system path to a cloned repo, e.g., /home/stack/airshipctl
export AIRSHIP_CONFIG_PRIMARY_REPO_URL=${REPO:-"https://review.opendev.org/airship/airshipctl"}
export AIRSHIP_SITE_NAME="manifests/site/test-site"
export AIRSHIP_CONFIG_MANIFEST_DIRECTORY=${AIRSHIP_CONFIG_MANIFEST_DIRECTORY:-"/tmp/airship"}
export AIRSHIP_CONFIG_CA_DATA=$(cat tools/deployment/certificates/airship_config_ca_data| base64 -w0)
export AIRSHIP_CONFIG_EPHEMERAL_IP=${IP_Ephemeral:-"10.23.25.101"}
export AIRSHIP_CONFIG_CLIENT_CERT_DATA=$(cat tools/deployment/certificates/airship_config_client_cert_data| base64 -w0)
export AIRSHIP_CONFIG_CLIENT_KEY_DATA=$(cat tools/deployment/certificates/airship_config_client_key_data| base64 -w0)

#Remove and Create .airship folder
rm -rf $HOME/.airship
mkdir -p $HOME/.airship

echo "Generate ~/.airship/config and ~/.airship/kubeconfig"
envsubst <"${AIRSHIPCTL_WS}/tools/deployment/templates/airshipconfig_template" > ~/.airship/config
envsubst <"${AIRSHIPCTL_WS}/tools/deployment/templates/kubeconfig_template" > ~/.airship/kubeconfig

if ! airshipctl config get-cluster | grep -q 'dummycluster_ephemeral' ; then
  echo "Unable to verify the ephemeral cluster details. Please verify the ephemeral cluster configuration."
else
  echo "Verify airshipctl configuration"
  airshipctl config get-cluster
fi
