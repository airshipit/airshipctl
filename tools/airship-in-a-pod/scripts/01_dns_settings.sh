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

# Running K8s pods in zuul can cause a lot of issues with resolving domains
# inside of the running pods and catching coredns in a loop. This aims to
# tackle a few issues that have been hit to resolve this.


NAMESERVER="1.0.0.1"

# Grab the real nameservers instead of the local one listed in the original
sudo ln -sf /run/systemd/resolve/resolv.conf /etc/resolv.conf

# Add the known good DNS server
sudo sed -i "1i\nameserver $NAMESERVER\n" /etc/resolv.conf
# Remove DNS servers pointing to localhost so coredns doesn't get caught in a loop
sudo sed -i '/127\.0/d' /etc/resolv.conf
# Spit out the nameservers for the logs
cat /etc/resolv.conf

# Running unbound server can cause issues with coredns, disabling
if [[ -f "/etc/unbound/unbound.pid" ]]; then
    sudo kill "$(cat /etc/unbound/unbound.pid)"
fi

# flush iptables so coredns doesn't get caught up
# be sure to stop docker if it is installed
dpkg -l | grep -i docker | head -1 | if [[ "$(cut -d ' ' -f 1)" == "ii" ]]; then
    sudo systemctl stop docker
fi
sudo iptables --flush
sudo iptables -tnat --flush

dpkg -l | grep -i docker | head -1 | if [[ "$(cut -d ' ' -f 1)" == "ii" ]]; then
    sudo systemctl start docker
fi
