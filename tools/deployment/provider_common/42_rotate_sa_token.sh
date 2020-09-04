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

export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-cluster"}

echo "Getting pod Name in CAPD namespace"
podName=$(kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT --namespace capi-system get pods -o jsonpath='{.items[0].metadata.name}')

secretName=$(kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get secrets -n capi-system | grep 'kubernetes.io/service-account-token' | grep default |awk '{print $1}' | head -1)
echo "Checking airshipctl cluster rotate-sa-token in capd-system namespace"
airshipctl --kubeconfig $KUBECONFIG cluster rotate-sa-token --secret-namespace capi-system --secret-name $secretName

sleep 5
echo "Checking for ready state of all pods"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait -n capi-system --for=condition=Available deploy --all --timeout=600s
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get pods -n capi-system

podNameNew=$(kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT --namespace capi-system get pods | grep -i running | awk '{print $1}')

if [[ -z $podNameNew || $podName == $podNameNew ]]; then
    echo "Rotation of SA token is unsuccessful"
    exit 1
fi

echo "Rotate SA token is successful"

echo "Testing in default namespace with Nginx pod"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT create deploy nginx-deployment --image=nginx

kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait --for=condition=Available deploy --all --timeout=600s

echo "Getting the current Nginx podname"
podName=$(kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get pods -o jsonpath='{.items[0].metadata.name}')

secretName=$(kubectl get secret --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT | grep 'kubernetes.io/service-account-token' | grep default |awk '{print $1}' | head -1)

airshipctl cluster rotate-sa-token --kubeconfig $KUBECONFIG --secret-namespace default --secret-name $secretName

sleep 5
echo "Checking for ready state of all pods"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait --for=condition=Available deploy --all --timeout=600s
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get pods

podNameNew=$(kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get pods | grep -i running | awk '{print $1}')

if [[ -z $podNameNew || $podName == $podNameNew ]]; then
    echo "Rotation of SA token is unsuccessful"
    kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT delete deploy nginx-deployment
    exit 1
fi
echo "Rotate SA token is successful"

kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT delete deploy nginx-deployment
