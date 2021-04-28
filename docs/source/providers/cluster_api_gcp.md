# Airshipctl and Cluster API GCP Provider Integration

## Overview

Airshipctl and cluster api gcp integration facilitates usage of `airshipctl` to
create cluster api management and workload clusters using `gcp as infrastructure
provider`.

## Workflow

A simple workflow that can be tested, involves the following
operations:

- create a single node kubernetes cluster using kind
- initialize the kind cluster with cluster api management components and
  capg infrastructure provider components
- use the kind management cluster to create a target cluster with one control
  plane
- apply cni solution on the target cluster
- initialize the target cluster with cluster api management components
- move the cluster api management crds from kind cluster to target management
  cluster
- tear down the kind management cluster
- use the target management cluster to create worker nodes

## Airshipctl commands used

```
Pull documents from the remote git repository
> airshipctl document pull

Initialize the kind cluster with cluster api and gcp provider components
> airshipctl phase run clusterctl-init-ephemeral

Use the management cluster to create a target cluster with one control plane
> airshipctl phase run controlplane-ephemeral

Get multiple contexts for every cluster in the airship site
> airshipctl cluster get-kubeconfig > ~/.airship/kubeconfig-tmp

Initialize CNI on target cluster`
> airshipctl phase run initinfra-networking-target

Initialize Target Cluster with cluster api and gcp proivder components
> airshipctl phase run clusterctl-init-target

Move managment CRDs from kind management cluster to target management cluster
> airshipctl phase run clusterctl-move

Use target management cluster to deploy workers
> airshipctl phase run  workers-target
```

## GCP Prerequisites

### Create Service Account

To create and manager clusters, this infrastructure providers uses a service
account to authenticate with GCP's APIs. From your cloud console, follow [these
instructions](https://cloud.google.com/iam/docs/creating-managing-service-accounts#creating)
to create a new service account with Editor permissions. Afterwards, generate a
JSON Key and store it somewhere safe. Use cloud shell to install ansible,
packer, and build the CAPI compliant vm image.

### Build Cluster API Compliant VM Image

#### Install Ansible

Start by launching cloud shell.

$ export GCP_PROJECT_ID=<project-id>

$ export GOOGLE_APPLICATION_CREDENTIALS=</path/to/serviceaccount-key.json>

$ sudo apt-get update

$ sudo apt-get install ansible -y

#### Install Packer

$ mkdir packer

$ cd packer

$ wget https://releases.hashicorp.com/packer/1.6.0/packer_1.6.0_linux_amd64.zip

$ unzip packer_1.6.0_linux_amd64.zip

$ sudo mv packer /usr/local/bin/

#### Build GCP Compliant CAPI-Ubuntu Image

$ git clone https://sigs.k8s.io/image-builder.git

$ cd image-builder/images/capi/

$ make build-gce-ubuntu-1804

List the image

$ gcloud compute images list --project ${GCP_PROJECT_ID} --no-standard-images --filter="family:capi-ubuntu-1804-k8s"

```
NAME                                         PROJECT      FAMILY                      DEPRECATED  STATUS
cluster-api-ubuntu-1804-v1-17-11-1607489276  airship-gcp  capi-ubuntu-1804-k8s-v1-17              READY
```

### Create Cloud NAT Router

Kubernetes nodes, to communicate with the control plane, pull container images
from registried (e.g. gcr.io or dockerhub) need to have NAT access or a public
ip. By default, the provider creates Machines without a public IP.

To make sure your cluster can communicate with the outside world, and the load
balancer, you can create a Cloud NAT in the region you'd like your Kubernetes
cluster to live in by following [these
instructions](https://cloud.google.com/nat/docs/using-nat#specify_ip_addresses_for_nat).

Below cloud NAT router is created in `us-east1` region.

![nat-router](https://i.imgur.com/TKO6xSE.png)

## Getting Started

Kind will be used to setup a kubernetes cluster, that will be later transformed
into a management cluster using airshipctl. The kind kubernetes cluster will be
initialized with cluster API and Cluster API gcp provider components.

$ export KIND_EXPERIMENTAL_DOCKER_NETWORK=bridge

$ export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}

$ kind create cluster --name ephemeral-cluster --wait 120s \
--kubeconfig "$KUBECONFIG"

```bash
Creating cluster "ephemeral-cluster" ...
WARNING: Overriding docker network due to KIND_EXPERIMENTAL_DOCKER_NETWORK
WARNING: Here be dragons! This is not supported currently.
 ✓ Ensuring node image (kindest/node:v1.19.1) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
 ✓ Waiting ≤ 3m20s for control-plane = Ready ⏳
 • Ready after 1m3s 💚
Set kubectl context to "kind-ephemeral-cluster"
You can now use your cluster with:

kubectl cluster-info --context kind-ephemeral-cluster

Thanks for using kind! 😊
```

$ kubectl get pods -A

```bash
NAMESPACE            NAME                                                      READY   STATUS    RESTARTS   AGE
kube-system          coredns-f9fd979d6-g8wrd                                   1/1     Running   0          3m22s
kube-system          coredns-f9fd979d6-wrc5r                                   1/1     Running   0          3m22s
kube-system          etcd-ephemeral-cluster-control-plane                      1/1     Running   0          3m32s
kube-system          kindnet-p8bx7                                             1/1     Running   0          3m22s
kube-system          kube-apiserver-ephemeral-cluster-control-plane            1/1     Running   0          3m32s
kube-system          kube-controller-manager-ephemeral-cluster-control-plane   1/1     Running   0          3m32s
kube-system          kube-proxy-zl7jg                                          1/1     Running   0          3m22s
kube-system          kube-scheduler-ephemeral-cluster-control-plane            1/1     Running   0          3m32s
local-path-storage   local-path-provisioner-78776bfc44-q7gtr                   1/1     Running   0          3m22s
```

## Create airshipctl configuration files

Create airshipctl configuration to use `gcp-test-site`.

$ cat ~/.airship/config

```bash
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
    manifest: dummy_manifest
    managementConfiguration: dummy_management_config
  target-cluster:
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
    metadataPath: manifests/site/gcp-test-site/metadata.yaml
    targetPath: /tmp/airship
```

$ kubectl config set-context ephemeral-cluster \
--cluster kind-ephemeral-cluster \
--user kind-ephemeral-cluster --kubeconfig $KUBECONFIG

$ kubectl config set-context target-cluster --user target-cluster-admin \
--cluster target-cluster --kubeconfig  $KUBECONFIG

$ airshipctl document pull --debug

### Configure Environment Variables

For GCP provider, following envs should be set with correct values as per the google cloud project.

All values should be in Base64 encoded format.

Replace these values with specific configuration and credential as per your google cloud project configuration.

```bash
$cat gcp_env

GCP_CONTROL_PLANE_MACHINE_TYPE="bjEtc3RhbmRhcmQtNA=="
GCP_NODE_MACHINE_TYPE="bjEtc3RhbmRhcmQtNA=="
GCP_REGION="dXMtZWFzdDE="
GCP_NETWORK_NAME="ZGVmYXVsdA=="
GCP_PROJECT="YWlyc2hpcC1nY3A="
GCP_B64ENCODED_CREDENTIALS="base64 encoded value of </path/to/serviceaccount-key.json>"
```

Export all the envs
$ export $(cat gcp_env)

## Initialize `ephemeral-cluster` with `capi` and `capg` components

$ airshipctl phase run clusterctl-init-ephemeral --debug --wait-timeout 300s

## Deploy control plane nodes in the `target-cluster`

$ airshipctl phase run controlplane-ephemeral --debug --wait-timeout 300s

To check logs run the below command

$  kubectl logs capg-controller-manager-xxxxxxxxx-yyyyy -n capg-system --all-containers=true -f --kubeconfig $KUBECONFIG

$ kubectl get machines

```bash
NAME                                 PROVIDERID                                                        PHASE
target-cluster-control-plane-pbf4n   gce://airship-gcp/us-east1-b/target-cluster-control-plane-qkgtx   Running
```

$ airshipctl cluster get-kubeconfig > ~/.airship/kubeconfig-tmp

$ mv ~/.airship/kubeconfig-tmp "${KUBECONFIG}"

## Deploy Calico cni in the `target-cluster`

```bash
$ kubectl get nodes --context target-cluster
NAME                                 STATUS     ROLES    AGE     VERSION
target-cluster-control-plane-qkgtx   NotReady   master   5h53m   v1.17.11
```

Deploy calico cni using `initinfra-networking` phase

$ airshipctl phase run initinfra-networking-target --debug

Check on control plane node status. It should be in `Ready` state.

$ kubectl get nodes --context target-cluster
NAME                                 STATUS   ROLES    AGE     VERSION
target-cluster-control-plane-qkgtx   Ready    master   5h59m   v1.17.11

Check all pods including calico pods

```bash
$ kubectl get po -A --kubeconfig target-cluster.kubeconfig
NAMESPACE         NAME                                                         READY   STATUS    RESTARTS   AGE
calico-system     calico-kube-controllers-55cc6844cb-h4gzh                     1/1     Running   0          2m11s
calico-system     calico-node-qdjsm                                            1/1     Running   1          2m11s
calico-system     calico-typha-667c57fb6b-kjpfz                                1/1     Running   0          2m12s
cert-manager      cert-manager-cainjector-55d9fb4b8-fk5z8                      1/1     Running   0          2m18s
cert-manager      cert-manager-dfbc75865-mfjz9                                 1/1     Running   0          2m18s
cert-manager      cert-manager-webhook-66fc9cf7c-fbgx4                         1/1     Running   0          2m18s
kube-system       coredns-6955765f44-pl4zv                                     1/1     Running   0          6h
kube-system       coredns-6955765f44-wwkxt                                     1/1     Running   0          6h
kube-system       etcd-target-cluster-control-plane-qkgtx                      1/1     Running   0          6h
kube-system       kube-apiserver-target-cluster-control-plane-qkgtx            1/1     Running   0          6h
kube-system       kube-controller-manager-target-cluster-control-plane-qkgtx   1/1     Running   0          6h
kube-system       kube-proxy-cfn6x                                             1/1     Running   0          6h
kube-system       kube-scheduler-target-cluster-control-plane-qkgtx            1/1     Running   0          6h
tigera-operator   tigera-operator-8dc4c7cb6-h9wbj                              1/1     Running   0          2m18s
```

## Initialize the `target-cluster` with `capi` and `capg` infrastructure provider components

```bash
$  kubectl taint node target-cluster-control-plane-bd6gq node-role.kubernetes.io/master- --context target-cluster --request-timeout 10s
node/target-cluster-control-plane-qkgtx untainted

$ airshipctl phase run clusterctl-init-target --debug --wait-timeout 300s

$ kubectl get pods -A --context target-cluster
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
calico-system                       calico-kube-controllers-55cc6844cb-h4gzh                         1/1     Running   0          10m
calico-system                       calico-node-qdjsm                                                1/1     Running   1          10m
calico-system                       calico-typha-667c57fb6b-kjpfz                                    1/1     Running   0          10m
capg-system                         capg-controller-manager-69c6c9f5d6-wc7mw                         2/2     Running   0          2m39s
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-58bc7fcf9b-v9w24       2/2     Running   0          2m46s
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-867bc8f784-4t7ck   2/2     Running   0          2m42s
capi-system                         capi-controller-manager-78b7d8b9b8-69nwp                         2/2     Running   0          2m51s
capi-webhook-system                 capg-controller-manager-55bb898db6-g6nlw                         2/2     Running   0          2m41s
capi-webhook-system                 capi-controller-manager-7b7c9f89d9-5nh75                         2/2     Running   0          2m53s
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-699b84775f-prwn5       2/2     Running   0          2m49s
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-b8b48d45f-bcvq4    2/2     Running   0          2m45s
cert-manager                        cert-manager-cainjector-55d9fb4b8-fk5z8                          1/1     Running   0          10m
cert-manager                        cert-manager-dfbc75865-mfjz9                                     1/1     Running   0          10m
cert-manager                        cert-manager-webhook-66fc9cf7c-fbgx4                             1/1     Running   0          10m
kube-system                         coredns-6955765f44-pl4zv                                         1/1     Running   0          6h9m
kube-system                         coredns-6955765f44-wwkxt                                         1/1     Running   0          6h9m
kube-system                         etcd-target-cluster-control-plane-qkgtx                          1/1     Running   0          6h9m
kube-system                         kube-apiserver-target-cluster-control-plane-qkgtx                1/1     Running   0          6h9m
kube-system                         kube-controller-manager-target-cluster-control-plane-qkgtx       1/1     Running   0          6h9m
kube-system                         kube-proxy-cfn6x                                                 1/1     Running   0          6h9m
kube-system                         kube-scheduler-target-cluster-control-plane-qkgtx                1/1     Running   0          6h9m
tigera-operator                     tigera-operator-8dc4c7cb6-h9wbj                                  1/1     Running   0          10m
```

## Perform cluster move operation

$ airshipctl phase run clusterctl-move --debug

Check that machines have moved

```bash
$ kubectl get machines --context target-cluster
NAME                                 PROVIDERID                                                        PHASE
target-cluster-control-plane-pbf4n   gce://airship-gcp/us-east1-b/target-cluster-control-plane-qkgtx   Provisioned
```

At this point, the ephemeral-cluster can be deleted.
$ kind delete cluster --name "ephemeral-cluster"

## Deploy worker machines in the `target-cluster`

$ airshipctl phase run  workers-target --debug

Now, the control plane and worker node are created on google cloud.

Check machine status

$ kubectl get machines --context target-cluster
NAME                                   PROVIDERID                                                        PHASE
target-cluster-control-plane-pbf4n     gce://airship-gcp/us-east1-b/target-cluster-control-plane-qkgtx   Running
target-cluster-md-0-7bffdbfd9f-dqrf7   gce://airship-gcp/us-east1-b/target-cluster-md-0-7jtz5            Running

![Machines](https://i.imgur.com/XwAOoar.png)

## Tear Down Cluster

```bash
$ airshipctl phase render controlplane-ephemeral -k Cluster | kubectl --context target-cluster delete -f -

cluster.cluster.x-k8s.io "target-cluster" deleted
```

## Reference

### Pre-requisites

* Install [Docker](https://www.docker.com/)
* Install [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* Install [Kind](https://kind.sigs.k8s.io/)

Also, check [Software Version Information](#Software-Version-Information),
[Special Instructions](#Special-Instructions) and [Virtual Machine
Specification](#Virtual-Machine-Specification)

### Provider Manifests

Provider Configuration is referenced from https://github.com/kubernetes-sigs/cluster-api-provider-gcp/tree/master/config

#### CAPG Specific Variables

capg-resources.yaml consists of `gcp provider specific` variables required to
initialize the management cluster. The values for these variables can be
exported before running `airshipctl phase run clusterctl-init-ephemeral` or they can be defined
explicitly in clusterctl.yaml

$ cat  airshipctl/manifests/function/capg/v0.3.0/data/capg-resources.yaml

```
apiVersion: v1
kind: Secret
metadata:
  name: manager-bootstrap-credentials
  namespace: system
type: Opaque
data:
  GCP_CONTROL_PLANE_MACHINE_TYPE: ${GCP_CONTROL_PLANE_MACHINE_TYPE}
  GCP_NODE_MACHINE_TYPE: ${GCP_NODE_MACHINE_TYPE}
  GCP_PROJECT: ${GCP_PROJECT}
  GCP_REGION: ${GCP_REGION}
  GCP_NETWORK_NAME: ${GCP_NETWORK_NAME}
  GCP_B64ENCODED_CREDENTIALS: ${GCP_B64ENCODED_CREDENTIALS}

```

### Software Version Information

All the instructions provided in the document have been tested using the
software and version, provided in this section.

#### Virtual Machine Specification

All the instructions in the document were perfomed on a Oracle Virtual Box(6.1)
VM running Ubuntu 18.04.4 LTS (Bionic Beaver) with 16G of memory and 4 VCPUs

#### Docker

$ docker version

```
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

$ kind version

```bash
kind v0.9.0 go1.15.2 linux/amd64
```

#### Kubectl

$ kubectl version

```
Client Version: version.Info{Major:"1", Minor:"17", GitVersion:"v1.17.4", GitCommit:"8d8aa39598534325ad77120c120a22b3a990b5ea", GitTreeState:"clean", BuildDate:"2020-03-12T21:03:42Z", GoVersion:"go1.13.8", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"17", GitVersion:"v1.17.0", GitCommit:"70132b0f130acc0bed193d9ba59dd186f0e634cf", GitTreeState:"clean", BuildDate:"2020-01-14T00:09:19Z", GoVersion:"go1.13.4", Compiler:"gc", Platform:"linux/amd64"}
```

#### Go

$ go version

```
go version go1.14.1 linux/amd64
```

#### Kustomize

$ kustomize version

```
{Version:kustomize/v3.8.0 GitCommit:6a50372dd5686df22750b0c729adaf369fbf193c BuildDate:2020-07-05T14:08:42Z GoOs:linux GoArch:amd64}
```

#### OS

$ cat /etc/os-release

```
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