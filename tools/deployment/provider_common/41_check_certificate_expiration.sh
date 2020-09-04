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

export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}
export TARGET_CLUSTER_NAME=${TARGET_CLUSTER_NAME:-"target-cluster"}
export KUBECONFIG_TARGET_CONTEXT=${KUBECONFIG_TARGET_CONTEXT:-"target-cluster"}

check_nodecerts(){
    nodecerts=""
    if $2; then
        nodecerts=$(airshipctl --kubeconfig "/tmp/${TARGET_CLUSTER_NAME}.kubeconfig" cluster check-certificate-expiration --threshold $1 -o json | jq .nodeCerts)
    else
        nodecerts=$(airshipctl --kubeconfig ${KUBECONFIG} cluster check-certificate-expiration --threshold $1 -o json | jq .nodeCerts)
    fi
    nodecert=$(echo $nodecerts | jq '. | length')
    if [ -z $nodecert ]; then
        echo "Unable to verify node certificate expiration. Exiting!"
        exit 1
    else
        verify=false
        for ((i=0;i<nodecert;i++)); do
            nodeName=$(echo $nodecerts | jq -r ".[$i].name")
            if [[ $3 == $nodeName ]]; then
                echo "Node certificate on $nodeName verified"
                verify=true
                break
            fi
        done
        if $verify; then
            echo "Node certificate annotation verified!"
        else
            echo "Node certificate verification failed!"
	    exit 1
        fi
    fi
}

check_tls_certs(){
    if [[ $1 = "json" ]]; then
        tlscerts=$(airshipctl cluster check-certificate-expiration --threshold 365 -o json --kubeconfig ${KUBECONFIG} | jq '.tlsSecrets')
    else
	tlscerts=$(airshipctl cluster check-certificate-expiration --threshold 365 -o yaml --kubeconfig ${KUBECONFIG} | yq '.tlsSecrets')
    fi
    tlscert=$(echo $tlscerts | jq '. | length')
    secret_names=$(kubectl --kubeconfig ${KUBECONFIG} get secrets -A --field-selector="type=kubernetes.io/tls" -o jsonpath='{.items[*].metadata.name}')

    if [[ $tlscert -gt 0 ]]; then
        for ((i=0;i<tlscert;i++)); do
            secretName=$(echo $tlscerts | jq -r ".[$i].name")
            if [[ ${secret_names[@]} =~ $secretName ]]; then
                echo "Secret name matched for 365 day threshold!"
                continue
            else
                echo "Secret name, $secretName not matched for 365 day threshold!"
                exit 1
            fi
        done
    else
        echo "No TLS certificate expires in 365 days threshold!"
        exit 1
    fi
}

echo "Checking TLS certificate expiry within 365 days in JSON format"
check_tls_certs json

echo "Checking TLS certificate expiry within 365 days in YAML format"
check_tls_certs yaml

echo "Checking kubeconfig certificates"
kubecerts=$(airshipctl --kubeconfig "/tmp/${TARGET_CLUSTER_NAME}.kubeconfig" cluster check-certificate-expiration --threshold 365 | jq '.kubeconfs')
kubecert=$(echo $kubecerts | jq '. | length ')
if [[ $kubecert -gt 0 ]]; then
    kubecertName=$(echo $kubecerts | jq -r ".[0].secretName")
    kubecertNamespace=$(echo $kubecerts | jq -r ".[0].secretNamespace")
    echo "Copying target kubeconfig secret to another namespace to validate if this new kubenconfig will also be listed"
    kubectl --kubeconfig "/tmp/${TARGET_CLUSTER_NAME}.kubeconfig" create ns testing
    kubectl --kubeconfig "/tmp/${TARGET_CLUSTER_NAME}.kubeconfig" get secret $kubecertName --namespace $kubecertNamespace --export -o yaml | kubectl --kubeconfig "/tmp/${TARGET_CLUSTER_NAME}.kubeconfig" apply --namespace=testing -f -
    kubecerts=$(airshipctl --kubeconfig "/tmp/${TARGET_CLUSTER_NAME}.kubeconfig" cluster check-certificate-expiration --threshold 365 -o json | jq .kubeconfs)
    kubecert=$(echo $kubecerts | jq '. |length ')
    verify=false
    for ((i=0;i<kubecert;i++)); do
        kubecertNewName=$(echo $kubecerts | jq -r ".[$i].secretName")
	kubecertNewNamespace=$(echo $kubecerts | jq -r ".[$i].secretNamespace")
	if [[ $kubecertNewNamespace == "testing" ]] && [[ $kubecertNewName == $kubecertName ]]; then
            echo "Kubecert with name $kubecertName verified in testing namespace !"
	    verify=true
	    break
        fi
    done
    kubectl --kubeconfig "/tmp/${TARGET_CLUSTER_NAME}.kubeconfig" delete ns testing
    if !$verify; then
        echo "Kubecerts verification failed!"
        exit 1
    fi
else
    echo "Unable to get kubecertificates!!"
    exit 1
fi


MgmtNode=$(kubectl --kubeconfig ${KUBECONFIG} get nodes -o jsonpath={.items[0].metadata.name})
echo "Checking Node Certificate Expiry of $MgmtNode Node"

end=$(date --date='now next Year' '+%b %d, %Y %H:%M %Z' -u)
kubectl --kubeconfig ${KUBECONFIG} annotate node $MgmtNode cert-expiration="{ admin.conf: $end },{ apiserver: $end },{ apiserver-etcd-client: $end },{ apiserver-kubelet-client: $end },{ controller-manager.conf: $end },{ etcd-healthcheck-client: $end },{ etcd-peer: $end },{ etcd-server: $end },{ front-proxy-client: $end },{ scheduler.conf: $end },{ ca: $end },{ etcd-ca: $end },{ front-proxy-ca: $end}" --overwrite=true

check_nodecerts 365 false $MgmtNode

end=$(date '+%b %d, %Y %H:%M %Z' -u)
kubectl --kubeconfig ${KUBECONFIG} annotate node $MgmtNode cert-expiration="{ admin.conf: $end },{ apiserver: $end },{ apiserver-etcd-client: $end },{ apiserver-kubelet-client: $end },{ controller-manager.conf: $end },{ etcd-healthcheck-client: $end },{ etcd-peer: $end },{ etcd-server: $end },{ front-proxy-client: $end },{ scheduler.conf: $end },{ ca: $end },{ etcd-ca: $end },{ front-proxy-ca: $end}" --overwrite=true

check_nodecerts 1 false $MgmtNode

TargetNode=$(kubectl get nodes --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT | grep master | awk '{print $1}' | head -1)

end=$(date --date='now next Year' '+%b %d, %Y %H:%M %Z' -u)
kubectl --kubeconfig $KUBECONFIG --context $KUBECONFIG_TARGET_CONTEXT annotate node $TargetNode cert-expiration="{ admin.conf: $end },{ apiserver: $end },{ apiserver-etcd-client: $end },{ apiserver-kubelet-client: $end },{ controller-manager.conf: $end },{ etcd-healthcheck-client: $end },{ etcd-peer: $end },{ etcd-server: $end },{ front-proxy-client: $end },{ scheduler.conf: $end },{ ca: $end },{ etcd-ca: $end },{ front-proxy-ca: $end}" --overwrite=true

check_nodecerts 365 true $TargetNode
