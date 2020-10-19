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

export USE_PROXY=${USE_PROXY:-"false"}
export HTTPS_PROXY=${HTTPS_PROXY:-${https_proxy}}
export HTTP_PROXY=${HTTP_PROXY:-${http_proxy}}
export NO_PROXY=${NO_PROXY:-${no_proxy}}

echo "Build airshipctl docker images"
make images

echo "Copy airshipctl from docker image"
DOCKER_IMAGE_TAG=$(make print-docker-image-tag)
CONTAINER=$(docker create "${DOCKER_IMAGE_TAG}")
sudo docker cp "${CONTAINER}:/usr/local/bin/airshipctl" "/usr/local/bin/airshipctl"
sudo docker rm "${CONTAINER}"

if ! airshipctl version | grep -q 'airshipctl'; then
  echo "Unable to verify airshipctl command. Please verify if the airshipctl is installed in /usr/local/bin/"
else
  echo "Airshipctl version"
  airshipctl version
fi

echo "Install airshipctl as kustomize plugins"
AIRSHIPCTL="/usr/local/bin/airshipctl" ./tools/document/build_kustomize_plugin.sh
