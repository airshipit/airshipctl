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

: ${SONOBUOY_VERSION:="0.18.2"}
: ${KUBECONFIG:="$HOME/.airship/kubeconfig"}
URL="https://github.com/vmware-tanzu/sonobuoy/releases/download/v${SONOBUOY_VERSION}/sonobuoy_${SONOBUOY_VERSION}_linux_amd64.tar.gz"
rm -rf /tmp/sonobuoy
mkdir /tmp/sonobuoy
sudo -E curl -sSLo "/tmp/sonobuoy/sonobuoy_${SONOBUOY_VERSION}_linux_amd64.tar.gz" ${URL}
tar xvf /tmp/sonobuoy/sonobuoy_${SONOBUOY_VERSION}_linux_amd64.tar.gz -C /tmp/sonobuoy/
sudo install -m 755 -o root /tmp/sonobuoy/sonobuoy /usr/local/bin
echo ${KUBECONFIG}
sonobuoy version --kubeconfig ${KUBECONFIG}
