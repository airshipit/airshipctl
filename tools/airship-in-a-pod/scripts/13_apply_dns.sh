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

# This utilizes some of the work that was done to the nameservers in
# 01_dns_settings.sh to take those DNS servers and force the coredns pod
# of the minikube cluster to use those instead of the default.


# Grab a list of the nameservers IPs in /etc/resolv.conf
NAMESERVERS=$(grep nameserver /etc/resolv.conf | awk '{print $2}' | tr '\n' ' ')


kubectl -n kube-system get pods -o wide
# Configure coredns with an upstream DNS to ensure the pod can resolve
# domains outside of the cluster
kubectl -n kube-system get cm -o yaml coredns | sed "s/\/etc\/resolv\.conf/$NAMESERVERS/" > tools/airship-in-a-pod/coredns-upstream-dns.yaml
cat tools/airship-in-a-pod/coredns-upstream-dns.yaml
kubectl apply -f tools/airship-in-a-pod/coredns-upstream-dns.yaml
kubectl rollout restart -n kube-system deployment/coredns
