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

echo "Deploy Target Workload Cluster"
airshipctl phase apply azure

echo "Get kubeconfig from secret"
KUBECONFIG=""
N=0
MAX_RETRY=30
DELAY=60
until [ "$N" -ge ${MAX_RETRY} ]
do
  KUBECONFIG=$(kubectl --kubeconfig ~/.airship/kubeconfig --namespace=default get secret/az-workload-cluster-kubeconfig -o jsonpath={.data.value}  || true)

  if [[ ! -z "$KUBECONFIG" ]]; then
      break
  fi

  N=$((N+1))
  echo "$N: Retry to get target cluster kubeconfig from secret."
  sleep ${DELAY}
done

if [[ -z "$KUBECONFIG" ]]; then
  echo "Could not get target cluster kubeconfig from sceret."
  exit 1
fi

echo "Create kubeconfig"
echo ${KUBECONFIG} | base64 -d > /tmp/target.kubeconfig

echo "Get Machine State"
kubectl get machines

echo "Check kubectl version"
VERSION=""
N=0
MAX_RETRY=30
DELAY=60
until [ "$N" -ge ${MAX_RETRY} ]
do
  VERSION=$(timeout 20 kubectl --kubeconfig /tmp/target.kubeconfig version | grep 'Server Version' || true)

  if [[ ! -z "$VERSION" ]]; then
      break
  fi

  N=$((N+1))
  echo "$N: Retry to get kubectl version."
  sleep ${DELAY}
done

if [[ -z "$VERSION" ]]; then
  echo "Could not get kubectl version."
  exit 1
fi

echo "Check nodes status"

kubectl --kubeconfig /tmp/target.kubeconfig wait --for=condition=Ready node --all --timeout 900s
kubectl get nodes --kubeconfig /tmp/target.kubeconfig

echo "Get cluster state"
kubectl --kubeconfig ${HOME}/.airship/kubeconfig get cluster

