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

echo "Installing kustomize"
kustomize_version=v3.8.5
kustomize_download_url="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/${kustomize_version}/kustomize_${kustomize_version}_linux_amd64.tar.gz"
curl -sSL "$kustomize_download_url" | tar -C /tmp -xzf -
sudo install /tmp/kustomize /usr/local/bin
