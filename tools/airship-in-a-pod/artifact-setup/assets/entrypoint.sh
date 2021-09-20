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

/signal_status "artifact-setup" "RUNNING"
success=false
function reportStatus() {
  if [[ "$success" == "false" ]]; then
    /signal_status "artifact-setup" "FAILED"
  else
    /signal_status "artifact-setup" "SUCCESS"
  fi
  # Keep the container running for debugging/monitoring purposes
  sleep infinity
}
trap reportStatus EXIT

function cloneRepo() {
  repo_dir=$1
  repo_url=$2
  repo_ref=$3

  mkdir -p "$repo_dir"
  cd "$repo_dir"

  git init
  git fetch "$repo_url" "$repo_ref"
  git checkout FETCH_HEAD
}

function check_docker_readiness() {
  timeout=300

  #add wait condition
  end=$(($(date +%s) + $timeout))
  echo "Waiting $timeout seconds for docker to be ready."
  while true; do
    if ( docker version | grep 'Version' ); then
      echo "docker is now ready"
      break
    else
      echo "docker is not ready"
    fi
    now=$(date +%s)
    if [ $now -gt $end ]; then
      echo -e "\n Docker failed to become ready within a reasonable timeframe."
      exit 1
    fi
    sleep 10
  done
}

if [[ "$USE_CACHED_AIRSHIPCTL" = "true" ]]
then
  printf "Using cached airshipctl\n"
  cp -r "$CACHE_DIR/*" "$ARTIFACTS_DIR"
else
  check_docker_readiness

  repo_dir="$ARTIFACTS_DIR/airshipctl"
  cloneRepo "$repo_dir" "$AIRSHIPCTL_REPO_URL" "$AIRSHIPCTL_REPO_REF"

  cd "$repo_dir"
  ./tools/deployment/21_systemwide_executable.sh
  mkdir -p bin
  cp "$(command -v airshipctl)" bin
fi

success=true
/signal_status artifact-setup
