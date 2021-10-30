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

# Installs kind and other dependencies required for the scripts to run


KIND_VERSION="v0.11.1"

install_pkg(){
  for i in "$@"; do
    dpkg -l "$i" 2> /dev/null | grep ^ii > /dev/null || sudo DEBIAN_FRONTEND=noninteractive -E apt -y install "$i"
  done
}

# Grab usefull packages needed for kind and other scripts
install_pkg curl conntrack make jq

curl -Lo ./kind "https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-amd64" \
  && chmod +x ./kind

sudo mkdir -p /usr/local/bin/
sudo mv ./kind /usr/local/bin/
