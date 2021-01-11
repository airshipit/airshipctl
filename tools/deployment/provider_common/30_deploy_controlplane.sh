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

# Example Usage
# CONTROLPLANE_COUNT=1 \
# SITE=docker-test-site \
# ./tools/deployment/provider_common/30_deploy_controlplane.sh

export AIRSHIP_SRC=${AIRSHIP_SRC:-"/tmp/airship"}
export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export CONTROLPLANE_COUNT=${CONTROLPLANE_COUNT:-"1"}
export SITE=${SITE:-"docker-test-site"}
export TARGET_CLUSTER_NAME=${TARGET_CLUSTER_NAME:-"target-cluster"}

# Adjust Control Plane Count (default 1)
# No. of control plane can be changed using
# CONTROLPLANE_COUNT=<replicas> tools/deployment/docker/30_deploy_controlplane.sh

sed -i "/value.*/s//value\": $CONTROLPLANE_COUNT }/g" \
${AIRSHIP_SRC}/airshipctl/manifests/site/${SITE}/ephemeral/controlplane/machine_count.json

echo "create control plane"
airshipctl phase run controlplane-ephemeral --debug --kubeconfig ${KUBECONFIG} --wait-timeout 1000s

TARGET_KUBECONFIG=""
TARGET_KUBECONFIG=$(kubectl --kubeconfig "${KUBECONFIG}" --namespace=default get secret/"${TARGET_CLUSTER_NAME}"-kubeconfig -o jsonpath={.data.value}  || true)

if [[ -z "$TARGET_KUBECONFIG" ]]; then
  echo "Error: Could not get kubeconfig from secret."
  exit 1
fi

echo "Generate kubeconfig"
echo ${TARGET_KUBECONFIG} | base64 -d > /tmp/${TARGET_CLUSTER_NAME}.kubeconfig
echo "Generate kubeconfig: /tmp/${TARGET_CLUSTER_NAME}.kubeconfig"

echo "add context target-cluster"
kubectl config set-context ${TARGET_CLUSTER_NAME} --user ${TARGET_CLUSTER_NAME}-admin --cluster ${TARGET_CLUSTER_NAME} \
--kubeconfig "/tmp/${TARGET_CLUSTER_NAME}.kubeconfig"

echo "apply cni as a part of initinfra-networking"
airshipctl phase run initinfra-networking-target --debug --kubeconfig "/tmp/${TARGET_CLUSTER_NAME}.kubeconfig"

echo "Check nodes status"
kubectl --kubeconfig /tmp/"${TARGET_CLUSTER_NAME}".kubeconfig wait --for=condition=Ready nodes --all --timeout 4000s
kubectl get nodes --kubeconfig /tmp/"${TARGET_CLUSTER_NAME}".kubeconfig

echo "Waiting for  pods to come up"
kubectl --kubeconfig /tmp/${TARGET_CLUSTER_NAME}.kubeconfig  wait --for=condition=ready pods --all --timeout=4000s -A
kubectl --kubeconfig /tmp/${TARGET_CLUSTER_NAME}.kubeconfig get pods -A

echo "Check machine status"
kubectl get machines --kubeconfig ${KUBECONFIG}

echo "Get cluster state for target workload cluster "
kubectl --kubeconfig ${KUBECONFIG} get cluster

echo "Target Cluster Kubeconfig"
echo "/tmp/${TARGET_CLUSTER_NAME}.kubeconfig"
