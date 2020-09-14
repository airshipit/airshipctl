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
vm_list=$(sudo virsh list --all | grep -E $vm_types | awk '{print $2}')

for vm in $vm_list; do sudo virsh destroy $vm; sudo virsh undefine $vm --nvram; done
for vol in $vol_list; do sudo virsh vol-delete $vol --pool airship; done
