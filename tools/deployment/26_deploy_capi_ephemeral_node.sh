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
export PROVIDER=${PROVIDER:-"metal3"}
export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export KUBECONFIG_EPHEMERAL_CONTEXT=${KUBECONFIG_EPHEMERAL_CONTEXT:-"ephemeral-cluster"}


if [ "$PROVIDER" = "metal3" ]; then

    echo "Deploy calico using tigera operator"
    airshipctl phase run initinfra-networking-ephemeral --debug

    echo "Wait for Calico to be deployed using tigera"
    kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT wait --all-namespaces --for=condition=Ready pods --all --timeout=1000s

    echo "Wait for Established condition of tigerastatus(CRD) to be true for tigerastatus(CR) to show up"
    kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT wait --for=condition=Established crd/tigerastatuses.operator.tigera.io --timeout=300s

    # Wait till CR(tigerastatus) shows up to query
    count=0
    max_retry_attempts=150
    until [[ $(kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT get tigerastatus 2>/dev/null) ]]; do
      count=$((count + 1))
      if [[ ${count} -eq "${max_retry_attempts}" ]]; then
        echo ' Timed out waiting for tigerastatus'
        exit 1
      fi
      sleep 2
    done

    # Wait till condition is available for tigerastatus
    kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT wait --for=condition=Available tigerastatus --all --timeout=1000s

    echo "Deploy metal3.io components to ephemeral node"
    airshipctl phase run initinfra-ephemeral --debug

    echo "Getting metal3 pods as debug information"
    kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT --namespace ${PROVIDER} get pods

    echo "Deploy cluster-api components to ephemeral node"
    airshipctl phase run clusterctl-init-ephemeral --debug
else
    echo "Deploy cluster-api components to ephemeral node"
    airshipctl phase run clusterctl-init-ephemeral --debug
    kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_EPHEMERAL_CONTEXT get pods -A
fi

echo "Waiting for clusterapi pods to come up"
kubectl --kubeconfig $KUBECONFIG  --context $KUBECONFIG_EPHEMERAL_CONTEXT wait --for=condition=available deploy --all --timeout=1000s -A
kubectl --kubeconfig $KUBECONFIG  --context $KUBECONFIG_EPHEMERAL_CONTEXT get pods --all-namespaces
