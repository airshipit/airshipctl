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

set -ex

# This script starts up kind cluster set for dual stack and proxy mode set for ipvs

set +e

create_cluster(){
# possibly a cluster by name kind exists
kind delete cluster --name kind
cat <<EOF | kind create cluster --config -
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true # disable kindnet
  kubeProxyMode: "ipvs"
  ipFamily: dual
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30002
    hostPort: 30002
    listenAddress: 0.0.0.0 # Optional, defaults to 0.0.0.0
    protocol: tcp # Optional, defaults to tcp
EOF
}

create_cluster

status=$?
if [[ $status -gt 0 ]]; then
   echo "Failed to create kind cluster"
   return -1
fi

# Give cluster a chance to start up
sleep 10


kubectl create -f https://docs.projectcalico.org/manifests/tigera-operator.yaml


cat <<EOF | kubectl apply -f -
apiVersion: operator.tigera.io/v1
kind: Installation
metadata:
  name: default
spec:
  # Configures Calico networking.
  calicoNetwork:
    # Note: The ipPools section cannot be modified post-install.
    ipPools:
    - blockSize: 26
      cidr: 10.244.0.0/16
      encapsulation: VXLANCrossSubnet
      natOutgoing: Enabled
      nodeSelector: all()
    - blockSize: 116 # must be greater than 115 and < 128
      cidr: fd00:10:244::/56
      encapsulation: None # Does not support
      natOutgoing: Enabled
      nodeSelector: all()
    nodeAddressAutodetectionV4:
      interface: eth0
    nodeAddressAutodetectionV6:
      interface: eth0

---

# This section configures the Calico API server.
apiVersion: operator.tigera.io/v1
kind: APIServer
metadata:
  name: default
spec: {}
EOF

# Make a copy of the kubeconfig for the log playbooks
mkdir -p "$HOME"/.airship
kind get kubeconfig > "$HOME"/.airship/kubeconfig

# Ensure all of the downloaded images are loaded into kind
# Redefining the environment variables (as export does not seem to work in Zuul environment)

BUILD_LIST="status-checker artifact-setup base infra-builder runner"
PULL_LIST="docker:stable-dind nginx quay.io/metal3-io/sushy-tools quay.io/airshipit/libvirt:aiap-v1"

kind load docker-image ${PULL_LIST}

for IMAGE in ${BUILD_LIST}; do
	kind load docker-image "quay.io/airshipit/aiap-$IMAGE:latest"
done

