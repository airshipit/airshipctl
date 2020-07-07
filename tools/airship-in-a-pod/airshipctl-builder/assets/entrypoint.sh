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

if [[ "$USE_CACHED_AIRSHIPCTL" = "true" ]]
then
  printf "Using cached airshipctl\n"
  cp -r "$CACHE_DIR/airshipctl" "$ARTIFACTS_DIR/airshipctl"
else
  printf "Waiting 30 seconds for the libvirt, sushy, and docker services to be ready\n"
  sleep 30

  airshipctl_dir="$ARTIFACTS_DIR/airshipctl"
  mkdir -p "$airshipctl_dir"
  cd "$airshipctl_dir"

  git init
  git fetch "$AIRSHIPCTL_REPO" "$AIRSHIPCTL_REF"
  git checkout FETCH_HEAD

  ./tools/deployment/21_systemwide_executable.sh
  mkdir -p bin
  cp "$(which airshipctl)" bin
fi

/signal_complete airshipctl-builder
