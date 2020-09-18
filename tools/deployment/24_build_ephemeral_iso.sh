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

ISO_DIR=${ISO_DIR:-"/srv/iso"}
CLEANUP_SERVE_DIR=${CLEANUP_SERVE_DIR:-"false"}
SITE_NAME=${SITE_NAME:-test-site}

#Create serving directories and assign permission and ownership
sudo rm -rf ${ISO_DIR}
sudo mkdir -p ${ISO_DIR}
sudo chmod -R 755 ${ISO_DIR}
sudo chown -R ${USER_NAME} ${ISO_DIR}

echo "Build ephemeral iso"
airshipctl phase run bootstrap --debug

echo "List generated iso"
ls -lth ${ISO_DIR}

echo "Remove the container used for iso generation"
sudo docker rm $(docker ps -a -f status=exited -q)

#cleanup the directories
if [ "${CLEANUP_SERVE_DIR}" == "true" ] || [ "${CLEANUP_SERVE_DIR}" == "True" ]; then
  echo "Clean directories used by ephemeral iso build"
  sudo rm -rf ${ISO_DIR}
fi
