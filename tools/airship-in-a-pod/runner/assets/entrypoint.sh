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

# Create the "canary" file, indicating that the container is healthy
mkdir -p /tmp/healthy
touch /tmp/healthy/runner

success=false
function cleanup() {
  if [[ "$success" == "false" ]]; then
    rm /tmp/healthy/runner
  fi
}
trap cleanup EXIT

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

install "$ARTIFACTS_DIR/airshipctl/bin/airshipctl" /usr/local/bin
cd "$ARTIFACTS_DIR/airshipctl"

export AIRSHIP_CONFIG_MANIFEST_DIRECTORY="$ARTIFACTS_DIR/manifests"
./tools/deployment/22_test_configs.sh
if [[ -n "$AIRSHIP_CONFIG_PHASE_REPO_REF" ]]; then
  export NO_CHECKOUT="false"
else
  export NO_CHECKOUT="true"
fi
./tools/deployment/23_pull_documents.sh
./tools/deployment/23_generate_secrets.sh

echo "export KUBECONFIG=$HOME/.airship/kubeconfig" >> ~/.bashrc

repo_name=$(yq -r .manifests.dummy_manifest.repositories.primary.url /root/.airship/config | awk 'BEGIN {FS="/"} {print $NF}' | cut -d'.' -f1)
sed -i -e 's#bmcAddress: redfish+http://\([0-9]\+\.[0-9]\+\.[0-9]\+\.[0-9]\+\):8000#bmcAddress: redfish+https://10.23.25.1:8443#' "$AIRSHIP_CONFIG_MANIFEST_DIRECTORY/$repo_name/manifests/site/test-site/target/catalogues/hosts.yaml"
sed -i -e 's#root#username#' "$AIRSHIP_CONFIG_MANIFEST_DIRECTORY/$repo_name/manifests/site/test-site/target/catalogues/hosts.yaml"
sed -i -e 's#r00tme#password#' "$AIRSHIP_CONFIG_MANIFEST_DIRECTORY/$repo_name/manifests/site/test-site/target/catalogues/hosts.yaml"
sed -i -e 's#disableCertificateVerification: false#disableCertificateVerification: true#' "$AIRSHIP_CONFIG_MANIFEST_DIRECTORY/$repo_name/manifests/site/test-site/target/catalogues/hosts.yaml"

if [[ "$USE_CACHED_ISO" = "true" ]]; then
  mkdir -p /srv/images
  tar -xzf "$CACHE_DIR/iso.tar.gz" --directory /srv/images
else
  ./tools/deployment/24_build_images.sh
  tar -czf "$ARTIFACTS_DIR/iso.tar.gz" --directory=/srv/images .
fi

./tools/deployment/25_deploy_gating.sh

success=true
/signal_complete runner
