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

set -xe

export TIMEOUT=${TIMEOUT:-1000}

echo "Wait $TIMEOUT seconds for tigera status to be in Available state." 1>&2
end=$(($(date +%s) + $TIMEOUT))

until [ "$(kubectl --kubeconfig $KUBECONFIG --context $KCTL_CONTEXT wait --for=condition=Available --all tigerastatus 2>/dev/null)" ]; do
  now=$(date +%s)
  if [ $now -gt $end ]; then
    echo "Tigera status is not ready before TIMEOUT=$TIMEOUT" 1>&2
    exit 1
  fi
  sleep 10
done
