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
  repo_name=$1
  repo_url=$2
  repo_ref=$3

  repo_dir="$ARTIFACTS_DIR/$repo_name"
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

  cloneRepo $MANIFEST_REPO_NAME $MANIFEST_REPO_URL $MANIFEST_REPO_REF

  if [[ "$MANIFEST_REPO_NAME" !=  "airshipctl" ]]
  then
	  cloneRepo airshipctl https://github.com/airshipit/airshipctl $AIRSHIPCTL_REF
  fi
  cd $ARTIFACTS_DIR/$MANIFEST_REPO_NAME

  if [[ "$MANIFEST_REPO_NAME" ==  "airshipctl" ]]
  then
    ./tools/deployment/21_systemwide_executable.sh
  else
    ./tools/deployment/airship-core/21_systemwide_executable.sh
  fi
  mkdir -p bin
  cp "$(which airshipctl)" bin
fi

/signal_complete artifact-setup
