# Airship 2.0 Integration with Azure Cloud Platform
This document provides the instructions to setup and execute *airshipctl*
commands to deploy a Target cluster in Azure cloud platform.
The manifest for the Target cluster deployment can be found at
**manifest/site/az-test-site/target/azure-target**.
It will deploy:
- CAPZ v0.4.5 Management component
- Region: US East
- Control Plane: 1 VM (Standard_B2s)
- Worker: 2 VMs (Standard_B2s)
- Deploying K8S 1.18.3

## Pre-requisites
The list below are the expected pre-requisites for this integration.

- Create your *$HOME/.airship/config*
- Instantiate the Management cluster using Kind
- Update the manifest *manifest/function/capz/v.4.5/default/credentials.yaml*
with the Azure subscription credentials

TODO: Azure subscription credentials to be passed as environment variables

## Steps to create a Management cluster with Kind
The list of commands below creates a K8S cluster to be used as Management cluster

```bash
$ kind create cluster --name airship2-kind-api --kubeconfig /your/folder/kubeconfig.yaml
$ cp /your/folder/kubeconfig.yaml $HOME/.airship/kubeconfig
$ cp /your/folder/kubeconfig.yaml $HOME/.kube/config
```

## Initialize Management cluster
Execute the following command to initialize the Management cluster with CAPI and
CAPZ components.
```bash
$ airshipctl cluster init
```
## Deploy Target cluster on Azure
To deploy the Target cluster on Azure cloude execute the following command.
```bash
$ airshipctl phase apply azure-target
```

Verify the status of Target cluster deployment
```bash
$ kubectl get cluster --all-namespaces
```
Check status of Target cluster KUBEADM control plane deployment
```bash
$ kubectl get kubeadmcontrolplane --all-namespaces
```

Retrieve the kubeconfig of Target cluster
```bash
$ kubectl --namespace=default get secret/az-target-cluster-kubeconfig -o jsonpath={.data.value} \
| base64 --decode > ./az-target-cluster.kubeconfig
```

Check the list of nodes create for the Target cluster
```bash
  $ kubectl --kubeconfig=./az-target-cluster.kubeconfig get nodes
```

When all control plane and worker nodes have been created, they will stay in Not Ready state until
CNI is configured. See next step below.

## Configure CNI on the Target cluster with Calico
Calico will be initialized as part of control plane VM *postKubeadmCommands*, which executes the
*sudo kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api-provider-azure/master/templates/addons/calico.yaml* command.

See snippet of manifest integrating Calico initialization below:

```yaml
apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
kind: KubeadmControlPlane
metadata:
  name: az-target-cluster-control-plane
  namespace: default
spec:
  infrastructureTemplate:
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
    kind: AzureMachineTemplate
    name: az-target-cluster-control-plane
  kubeadmConfigSpec:
...
    files:
    - path: /calico.sh
      owner: root:root
      permissions: "0755"
      content: |
        #!/bin/bash -x
        sudo kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api-provider-azure/master/templates/addons/calico.yaml
...
    postKubeadmCommands:
      - /calico.sh
    useExperimentalRetryJoin: true
  replicas: 3
  version: v1.18.2
```

This approach automates the initialization of Calico and saves the need to execute manually
the list of commands described below.

First we need to provision the Target cluster context in the airship config file

Add Target Cluster manifest to azure_manifest
```bash
$ airshipctl config import ./az-target-cluster.kubeconfig
```
Replace Target Cluster kubeconfig Context in the airship config file
```bash
$ airshipctl config set-context az-target-cluster-admin@az-target-cluster --manifest azure_manifest
```

Set Current Context to the Target Cluster kubeconfig Context in the airship config file
```bash
$ airshipctl config use-context az-target-cluster-admin@az-target-cluster
```

Now we can trigger the configuration of Calico on the Target Cluster
```bash
$ airshipctl phase apply calico --kubeconfig az-target-cluster.kubeconfig
```

Once the Calico provisionning has been completed you should see all the nodes instantiated for the
Target cluster in Ready state.
```bash
$ kubectl --kubeconfig=./az-target-cluster.kubeconfig get nodes

NAME                                    STATUS   ROLES    AGE   VERSION
az-target-cluster-control-plane-28ghk   Ready    master   17h   v1.18.2
az-target-cluster-md-0-46zfv            Ready    <none>   17h   v1.18.2
az-target-cluster-md-0-z5lff            Ready    <none>   17h   v1.18.2
```

## APPENDIX: $HOME/.airship/config

```yaml
apiVersion: airshipit.org/v1alpha1
contexts:
  az-target-cluster-admin@az-target-cluster:
    contextKubeconf: az-target-cluster_target
    manifest: azure_manifest
currentContext: az-target-cluster-admin@az-target-cluster
kind: Config
managementConfiguration:
  azure_management_config:
    insecure: true
    systemActionRetries: 30
    systemRebootDelay: 30
    type: azure
  default:
    systemActionRetries: 30
    systemRebootDelay: 30
    type: azure
manifests:
  azure_manifest:
    primaryRepositoryName: primary
    repositories:
      primary:
        checkout:
          branch: master
          commitHash: ""
          force: false
          tag: ""
        url: https://review.opendev.org/airship/airshipctl
    subPath: airshipctl/manifests/site/az-test-site
    targetPath: /tmp/airship
  default:
    primaryRepositoryName: primary
    repositories:
      primary:
        checkout:
          branch: master
          commitHash: ""
          force: false
          tag: ""
        url: https://opendev.org/airship/treasuremap
    subPath: treasuremap/manifests/site
    targetPath: /tmp/default
```