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
export CONDITION=${CONDITION:-"controlPlaneReady"}

end=$(($(date +%s) + $TIMEOUT))
echo "Waiting $TIMEOUT seconds for cluster to reach $CONDITION condition" 1>&2
while true; do
    # TODO(vkuzmin): Add ability to wait for multiple clusters
    if [ "$(kubectl \
              --request-timeout 20s \
              --context $KCTL_CONTEXT \
              get -f $RENDERED_BUNDLE_PATH \
              -o jsonpath={.status.$CONDITION})" == "true" ]
    then
        echo "Getting information about cluster" 1>&2
        kubectl \
          --request-timeout 20s \
          --context $KCTL_CONTEXT \
          get -f $RENDERED_BUNDLE_PATH 1>&2
        echo "Cluster successfully reach $CONDITION condition" 1>&2
        break
    else
        now=$(date +%s)
        if [ $now -gt $end ]; then
            echo "Cluster didn't reach $CONDITION condition before TIMEOUT=$TIMEOUT, exiting" 1>&2
            exit 1
        fi
        sleep 15
    fi
done
