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

set -x

sudo rm -rf ~/.airship/ ~/.ansible.cfg /srv/iso/*
sudo service sushy-tools stop
sudo service apache2 stop

vm_types='ephemeral|target|worker'

vol_list=$(sudo virsh vol-list --pool airship | grep -E $vm_types | awk '{print $1}')
iso_list=$(sudo virsh vol-list --pool default | awk '{print $1}'| grep -i 'ubuntu.*\.img$')
vm_list=$(sudo virsh list --all | grep -E $vm_types | awk '{print $2}')
net_list=$(sudo virsh net-list --all | awk '{print $1}'| grep -i air)

for vm in $vm_list; do sudo virsh destroy $vm; sudo virsh undefine $vm --nvram --remove-all-storage; done
for vol in $vol_list; do sudo virsh vol-delete $vol --pool airship; done
for iso in $iso_list; do sudo virsh vol-delete $iso --pool default; done
for net in $net_list; do sudo virsh net-destroy $net; sudo virsh net-undefine $net; done
