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
export PROXY=${PROXY:-${http_proxy}}

set +e
echo "Build airshipctl docker images"
for i in {1..3}; do
    sudo -E make images && break
done
[ "$?" -ne 0 ] && exit 1
set -e

echo "Copy airshipctl from docker image"
DOCKER_IMAGE_TAG=$(sudo -E make print-docker-image-tag)
CONTAINER=$(sudo -E docker create "${DOCKER_IMAGE_TAG}")
sudo -E docker cp "${CONTAINER}:/usr/local/bin/airshipctl" "/usr/local/bin/airshipctl"
sudo -E docker rm "${CONTAINER}"

if ! airshipctl version | grep -q 'airshipctl'; then
  echo "Unable to verify airshipctl command. Please verify if the airshipctl is installed in /usr/local/bin/"
else
  echo "Airshipctl version"
  airshipctl version
fi

# Outside of releases, the airshipctl/treasuremap manifests reference krm functions via
# local-only image tags, specifically `localhost/<function name>`, so that we can
# set them externally in a single place (below parameters/logic), rather than maintaining
# explicit versions directly in the manifests. By default, these parameters
# reference the krm functions built above via `make images`, so that treasuremap
# and other downstream consumers can easily use the krm function versions matching
# the version of airshipctl that they are installing via this script.
export AIRSHIP_KRM_FUNCTION_REPO=${AIRSHIP_KRM_FUNCTION_REPO:-"quay.io/airshipit"}
export AIRSHIP_KRM_FUNCTION_TAG=${AIRSHIP_KRM_FUNCTION_TAG:-"latest"}
export SOPS_KRM_FUNCTION=${SOPS_KRM_FUNCTION:-"gcr.io/kpt-fn-contrib/sops:v0.3.0"}

echo "Resolve krm function versions"

set_krm_function () {
  if [[ "$(docker images -q "$2" 2> /dev/null)" == "" ]]; then
    docker pull "$2"
  fi
  docker tag "$2" "localhost/$1"
}

for FUNC in $(cd krm-functions; echo */ | tr -d /)
do
  IMG="${AIRSHIP_KRM_FUNCTION_REPO}/${FUNC}:${AIRSHIP_KRM_FUNCTION_TAG}"
  set_krm_function "$FUNC" "$IMG"
done

set_krm_function "sops" "$SOPS_KRM_FUNCTION"
