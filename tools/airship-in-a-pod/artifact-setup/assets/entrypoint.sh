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

function cloneRepo(){
  repo_dir=$1
  repo_url=$2
  repo_ref=$3

  mkdir -p "$repo_dir"
  cd "$repo_dir"

  git init
  git fetch "$repo_url" "$repo_ref"
  git checkout FETCH_HEAD
}

if [[ "$USE_CACHED_ARTIFACTS" = "true" ]]
then
  printf "Using cached airshipctl\n"
  cp -r "$CACHE_DIR/*" "$ARTIFACTS_DIR"
else
  printf "Waiting 30 seconds for the libvirt and docker services to be ready\n"
  sleep 30

  repo_dir="$ARTIFACTS_DIR/airshipctl"
  cloneRepo "$repo_dir" "$AIRSHIPCTL_REPO_URL" "$AIRSHIPCTL_REPO_REF"

  cd "$repo_dir"
  ./tools/deployment/21_systemwide_executable.sh
  mkdir -p bin
  cp "$(command -v airshipctl)" bin
fi

/signal_complete artifact-setup
