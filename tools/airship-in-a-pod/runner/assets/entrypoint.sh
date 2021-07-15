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

cp "$ARTIFACTS_DIR/$MANIFEST_REPO_NAME/bin/airshipctl" /usr/local/bin/airshipctl
if [ $MANIFEST_REPO_NAME != "airshipctl" ]
then
  export AIRSHIP_CONFIG_PHASE_REPO_URL="https://opendev.org/airship/treasuremap"
  cp -r $ARTIFACTS_DIR/airshipctl/ /opt/airshipctl
fi

cp  -r $ARTIFACTS_DIR/$MANIFEST_REPO_NAME/ /opt/$MANIFEST_REPO_NAME
cd /opt/$MANIFEST_REPO_NAME

curl -fsSL -o /sops-key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
SOPS_PGP_FP="FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4"
SOPS_IMPORT_PGP="$(cat /sops-key.asc)"
export SOPS_IMPORT_PGP
export SOPS_PGP_FP
echo 'export SOPS_IMPORT_PGP="$(cat /sops-key.asc)"' >> ~/.bashrc
echo "export SOPS_PGP_FP=${SOPS_PGP_FP}" >> ~/.bashrc

export AIRSHIP_CONFIG_MANIFEST_DIRECTORY="/tmp/airship"

# By default, don't build airshipctl - use the binary from the shared volume instead
# ./tools/deployment/21_systemwide_executable.sh
if [ "$MANIFEST_REPO_NAME" == "airshipctl" ]
then
  ./tools/deployment/22_test_configs.sh
  # `airshipctl document pull` doesn't support pull patchsets yet
  #./tools/deployment/23_pull_documents.sh
  mkdir /tmp/airship
  cp -rp /opt/airshipctl /tmp/airship/airshipctl
  ./tools/deployment/23_generate_secrets.sh
else
  ./tools/deployment/airship-core/22_test_configs.sh
  ./tools/deployment/airship-core/23_pull_documents.sh
  ./tools/deployment/airship-core/23_generate_secrets.sh

fi

sed -i -e 's#bmcAddress: redfish+http://\([0-9]\+\.[0-9]\+\.[0-9]\+\.[0-9]\+\):8000#bmcAddress: redfish+https://10.23.25.1:8443#' "/tmp/airship/$MANIFEST_REPO_NAME/manifests/site/test-site/target/catalogues/hosts.yaml"
sed -i -e 's#root#username#' "/tmp/airship/$MANIFEST_REPO_NAME/manifests/site/test-site/target/catalogues/hosts.yaml"
sed -i -e 's#r00tme#password#' "/tmp/airship/$MANIFEST_REPO_NAME/manifests/site/test-site/target/catalogues/hosts.yaml"
sed -i -e 's#disableCertificateVerification: false#disableCertificateVerification: true#' "/tmp/airship/$MANIFEST_REPO_NAME/manifests/site/test-site/target/catalogues/hosts.yaml"

if [[ "$USE_CACHED_ISO" = "true" ]]; then
  mkdir -p /srv/images
  tar -xzf "$CACHE_DIR/iso.tar.gz" --directory /srv/images
else
  if [ "$MANIFEST_REPO_NAME" == "airshipctl" ]
  then
    ./tools/deployment/24_build_images.sh
  else
    ./tools/deployment/airship-core/24_build_images.sh
  fi

  tar -czf "$ARTIFACTS_DIR/iso.tar.gz" --directory=/srv/images .
fi

if [ "$MANIFEST_REPO_NAME" == "airshipctl" ]
then
  ./tools/deployment/25_deploy_gating.sh
else
  ./tools/deployment/airship-core/25_deploy_gating.sh
fi

/signal_complete runner
