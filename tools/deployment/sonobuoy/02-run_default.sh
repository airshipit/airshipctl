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
: ${KUBECONFIG:="$HOME/.airship/kubeconfig"}
# Available Modes: quick, certified-conformance, non-disruptive-conformance.
# (default quick)
: ${CONFORMANCE_MODE:="quick"}
: ${TIMEOUT:=10800}
: ${TARGET_CLUSTER_CONTEXT:="target-cluster"}

mkdir -p /tmp/sonobuoy_snapshots/e2e
cd /tmp/sonobuoy_snapshots/e2e

# Run aggregator, and default plugins e2e and systemd-logs
sonobuoy run --plugin e2e --plugin systemd-logs -m ${CONFORMANCE_MODE} \
--context "$TARGET_CLUSTER_CONTEXT" \
--kubeconfig ${KUBECONFIG} \
--wait --timeout ${TIMEOUT} \
--log_dir /tmp/sonobuoy_snapshots/e2e

# Get information on pods
kubectl get all -n sonobuoy --kubeconfig ${KUBECONFIG} --context "$TARGET_CLUSTER_CONTEXT"

# Check sonobuoy status
sonobuoy status --kubeconfig ${KUBECONFIG} --context "$TARGET_CLUSTER_CONTEXT"

# Get logs
sonobuoy logs --kubeconfig ${KUBECONFIG} --context "$TARGET_CLUSTER_CONTEXT"

# Store Results
results=$(sonobuoy retrieve --kubeconfig ${KUBECONFIG} --context $TARGET_CLUSTER_CONTEXT)
echo "Results: ${results}"

# Display Results
sonobuoy results $results
ls -ltr /tmp/sonobuoy_snapshots/e2e