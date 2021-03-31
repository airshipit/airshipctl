# Airshipctl and Cluster API Docker Integration

## Overview

This document demonstrates usage of airshipctl to create kubernetes clusters
locally using docker and kind. Airshipctl requires an existing kubernetes cluster
accessible via kubectl. We will use kind as a local bootstrap cluster, to provision
a target management cluster on the docker infrastructure provider.
The target management cluster will then be used to create a workload cluster
with one or more worker nodes.

## Workflow

- create a single node kubernetes cluster using kind
- initialize the kind cluster with cluster api management components
- use the kind management cluster to create a target cluster with one control
  plane
- apply cni solution on the target cluster
- initialize the target cluster with cluster api management components
- move the cluster api management crds from kind cluster to target management
  cluster
- tear down the kind management cluster
- use the target management cluster to create worker nodes
- increase/decrease the worker count as required

## Airshipctl Commands Used And Purpose

```
Pull documents from the remote git repository
> airshipctl document pull

Initialize the kind cluster with cluster api and docker provider components
> airshipctl phase run clusterctl-init-ephemeral

Use the management cluster to create a target cluster with one control plane
> airshipctl phase run controlplane-ephemeral

Get multiple contexts for every cluster in the airship site
> airshipctl cluster get-kubeconfig > ~/.airship/kubeconfig-tmp

Initialize CNI on target cluster`
> airshipctl phase run initinfra-networking-target

Initialize Target Cluster with cluster api and docker proivder components
> airshipctl phase run clusterctl-init-target

Move managment CRDs from kind management cluster to target management cluster
> airshipctl phase run clusterctl-move

Use target management cluster to deploy workers
> airshipctl phase run  workers-target
```

## Getting Started

### Build [Airshipctl](https://docs.airshipit.org/airshipctl/developers.html)

```
$ git clone https://review.opendev.org/airship/airshipctl

$ cd airshipctl

$ ./tools/deployment/21_systemwide_executable.sh
```

### Create airship configuration

```
$ cat ~/.airship/config

apiVersion: airshipit.org/v1alpha1
managementConfiguration:
  dummy_management_config:
    type: redfish
    insecure: true
    useproxy: false
    systemActionRetries: 30
    systemRebootDelay: 30
contexts:
  ephemeral-cluster:
    contextKubeconf: ephemeral-cluster_ephemeral
    manifest: dummy_manifest
    managementConfiguration: dummy_management_config
  target-cluster:
    contextKubeconf: target-cluster_target
    manifest: dummy_manifest
    managementConfiguration: dummy_management_config
currentContext: ephemeral-cluster
kind: Config
manifests:
  dummy_manifest:
    phaseRepositoryName: primary
    repositories:
      primary:
        checkout:
          branch: master
          force: false
          remoteRef: ""
          tag: ""
        url: https://review.opendev.org/airship/airshipctl
    metadataPath: manifests/site/docker-test-site/metadata.yaml
    targetPath: /tmp/airship
```

### Deploy Control plane and Workers

```
$ export KIND_EXPERIMENTAL_DOCKER_NETWORK=bridge

$ export KUBECONFIG="${HOME}/.airship/kubeconfig"

$ kind create cluster --name ephemeral-cluster --wait 120s \
--kubeconfig "${HOME}/.airship/kubeconfig" \
--config ./tools/deployment/templates/kind-cluster-with-extramounts

$ kubectl config set-context ephemeral-cluster \
--cluster kind-ephemeral-cluster \
--user kind-ephemeral-cluster --kubeconfig $KUBECONFIG

$ kubectl config set-context target-cluster --user target-cluster-admin \
--cluster target-cluster --kubeconfig  $KUBECONFIG

$ airshipctl document pull -n --debug

$ airshipctl phase run clusterctl-init-ephemeral --debug --wait-timeout 300s

$ airshipctl phase run controlplane-ephemeral --debug --wait-timeout 300s

$ airshipctl cluster get-kubeconfig > ~/.airship/kubeconfig-tmp

$ mv ~/.airship/kubeconfig-tmp "${KUBECONFIG}"

$ airshipctl phase run initinfra-networking-target --debug

$ kubectl --context target-cluster wait \
--for=condition=Ready nodes --all --timeout 300s

$ kubectl get nodes --context target-cluster -A

```

Note: Please take note of the control plane node name from the output of previous
command because it is untainted in the next step. For eg. control plane node
name could be something like target-cluster-control-plane-twwsv

```
$ kubectl taint node target-cluster-control-plane-twwsv \
node-role.kubernetes.io/master- --context target-cluster --request-timeout 10s

$ airshipctl phase run clusterctl-init-target --debug --wait-timeout 300s

$ kubectl get pods -A --context target-cluster

$ airshipctl phase run clusterctl-move --debug

$ kubectl get machines --context target-cluster

$ kind delete cluster --name "ephemeral-cluster"

$ airshipctl phase run  workers-target --debug

$ kubectl get machines --context target-cluster

NAME                                   PROVIDERID                                        PHASE
target-cluster-control-plane-m5jf7     docker:////target-cluster-control-plane-m5jf7     Running
target-cluster-md-0-84db44cdff-r8dkr   docker:////target-cluster-md-0-84db44cdff-r8dkr   Running

```

## Scale Workers

Worker count can be adjusted in airshipctl/manifests/site/docker-test-site/
target/workers/machine_count.json.

In this example, we have changed it to 3.

```

$ cat /tmp/airship/airshipctl/manifests/site/docker-test-site/target/workers/machine_count.json

[
  { "op": "replace","path": "/spec/replicas","value": 3 }
]

$ airshipctl phase run  workers-target --debug

$ kubectl get machines --kubeconfig /tmp/target-cluster.kubeconfig

NAME                                   PROVIDERID                                        PHASE
target-cluster-control-plane-m5jf7     docker:////target-cluster-control-plane-m5jf7     Running
target-cluster-md-0-84db44cdff-b6zp6   docker:////target-cluster-md-0-84db44cdff-b6zp6   Running
target-cluster-md-0-84db44cdff-g4nm7   docker:////target-cluster-md-0-84db44cdff-g4nm7   Running
target-cluster-md-0-84db44cdff-r8dkr   docker:////target-cluster-md-0-84db44cdff-r8dkr   Running
```

## Clean Up

```
$ kind get clusters
target-cluster

$ kind delete cluster --name target-cluster
```

## More Information

- worker count can be adjusted from airshipctl/manifests/site/docker-test-site/
target/workers/machine_count.json

- control plane count can be adjusted from airshipctl/manifests/site/
docker-test-site/ephemeral/controlplane/machine_count.json

## Reference

### Pre-requisites

* Install [Docker](https://www.docker.com/)
* Install [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* Install [Kind](https://kind.sigs.k8s.io/)
* Check [Software Version Information](#Software-Version-Information),
[Special Instructions](#Special-Instructions) and [Virtual Machine
Specification](#Virtual-Machine-Specification)


### Software Version Information

All the instructions provided in the document have been tested using the
software and version, provided in this section.

#### Virtual Machine Specification

All the instructions in the document were perfomed on a Oracle Virtual Box(6.1)
VM running Ubuntu 18.04.4 LTS (Bionic Beaver) with 16G of memory and 4 VCPUs

#### Docker
```
$ docker version

Client: Docker Engine - Community
 Version:           19.03.9
 API version:       1.40
 Go version:        go1.13.10
 Git commit:        9d988398e7
 Built:             Fri May 15 00:25:18 2020
 OS/Arch:           linux/amd64
 Experimental:      false

Server: Docker Engine - Community
 Engine:
  Version:          19.03.9
  API version:      1.40 (minimum version 1.12)
  Go version:       go1.13.10
  Git commit:       9d988398e7
  Built:            Fri May 15 00:23:50 2020
  OS/Arch:          linux/amd64
  Experimental:     false
 containerd:
  Version:          1.2.13
  GitCommit:        7ad184331fa3e55e52b890ea95e65ba581ae3429
 runc:
  Version:          1.0.0-rc10
  GitCommit:        dc9208a3303feef5b3839f4323d9beb36df0a9dd
 docker-init:
  Version:          0.18.0
  GitCommit:        fec3683
```

#### Kind

```
$ kind version

kind v0.8.1 go1.14.2 linux/amd64
```

#### Kubectl

```
$ kubectl version

Client Version: version.Info{Major:"1", Minor:"17", GitVersion:"v1.17.4", GitCommit:"8d8aa39598534325ad77120c120a22b3a990b5ea", GitTreeState:"clean", BuildDate:"2020-03-12T21:03:42Z", GoVersion:"go1.13.8", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"17", GitVersion:"v1.17.0", GitCommit:"70132b0f130acc0bed193d9ba59dd186f0e634cf", GitTreeState:"clean", BuildDate:"2020-01-14T00:09:19Z", GoVersion:"go1.13.4", Compiler:"gc", Platform:"linux/amd64"}
```

#### Go
```
$ go version
go version go1.14.1 linux/amd64
```

#### OS
```
$ cat /etc/os-release

NAME="Ubuntu"
VERSION="18.04.4 LTS (Bionic Beaver)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 18.04.4 LTS"
VERSION_ID="18.04"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=bionic
UBUNTU_CODENAME=bionic
```

### Special Instructions

Swap was disabled on the VM using `sudo swapoff -a`
