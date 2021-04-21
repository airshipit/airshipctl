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

# Installs minikube and other dependencies required for the scripts to run


MINIKUBE_VERSION="latest"

install_pkg(){
  for i in "$@"; do
    dpkg -l "$i" 2> /dev/null | grep ^ii > /dev/null || sudo DEBIAN_FRONTEND=noninteractive -E apt -y install "$i"
  done
}

# Grab usefull packages needed for minikube and other scripts
install_pkg curl conntrack make docker.io jq

curl -Lo minikube "https://storage.googleapis.com/minikube/releases/$MINIKUBE_VERSION/minikube-linux-amd64" \
  && chmod +x minikube

sudo mkdir -p /usr/local/bin/
sudo install minikube /usr/local/bin/