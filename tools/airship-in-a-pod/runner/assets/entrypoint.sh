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

set -ex

/signal_status "runner" "RUNNING"
success=false
function reportStatus() {
  if [[ "$success" == "false" ]]; then
    /signal_status "runner" "FAILED"
  else
    /signal_status "runner" "SUCCESS"
  fi
  # Keep the container running for debugging/monitoring purposes
  sleep infinity
}
trap reportStatus EXIT

# Wait until artifact-setup and libvirt infrastructure has been built
/wait_for artifact-setup
/wait_for infra-builder

export USER=root
# https://github.com/sudo-project/sudo/issues/42
echo "Set disable_coredump false" >> /etc/sudo.conf

echo "Installing kustomize"
kustomize_version=v3.8.5
kustomize_download_url="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/${kustomize_version}/kustomize_${kustomize_version}_linux_amd64.tar.gz"
curl -sSL "$kustomize_download_url" | tar -C /tmp -xzf -
install /tmp/kustomize /usr/local/bin

curl -fsSL -o /sops-key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
SOPS_PGP_FP="FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4"
SOPS_IMPORT_PGP="$(cat /sops-key.asc)"
export SOPS_IMPORT_PGP
export SOPS_PGP_FP
echo 'export SOPS_IMPORT_PGP="$(cat /sops-key.asc)"' >> ~/.bashrc
echo "export SOPS_PGP_FP=${SOPS_PGP_FP}" >> ~/.bashrc
echo "export KUBECONFIG=$HOME/.airship/kubeconfig" >> ~/.bashrc

install "$ARTIFACTS_DIR/airshipctl/bin/airshipctl" /usr/local/bin
cd "$ARTIFACTS_DIR/airshipctl"

set +x
if [[ "$AIRSHIP_CONFIG_MANIFEST_REPO_AUTH_TYPE" = "http-basic" ]]
then
  export AIRSHIP_CONFIG_MANIFEST_REPO_AUTH_USERNAME=$( cat /opt/aiap-secret-volume/AIRSHIP_CONFIG_MANIFEST_REPO_AUTH_USERNAME )
  export AIRSHIP_CONFIG_MANIFEST_REPO_AUTH_HTTP_PASSWORD=$( cat /opt/aiap-secret-volume/AIRSHIP_CONFIG_MANIFEST_REPO_AUTH_HTTP_PASSWORD )
fi

if [[ "$AIRSHIP_CONFIG_MANIFEST_REPO_AUTH_TYPE" = "ssh-pass" ]]
then
  export AIRSHIP_CONFIG_MANIFEST_REPO_AUTH_SSH_PASSWORD=$( cat /opt/aiap-secret-volume/AIRSHIP_CONFIG_MANIFEST_REPO_AUTH_SSH_PASSWORD )
fi
set -x

export AIRSHIP_CONFIG_MANIFEST_DIRECTORY="/opt/manifests"
./tools/deployment/22_test_configs.sh
if [[ -n "$AIRSHIP_CONFIG_PHASE_REPO_REF" || -n "$AIRSHIP_CONFIG_PHASE_REPO_BRANCH" ]]; then
  export NO_CHECKOUT="false"
else
  export NO_CHECKOUT="true"
fi
./tools/deployment/23_pull_documents.sh
./tools/deployment/23_generate_secrets.sh

repo_name=$(yq -r .manifests.dummy_manifest.repositories.primary.url /root/.airship/config | awk 'BEGIN {FS="/"} {print $NF}' | cut -d'.' -f1)
hosts_file="$AIRSHIP_CONFIG_MANIFEST_DIRECTORY/$repo_name/manifests/site/test-site/target/catalogues/shareable/hosts.yaml"
sed -i -e 's#bmcAddress: redfish+http://\([0-9]\+\.[0-9]\+\.[0-9]\+\.[0-9]\+\):8000#bmcAddress: redfish+https://10.23.25.1:8443#' "$hosts_file"
sed -i -e 's#root#username#' "$hosts_file"
sed -i -e 's#r00tme#password#' "$hosts_file"
sed -i -e 's#disableCertificateVerification: false#disableCertificateVerification: true#' "$hosts_file"

cp -r /opt/manifests "$ARTIFACTS_DIR/manifests"

if [[ "$USE_CACHED_ISO" = "true" ]]; then
  mkdir -p /srv/images
  tar -xzf "$CACHE_DIR/iso.tar.gz" --directory /srv/images
else
  ./tools/deployment/24_build_images.sh
  tar -czf "$ARTIFACTS_DIR/iso.tar.gz" --directory=/srv/images .
fi

./tools/deployment/25_deploy_gating.sh

success=true
