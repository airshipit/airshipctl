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

# Wait until airshipctl and libvirt infrastructure has been built
/wait_for airshipctl-builder
/wait_for infra-builder

export USER=root
# https://github.com/sudo-project/sudo/issues/42
echo "Set disable_coredump false" >> /etc/sudo.conf

echo "Installing kustomize"
kustomize_version=v3.8.5
kustomize_download_url="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/${kustomize_version}/kustomize_${kustomize_version}_linux_amd64.tar.gz"
curl -sSL "$kustomize_download_url" | tar -C /tmp -xzf -
install /tmp/kustomize /usr/local/bin

cp "$ARTIFACTS_DIR/airshipctl/bin/airshipctl" /usr/local/bin/airshipctl
cp -r "$ARTIFACTS_DIR/airshipctl/" /opt/airshipctl
cd /opt/airshipctl


curl -fsSL -o key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
SOPS_IMPORT_PGP="$(cat key.asc)"
SOPS_PGP_FP="FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4"
export SOPS_IMPORT_PGP SOPS_PGP_FP

# By default, don't build airshipctl - use the binary from the shared volume instead
# ./tools/deployment/21_systemwide_executable.sh
./tools/deployment/22_test_configs.sh
./tools/deployment/23_pull_documents.sh
./tools/deployment/23_generate_secrets.sh

sed -i -e 's#bmcAddress: redfish+http://\([0-9]\+\.[0-9]\+\.[0-9]\+\.[0-9]\+\):8000#bmcAddress: redfish+https://10.23.25.1:8443#' /tmp/airship/airshipctl/manifests/site/test-site/target/catalogues/hosts.yaml
sed -i -e 's#root#username#' /tmp/airship/airshipctl/manifests/site/test-site/target/catalogues/hosts.yaml
sed -i -e 's#r00tme#password#' /tmp/airship/airshipctl/manifests/site/test-site/target/catalogues/hosts.yaml
sed -i -e 's#disableCertificateVerification: false#disableCertificateVerification: true#' /tmp/airship/airshipctl/manifests/site/test-site/target/catalogues/hosts.yaml

if [[ "$USE_CACHED_ISO" = "true" ]]; then
  mkdir -p /srv/images
  tar -xzf "$CACHE_DIR/iso.tar.gz" --directory /srv/images
else
  ./tools/deployment/24_build_images.sh
  tar -czf "$ARTIFACTS_DIR/iso.tar.gz" --directory=/srv/images .
fi

./tools/deployment/25_deploy_ephemeral_node.sh
./tools/deployment/26_deploy_capi_ephemeral_node.sh
./tools/deployment/30_deploy_controlplane.sh
./tools/deployment/31_deploy_initinfra_target_node.sh
./tools/deployment/32_cluster_init_target_node.sh
./tools/deployment/33_cluster_move_target_node.sh
./tools/deployment/34_deploy_worker_node.sh
./tools/deployment/35_deploy_workload.sh
./tools/deployment/36_verify_hwcc_profiles.sh

/signal_complete runner
