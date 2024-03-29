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

: ${KIND_VERSION:="v0.11.1"}
export KIND_URL="https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-$(uname)-amd64"
TMP=$(KIND_URL=${KIND_URL} tools/deployment/kind/get_kind.sh)
export KIND=${TMP}/kind
sudo cp ${KIND} /usr/local/bin/
${KIND} version
