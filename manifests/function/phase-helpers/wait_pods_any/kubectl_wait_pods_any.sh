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

N=0
MAX_RETRY=30
DELAY=60
until [ "$N" -ge ${MAX_RETRY} ]
do
    if [ "$(kubectl --context $KCTL_CONTEXT \
              --request-timeout 10s \
              get pods \
              --all-namespaces -o name | wc -l)" -ge "1" ]; then
      kubectl --context $KCTL_CONTEXT --request-timeout 10s get pods --all-namespaces 1>&2
      break
  fi

  N=$((N+1))
  echo "$N: Retrying to get any pods" 1>&2
  sleep ${DELAY}
done

if [ "$N" -ge ${MAX_RETRY} ]; then
  echo "Could not get any pods" 1>&2
  exit 1
fi
