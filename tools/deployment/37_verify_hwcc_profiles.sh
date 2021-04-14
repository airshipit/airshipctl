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

set -x

export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-cluster"}
declare -A PROFILES

PROFILES[hardwareclassification-profile1]=1
PROFILES[hardwareclassification-profile2]=0

declare -a ERRORS

# HWCC need BMH in Ready state.
for i in "${!PROFILES[@]}"
do
    nodes=$(kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get bmh --no-headers=true -l hardwareclassification.metal3.io/$i 2>/dev/null | wc -l)

    if [ $nodes != ${PROFILES[$i]} ]
    then
        ERRORS+=($i)
    fi
done

if [ ${#ERRORS[@]} != 0 ]
then
    echo FAILURE error with ${ERRORS[@]}
    exit 1
else
    echo "SUCCESS"
fi
