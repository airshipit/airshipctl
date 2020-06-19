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

: ${KUBE_VERSION:="v1.17.4"}

# Install kubectl
URL="https://storage.googleapis.com"
sudo -E curl -sSLo /usr/local/bin/kubectl \
  "${URL}"/kubernetes-release/release/"${KUBE_VERSION}"/bin/linux/amd64/kubectl

sudo -E chmod +x /usr/local/bin/kubectl
