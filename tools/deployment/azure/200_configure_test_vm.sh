#!/bin/bash

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

# This script transfers the local Airship project to the remote machine,
# copies the test/validation script to the remote machine and then executes it.
#
# This script is a local test simulating Zuul gates used to test and validate
# the Azure clout integration.

if [ $# -eq 3 ]; then
    echo "--- Remote VM Username@IP = $1"
    echo "--- Local Airship Dir = $2"
    echo "--- Credentials Script = $3"

    export REMOTE_VM=$1
    export LOCAL_AIRSHIP_DIR=$2
    export CREDENTIALS=$3
else
    echo "Syntax: 200_configure-remote-vm.sh <Remote VM Username> <Remote VM IP> <Local Airship Dir>"
    echo "    <Remote VM Username>: Username@VM_IP to login to the Remote VM"
    echo "    <Local Airship Dir>: Directory containing the Airship project, e.g., /home/esidshi/projects/airshipctl/"
    echo "    <Credentials script>: script to be used by remote VM for setting the credentials for public Clouds"
    exit 1
fi

echo "Remote Username@VM = $REMOTE_VM"
echo "Local Airship Dir" = $LOCAL_AIRSHIP_DIR
echo "Credentials Script = $CREDENTIALS"

# Pushing local SSH Public Key to Remote VM
echo "Adding local VM public in the Remote VM ..."
ssh-copy-id -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa.pub $REMOTE_VM

# Installing Docker in the remote VM
echo "Installing Docker ..."
# ssh $REMOTE_VM 'sudo apt update -y && sudo apt install docker.io && sudo usermod -aG docker $USER'
ssh $REMOTE_VM 'sudo apt-get remove docker docker-engine docker.io containerd runc && sudo apt-get update'
ssh $REMOTE_VM 'sudo apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common'
ssh $REMOTE_VM 'curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add - && sudo apt-key fingerprint 0EBFCD88'
ssh $REMOTE_VM 'sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"'
ssh $REMOTE_VM 'sudo apt-get update && sudo apt-get install -y docker-ce docker-ce-cli containerd.io && sudo usermod -aG docker $USER'

# Installing Make in the remote VM
ssh $REMOTE_VM 'sudo apt-get update -y && sudo apt install make'

# Transfer the manifests to the remote VM and start the local test
$LOCAL_AIRSHIP_DIR/tools/deployment/azure/200_transfer_airshipctl.sh $REMOTE_VM $LOCAL_AIRSHIP_DIR $CREDENTIALS