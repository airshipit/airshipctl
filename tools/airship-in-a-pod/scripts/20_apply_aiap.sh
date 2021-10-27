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

# This script first loops to check if the K8s cluster is ready before applying
# airship-in-a-pod to it. Once all of the kube-system pods are ready, it applies
# the yaml and then checks every so often to determine if all of the containers
# are ready.


INTERVAL=15
# Number of statuses that the status checker is looking at
STATUS_CHECKER_NUM=3

READY=false
KUBE_READY=false



# Wait for the Kubernetes environment to become completely ready
while [[ $KUBE_READY == "false" ]];
do
    # Grab the readiness from the kubectl output
    kube_pods=$(kubectl get pods -n kube-system | tail -n +2 | awk '{print $2}')
    for POD in $kube_pods; do
        # Compare the two values to determine if each pod is completely ready
        kube_ready_pod=$(echo "$POD" | cut -f1 -d/)
        kube_ready_total=$(echo "$POD" | cut -f2 -d/)
        if [[ "$kube_ready_pod" != "$kube_ready_total" ]]; then
            # If a pod is not ready yet, break out and try again next time
            KUBE_READY=false
            break
        fi
        # This will only stay "true" as long as the previous 'if' is never reached
        KUBE_READY=true
    done
    sleep "$INTERVAL"
done

kustomize build tools/airship-in-a-pod/examples/airshipctl | kubectl apply -f -

while [[ $READY == "false" ]];
do
    # Grab the number of ready containers from the kubectl output
    kubectl get pod airship-in-a-pod -o wide
    readiness=$(kubectl get pods | grep airship-in-a-pod | awk '{print $2}')
    ready_pod=$(echo "$readiness" | cut -f1 -d/)
    ready_total=$(echo "$readiness" | cut -f2 -d/)
    # if it is 7/7 ready (example with 7 containers), then mark ready
    if [[ "$ready_pod" == "$ready_total" ]]; then
        READY=true
    fi

    sleep "$INTERVAL"
done

# Now that all of the containers are ready, keep an eye on the status-checker
# container until it shows that every status succeeds, and exit on any failure
while [[ $(kubectl logs airship-in-a-pod -c status-checker --tail 1 | grep -o "SUCCESS" | wc -l) -lt ${STATUS_CHECKER_NUM} ]]
do
    # If any status is listed as "FAILED" then we should kill the script with a non 0 exit code
    if [[ $(kubectl logs airship-in-a-pod -c status-checker --tail 1 | grep -c -o "FAILED") -gt 0 ]]; then
        kubectl logs airship-in-a-pod -c status-checker --tail 1
        exit 3
    fi
	sleep "$INTERVAL"
done
