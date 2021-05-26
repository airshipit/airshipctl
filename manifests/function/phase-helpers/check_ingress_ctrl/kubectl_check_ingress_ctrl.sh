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

export TARGET_IP=$(kubectl --context $KCTL_CONTEXT \
                     --namespace ingress \
                     get pods \
                     -l app.kubernetes.io/component=controller \
                     -o jsonpath='{.items[*].status.hostIP}')


echo "Ensure we can reach ingress controller default backend" 1>&2
if [ "404" != "$(curl --head \
                   --write-out '%{http_code}' \
                   --silent \
                   --output /dev/null \
                   $TARGET_IP/should-404)" ]; then
    echo "Failed to reach ingress controller default backend." 1>&2

    kubectl --context $KCTL_CONTEXT get all -n flux-system 1>&2
    kubectl --context $KCTL_CONTEXT logs -n flux-system -l app=helm-controller 1>&2
    kubectl --context $KCTL_CONTEXT get hr --all-namespaces -o yaml 1>&2

    exit 1
fi
