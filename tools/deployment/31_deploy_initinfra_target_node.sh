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

echo "Deploy calico using tigera operator"
airshipctl phase run initinfra-networking-target --debug

echo "Wait for Calico to be deployed using tigera"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait --all-namespaces --for=condition=Ready pods --all --timeout=600s

echo "Wait for Established condition of tigerastatus(CRD) to be true for tigerastatus(CR) to show up"
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait --for=condition=Established crd/tigerastatuses.operator.tigera.io --timeout=300s

# Wait till CR(tigerastatus) is available
count=0
max_retry_attempts=150
until [[ $(kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT get tigerastatus 2>/dev/null) ]]; do
  count=$((count + 1))
  if [[ ${count} -eq "${max_retry_attempts}" ]]; then
    echo ' Timed out waiting for tigerastatus'
    exit 1
  fi
  sleep 2
done

# Wait till condition is available for tigerastatus
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT wait --for=condition=Available tigerastatus --all --timeout=1000s

echo "Deploy infra to cluster"
airshipctl phase run initinfra-target --debug

echo "List all pods"
kubectl \
  --kubeconfig $KUBECONFIG \
  --context $KUBECONFIG_TARGET_CONTEXT \
   get pods \
  --all-namespaces
