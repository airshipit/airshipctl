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

export TIMEOUT=${TIMEOUT:-3600}
WORKER_NODE=$(kubectl --context $KCTL_CONTEXT \
                get -f $RENDERED_BUNDLE_PATH \
                --output jsonpath='{..metadata.name}')

echo "Wait $TIMEOUT seconds for BMH to be in ready state." 1>&2
end=$(($(date +%s) + $TIMEOUT))
for worker in $WORKER_NODE
do
  while true; do
    if [ "$(kubectl --request-timeout 20s \
              --context $KCTL_CONTEXT \
              get bmh $worker \
              -o jsonpath='{.status.provisioning.state}')" == "provisioned" ] ; then

      echo "Get BMHs status" 1>&2
      kubectl \
        --context $KCTL_CONTEXT \
        get bmh 1>&2
      break
    else
      now=$(date +%s)
      if [ $now -gt $end ]; then
        echo "BMH is not ready before TIMEOUT=$TIMEOUT" 1>&2
        exit 1
      fi
      sleep 15
    fi
  done
done
