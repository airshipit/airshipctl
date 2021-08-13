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

# This script starts up minikube, and accounts for an issue that sometimes
# comes up when running minikube for the first time in some environments


set +e
sudo -E minikube start --driver=none
status=$?
sudo chown -R "$USER" "$HOME"/.minikube; chmod -R u+wrx "$HOME"/.minikube
if [[ $status -gt 0 ]]; then
    # Sometimes minikube fails to start if the directory permissions are not correct
    sudo -E minikube delete
    set -e
    sudo -E minikube start --driver=none
fi

set -e
sudo -E minikube status

# Ensure .kube and .minikube have proper ownership
sudo chown -R "$USER" "$HOME"/.kube "$HOME"/.minikube

# Make a copy of the kubeconfig for the log playbooks
mkdir -p "$HOME"/.airship
cp "$HOME"/.kube/config "$HOME"/.airship/kubeconfig

# Give cluster a chance to start up
sleep 10