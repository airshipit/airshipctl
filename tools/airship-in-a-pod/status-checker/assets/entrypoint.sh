#!/bin/sh

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

set -e

status_dir="/tmp/status"
mkdir -p "$status_dir"

while true; do
  artifact_setup_status="UNKNOWN"
  infra_builder_status="UNKNOWN"
  runner_status="UNKNOWN"
  if [ -f "$status_dir/artifact-setup" ]; then
    artifact_setup_status="$(cat $status_dir/artifact-setup)"
  fi
  if [ -f "$status_dir/infra-builder" ]; then
    infra_builder_status="$(cat $status_dir/infra-builder)"
  fi
  if [ -f "$status_dir/runner" ]; then
    runner_status="$(cat $status_dir/runner)"
  fi

  # Print all statuses on a single line
  printf "artifact-setup: <%s> " "$artifact_setup_status"
  printf "infra-builder: <%s> " "$infra_builder_status"
  printf "runner: <%s> " "$runner_status"
  printf "\n"

  sleep 5
done
