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

set -ex

EPHEMERAL_DOMAIN_NAME="air-ephemeral"

if type "virsh" > /dev/null; then
  for vm in $(virsh list --all --name --state-running |grep ${EPHEMERAL_DOMAIN_NAME})
  do
    echo "Stop ephemeral node '$vm'" 1>&2
    virsh destroy $vm 1>&2
  done
else
  echo "Can't find virsh" 1>&2
fi
