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

set -ex

TARGET_IMAGE_DIR="/srv/iso"
EPHEMERAL_DOMAIN_NAME="air-ephemeral"
TARGET_IMAGE_URL="https://cloud-images.ubuntu.com/focal/current/focal-server-cloudimg-amd64.img"
export WAIT_TIMEOUT=${WAIT_TIMEOUT:-"2000s"}
export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-context"}

# TODO (dukov) this is needed due to sushy tools inserts cdrom image to
# all vms. This can be removed once sushy tool is fixed
echo "Ensure all cdrom images are ejected."
for vm in $(sudo virsh list --all --name |grep -v ${EPHEMERAL_DOMAIN_NAME})
do
  sudo virsh domblklist $vm |
    awk 'NF==2 {print $1}' |
    grep -v Target |
    xargs -I{} sudo virsh change-media $vm {} --eject || :
done

echo "Download target image"
DOWNLOAD="200"
if [ -e ${TARGET_IMAGE_DIR}/target-image.qcow2 ]
then
    MTIME=$(date -d @$(stat -c %Y ${TARGET_IMAGE_DIR}/target-image.qcow2) +"%a, %d %b %Y %T %Z")
    DOWNLOAD=$(curl -sSLI \
        --write-out '%{http_code}' \
        -H "If-Modified-Since: ${MTIME}" \
        ${TARGET_IMAGE_URL} | tail -1)
fi
if [ "${DOWNLOAD}" != "304" ]
then
    curl -sSLo ${TARGET_IMAGE_DIR}/target-image.qcow2 ${TARGET_IMAGE_URL}
fi
md5sum /srv/iso/target-image.qcow2 | cut -d ' ' -f 1 > ${TARGET_IMAGE_DIR}/target-image.qcow2.md5sum

echo "Create target k8s cluster resources"
airshipctl phase apply controlplane --wait-timeout $WAIT_TIMEOUT --debug

echo "Switch context to target cluster and set manifest"
airshipctl config use-context target-context
airshipctl config set-context target-context --manifest dummy_manifest

echo "Wait for apiserver to become available"
N=0
MAX_RETRY=30
DELAY=60
until [ "$N" -ge ${MAX_RETRY} ]
do
  if timeout 20 kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get node; then
      break
  fi

  N=$((N+1))
  echo "$N: Retrying to reach the apiserver"
  sleep ${DELAY}
done

if [ "$N" -ge ${MAX_RETRY} ]; then
  echo "Could not reach the apiserver"
  exit 1
fi

echo "List all pods"
kubectl --kubeconfig  $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get pods --all-namespaces
