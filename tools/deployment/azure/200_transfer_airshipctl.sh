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

if [ -n "$3" ]; then
    echo "--- Remote username@VM = $1"
    echo "--- Local Airship Dir = $2"
    echo "--- Credentials script = $3"

    export REMOTE_VM=$1
    export LOCAL_AIRSHIP_DIR=$2
    export CREDENTIALS=$3
else
    echo "Syntax: 200_transfer_airshipctl.sh <Remote VM Username> <Remote VM IP> <Local Airship Dir>"
    echo "    <Remote VM>: Username@VM to login to the Remote VM"
    echo "    <Local Airship Dir>: Directory containing the Airship project, e.g., /home/esidshi/projects/airshipctl/"
    echo "    <Credentials script>: used by remote VM to set public Cloud credentials"
    exit 1
fi

export REMOTE_USERNAME=$(echo $REMOTE_VM | cut -d'@' -f 1)
echo "Remote Username = $REMOTE_USERNAME"
echo "Remote VM = $REMOTE_VM"
echo "Local Airshipt Dir = $LOCAL_AIRSHIP_DIR"
echo "Credentials Script = $CREDENTIALS"

# Preparing the Remote VM to the "Zero Touch" Validation
cd $LOCAL_AIRSHIP_DIR
echo "sudo mkdir /home/zuul"
ssh -o StrictHostKeyChecking=no $REMOTE_VM 'sudo mkdir /home/zuul'

echo "sudo chown ${REMOTE_USERNAME} /home/zuul"
ssh  $REMOTE_VM "sudo chown ${REMOTE_USERNAME} /home/zuul"

echo "sudo chgrp ${REMOTE_USERNAME} /home/zuul"
ssh  $REMOTE_VM "sudo chgrp ${REMOTE_USERNAME} /home/zuul"

echo "mkdir /home/zuul/src"
ssh  $REMOTE_VM 'mkdir /home/zuul/src'

echo "mkdir /home/zuul/src/opendev.org"
ssh  $REMOTE_VM 'mkdir /home/zuul/src/opendev.org'

echo "mkdir /home/zuul/src/opendev.org/airship"
ssh  $REMOTE_VM 'mkdir /home/zuul/src/opendev.org/airship'

echo "scp -r $LOCAL_AIRSHIP_DIR/  $REMOTE_VM:/home/zuul/src/opendev.org/airship/airshipctl"
scp -r $LOCAL_AIRSHIP_DIR/  $REMOTE_VM:/home/zuul/src/opendev.org/airship/airshipctl

echo "scp $LOCAL_AIRSHIP_DIR/tools/deployment/azure/201_zuul_local_test.sh  $REMOTE_VM:~"
scp $LOCAL_AIRSHIP_DIR/tools/deployment/azure/201_zuul_local_test.sh  $REMOTE_VM:~
scp $CREDENTIALS  $REMOTE_VM:~

# echo "ssh  $REMOTE_VM 'bash ~/201_zuul_local_test.sh'"
export CREDENTIALS_FILENAME=$(echo ${CREDENTIALS##*/}) # extract the script filename only
echo "CREDENTIALS_FILENAME = $CREDENTIALS_FILENAME"

# Setting Public Cloud credentials as environment variables in the remote VM
ssh $REMOTE_VM "cat ${CREDENTIALS_FILENAME} >> ~/.profile"

# Executing the local test
ssh $REMOTE_VM '/bin/bash ~/201_zuul_local_test.sh'
