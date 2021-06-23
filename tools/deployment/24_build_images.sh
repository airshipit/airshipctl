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

export USER_NAME=${USER:-"ubuntu"}

IMAGE_DIR=${IMAGE_DIR:-"/srv/images"}
CLEANUP_SERVE_DIR=${CLEANUP_SERVE_DIR:-"false"}
SITE=${SITE:-test-site}
# List of phases to run to build images.
IMAGE_PHASE_PLANS=${IMAGE_PHASE_PLANS:-"iso"}

#Create serving directories and assign permission and ownership
sudo rm -rf ${IMAGE_DIR}
sudo mkdir -p ${IMAGE_DIR}
sudo chmod -R 755 ${IMAGE_DIR}
sudo chown -R ${USER_NAME} ${IMAGE_DIR}

unset IFS
for plan in $IMAGE_PHASE_PLANS; do
  echo "Build phase plan: $plan"
  airshipctl plan run $plan --debug
done

echo "List generated images"
ls -lth ${IMAGE_DIR}

#cleanup the directories
if [ "${CLEANUP_SERVE_DIR}" == "true" ] || [ "${CLEANUP_SERVE_DIR}" == "True" ]; then
  echo "Clean directories used by image-builder"
  sudo rm -rf ${IMAGE_DIR}
fi
