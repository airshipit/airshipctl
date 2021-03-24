# Airshipctl and Cluster API GCP Provider Integration

## Overview

Airshipctl and cluster api gcp integration facilitates usage of `airshipctl` to
create cluster api management and workload clusters using `gcp as infrastructure
provider`.

![Machines](https://i.imgur.com/UfxDtNO.jpg)

## Workflow

A simple workflow that can be tested, involves the following
operations:

**Initialize the management cluster with cluster api and gcp provider
components**

> airshipctl phase run clusterctl-init-ephemeral

**Create a workload cluster, with control plane and worker nodes**

> airshipctl phase run controlplane-ephemeral

> airshipctl phase run workers-target

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

## Other Common Pre-requisites

These prerequistes are required on the VM  that will be used to create workload
cluster on gcp

* Install [Docker](https://www.docker.com/)
* Install [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* Install [Kind](https://kind.sigs.k8s.io/)
* Install
  [Kustomize](https://kubernetes-sigs.github.io/kustomize/installation/binaries/)
* Install [Airshipctl](https://docs.airshipit.org/airshipctl/developers.html)

Also, check [Software Version Information](#Software-Version-Information),
[Special Instructions](#Special-Instructions) and [Virtual Machine
Specification](#Virtual-Machine-Specification)

## Getting Started

Kind will be used to setup a kubernetes cluster, that will be later transformed
into a management cluster using airshipctl. The kind kubernetes cluster will be
initialized with cluster API and Cluster API gcp provider components.

$ export KIND_EXPERIMENTAL_DOCKER_NETWORK=bridge

$ export KUBECONFIG=${KUBECONFIG:-"$HOME/.airship/kubeconfig"}

$ kind create cluster --name ephemeral-cluster --wait 200s

```
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

```
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

```
apiVersion: airshipit.org/v1alpha1
contexts:
  default:
    managementConfiguration: default
    manifest: default
  ephemeral-cluster:
    managementConfiguration: ""
    manifest: gcp_manifest
currentContext: ephemeral-cluster
encryptionConfigs: null
kind: Config
managementConfiguration:
  default:
    systemActionRetries: 30
    systemRebootDelay: 30
    type: redfish
manifests:
  default:
    metadataPath: manifests/site/test-site/metadata.yaml
    phaseRepositoryName: primary
    repositories:
      primary:
        checkout:
          branch: master
          commitHash: ""
          force: false
          tag: ""
        url: https://opendev.org/airship/treasuremap
    targetPath: /tmp/default
  gcp_manifest:
    metadataPath: manifests/site/gcp-test-site/metadata.yaml
    phaseRepositoryName: primary
    repositories:
      primary:
        checkout:
          branch: master
          commitHash: ""
          force: false
          tag: ""
        url: https://opendev.org/airship/airshipctl
    targetPath: /tmp/airship
permissions:
  DirectoryPermission: 488
  FilePermission: 416
```

$ kubectl config set-context ephemeral-cluster --cluster kind-ephemeral-cluster --user kind-ephemeral-cluster
Context "ephemeral-cluster" modified.

$ airshipctl document pull --debug

```
[airshipctl] 2020/08/12 14:07:13 Reading current context manifest information from /home/rishabh/.airship/config
[airshipctl] 2020/08/12 14:07:13 Downloading primary repository airshipctl from https://review.opendev.org/airship/airshipctl into /tmp/airship
[airshipctl] 2020/08/12 14:07:13 Attempting to download the repository airshipctl
[airshipctl] 2020/08/12 14:07:13 Attempting to clone the repository airshipctl from https://review.opendev.org/airship/airshipctl
[airshipctl] 2020/08/12 14:07:23 Attempting to checkout the repository airshipctl from branch refs/heads/master
```

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

$ airshipctl phase run clusterctl-init-ephemeral --debug --kubeconfig ~/.airship/kubeconfig

```
[airshipctl] 2021/02/17 20:29:26 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:109: Verifying that variable CONTAINER_CAPD_AUTH_PROXY is allowed to be appended
[airshipctl] 2021/02/17 20:29:26 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:109: Verifying that variable CONTAINER_CAPD_MANAGER is allowed to be appended
[airshipctl] 2021/02/17 20:29:26 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:109: Verifying that variable CONTAINER_CAPO_AUTH_PROXY is allowed to be appended
[airshipctl] 2021/02/17 20:29:26 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:109: Verifying that variable CONTAINER_CAPO_MANAGER is allowed to be appended
[airshipctl] 2021/02/17 20:29:26 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:109: Verifying that variable CONTAINER_CAPZ_AUTH_PROXY is allowed to be appended
[airshipctl] 2021/02/17 20:29:26 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:109: Verifying that variable CONTAINER_CAPZ_MANAGER is allowed to be appended
[airshipctl] 2021/02/17 20:29:26 opendev.org/airship/airshipctl@/pkg/clusterctl/client/client.go:81: Starting cluster-api initiation
.
.
.
Patching Secret="capg-manager-bootstrap-credentials" Namespace="capg-system"
Creating Service="capg-controller-manager-metrics-service" Namespace="capg-system"
Creating Deployment="capg-controller-manager" Namespace="capg-system"
Creating inventory entry Provider="infrastructure-gcp" Version="v0.3.0" TargetNamespace="capg-system"
{"Message":"clusterctl init completed successfully","Operation":"ClusterctlInitEnd","Timestamp":"2021-02-17T20:31:10.081293629Z","Type":"ClusterctlEvent"}
```

$ kubectl get pods -A
```
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
capg-system                         capg-controller-manager-696f4fb4f-vbr8k                          2/2     Running   0          92s
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-6f669ccd7c-d59t9       2/2     Running   0          110s
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-5c95f59c5c-ptc2j   2/2     Running   0          104s
capi-system                         capi-controller-manager-5f677d7d65-xp6gj                         2/2     Running   0          2m3s
capi-webhook-system                 capg-controller-manager-6798d58795-5scrs                         2/2     Running   0          95s
capi-webhook-system                 capi-controller-manager-745689557d-8mqhq                         2/2     Running   0          2m6s
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-6949f44db8-lc8lk       2/2     Running   0          118s
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-7b6c4bf48d-997p9   2/2     Running   0          109s
cert-manager                        cert-manager-cainjector-fc6c787db-49jjz                          1/1     Running   0          2m30s
cert-manager                        cert-manager-d994d94d7-7lmgz                                     1/1     Running   0          2m30s
cert-manager                        cert-manager-webhook-845d9df8bf-nl8qd                            1/1     Running   0          2m30s
kube-system                         coredns-f9fd979d6-g8wrd                                          1/1     Running   0          74m
kube-system                         coredns-f9fd979d6-wrc5r                                          1/1     Running   0          74m
kube-system                         etcd-ephemeral-cluster-control-plane                             1/1     Running   0          75m
kube-system                         kindnet-p8bx7                                                    1/1     Running   0          74m
kube-system                         kube-apiserver-ephemeral-cluster-control-plane                   1/1     Running   0          75m
kube-system                         kube-controller-manager-ephemeral-cluster-control-plane          1/1     Running   0          75m
kube-system                         kube-proxy-zl7jg                                                 1/1     Running   0          74m
kube-system                         kube-scheduler-ephemeral-cluster-control-plane                   1/1     Running   0          75m
local-path-storage                  local-path-provisioner-78776bfc44-q7gtr                          1/1     Running   0          74m
```

## Deploy control plane nodes in the `target-cluster`

$ airshipctl phase run controlplane-ephemeral --debug --kubeconfig ~/.airship/kubeconfig

```bash

[airshipctl] 2021/02/17 20:34:30 opendev.org/airship/airshipctl@/pkg/phase/executors/k8s_applier.go:130: Getting kubeconfig context name from cluster map
[airshipctl] 2021/02/17 20:34:30 opendev.org/airship/airshipctl@/pkg/phase/executors/k8s_applier.go:135: Getting kubeconfig file information from kubeconfig provider
[airshipctl] 2021/02/17 20:34:30 opendev.org/airship/airshipctl@/pkg/phase/executors/k8s_applier.go:140: Filtering out documents that shouldn't be applied to kubernetes from document bundle
[airshipctl] 2021/02/17 20:34:30 opendev.org/airship/airshipctl@/pkg/phase/executors/k8s_applier.go:148: Using kubeconfig at '/home/stack/.airship/kubeconfig' and context 'ephemeral-cluster'
[airshipctl] 2021/02/17 20:34:30 opendev.org/airship/airshipctl@/pkg/phase/executors/k8s_applier.go:119: WaitTimeout: 33m20s
[airshipctl] 2021/02/17 20:34:30 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:76: Getting infos for bundle, inventory id is controlplane-ephemeral
[airshipctl] 2021/02/17 20:34:30 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:106: Inventory Object config Map not found, auto generating Inventory object
[airshipctl] 2021/02/17 20:34:30 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:113: Injecting Inventory Object: {"apiVersion":"v1","kind":"ConfigMap","metadata":{"creationTimestamp":null,"labels":{"cli-utils.sigs.k8s.io/inventory-id":"controlplane-ephemeral"},"name":"airshipit-controlplane-ephemeral","namespace":"airshipit"}}{nsfx:false,beh:unspecified} into bundle
[airshipctl] 2021/02/17 20:34:30 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:119: Making sure that inventory object namespace airshipit exists
cluster.cluster.x-k8s.io/target-cluster created
gcpcluster.infrastructure.cluster.x-k8s.io/target-cluster created
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/target-cluster-control-plane created
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/target-cluster-control-plane created
4 resource(s) applied. 4 created, 0 unchanged, 0 configured
cluster.cluster.x-k8s.io/target-cluster is NotFound: Resource not found
gcpcluster.infrastructure.cluster.x-k8s.io/target-cluster is NotFound: Resource not found
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/target-cluster-control-plane is NotFound: Resource not found
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/target-cluster-control-plane is NotFound: Resource not found
cluster.cluster.x-k8s.io/target-cluster is InProgress:
gcpcluster.infrastructure.cluster.x-k8s.io/target-cluster is Current: Resource is current
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/target-cluster-control-plane is Current: Resource is current
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/target-cluster-control-plane is Current: Resource is current
cluster.cluster.x-k8s.io/target-cluster is InProgress:
gcpcluster.infrastructure.cluster.x-k8s.io/target-cluster is Current: Resource is current
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/target-cluster-control-plane is InProgress:
cluster.cluster.x-k8s.io/target-cluster is InProgress: 0 of 1 completed
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/target-cluster-control-plane is InProgress: 0 of 1 completed
cluster.cluster.x-k8s.io/target-cluster is InProgress:
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/target-cluster-control-plane is InProgress:
cluster.cluster.x-k8s.io/target-cluster is Current: Resource is Ready
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/target-cluster-control-plane is Current: Resource is Ready
all resources has reached the Current status
```

To check logs run the below command

$  kubectl logs capg-controller-manager-696f4fb4f-vbr8k -n capg-system --all-containers=true -f --kubeconfig ~/.airship/kubeconfig

$ kubectl get machines

```bash
NAME                                 PROVIDERID                                                        PHASE
target-cluster-control-plane-pbf4n   gce://airship-gcp/us-east1-b/target-cluster-control-plane-qkgtx   Running
```

## Deploy Calico cni in the `target-cluster`

```bash
$ kubectl --namespace=default get secret/target-cluster-kubeconfig -o jsonpath={.data.value} | base64 --decode > ./target-cluster.kubeconfig

$ kubectl --namespace=default get secret/target-cluster-kubeconfig -o jsonpath={.data.value} | base64 --decode > ./target-cluster.kubeconfig

$ kubectl get nodes --kubeconfig target-cluster.kubeconfig
NAME                                 STATUS     ROLES    AGE     VERSION
target-cluster-control-plane-qkgtx   NotReady   master   5h53m   v1.17.11
```

Create target-cluster context

```bash
$ kubectl config set-context target-cluster --user target-cluster-admin --cluster target-cluster --kubeconfig target-cluster.kubeconfig
Context "target-cluster" created.
```

Deploy calico cni using `initinfra-networking` phase

```bash
$ airshipctl phase run initinfra-networking-target --kubeconfig target-cluster.kubeconfig
namespace/cert-manager created
namespace/tigera-operator created
customresourcedefinition.apiextensions.k8s.io/bgpconfigurations.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/bgppeers.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/blockaffinities.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/certificaterequests.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/certificates.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/challenges.acme.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/clusterinformations.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/clusterissuers.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/felixconfigurations.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/globalnetworkpolicies.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/globalnetworksets.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/hostendpoints.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/installations.operator.tigera.io created
customresourcedefinition.apiextensions.k8s.io/ipamblocks.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/ipamconfigs.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/ipamhandles.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/ippools.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/issuers.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/kubecontrollersconfigurations.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/networkpolicies.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/networksets.crd.projectcalico.org created
customresourcedefinition.apiextensions.k8s.io/orders.acme.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/tigerastatuses.operator.tigera.io created
mutatingwebhookconfiguration.admissionregistration.k8s.io/cert-manager-webhook created
serviceaccount/cert-manager created
serviceaccount/cert-manager-cainjector created
serviceaccount/cert-manager-webhook created
serviceaccount/tigera-operator created
podsecuritypolicy.policy/tigera-operator created
role.rbac.authorization.k8s.io/cert-manager-webhook:dynamic-serving created
role.rbac.authorization.k8s.io/cert-manager-cainjector:leaderelection created
role.rbac.authorization.k8s.io/cert-manager:leaderelection created
clusterrole.rbac.authorization.k8s.io/cert-manager-cainjector created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-certificates created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-challenges created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-clusterissuers created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-ingress-shim created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-issuers created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-orders created
clusterrole.rbac.authorization.k8s.io/cert-manager-edit created
clusterrole.rbac.authorization.k8s.io/cert-manager-view created
clusterrole.rbac.authorization.k8s.io/tigera-operator created
rolebinding.rbac.authorization.k8s.io/cert-manager-webhook:dynamic-serving created
rolebinding.rbac.authorization.k8s.io/cert-manager-cainjector:leaderelection created
rolebinding.rbac.authorization.k8s.io/cert-manager:leaderelection created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-cainjector created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-certificates created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-challenges created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-clusterissuers created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-ingress-shim created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-issuers created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-orders created
clusterrolebinding.rbac.authorization.k8s.io/tigera-operator created
service/cert-manager created
service/cert-manager-webhook created
deployment.apps/cert-manager created
deployment.apps/cert-manager-cainjector created
deployment.apps/cert-manager-webhook created
deployment.apps/tigera-operator created
installation.operator.tigera.io/default created
validatingwebhookconfiguration.admissionregistration.k8s.io/cert-manager-webhook created
63 resource(s) applied. 63 created, 0 unchanged, 0 configured
```

Check on control plane node status

$ kubectl get nodes --kubeconfig target-cluster.kubeconfig
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
$  kubectl taint node target-cluster-control-plane-bd6gq node-role.kubernetes.io/master- --kubeconfig target-cluster.kubeconfig --request-timeout 10s
node/target-cluster-control-plane-qkgtx untainted

$ airshipctl phase run clusterctl-init-target --debug --kubeconfig target-cluster.kubeconfig

$ kubectl get pods -A --kubeconfig target-cluster.kubeconfig
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

```bash
$ KUBECONFIG=~/.airship/kubeconfig:target-cluster.kubeconfig kubectl config view --merge --flatten > ~/ephemeral_and_target.kubeconfig

$ airshipctl phase run clusterctl-move --kubeconfig ~/ephemeral_and_target.kubeconfig
[airshipctl] 2021/02/18 02:50:32 command 'clusterctl move' is going to be executed
{"Message":"starting clusterctl move executor","Operation":"ClusterctlMoveStart","Timestamp":"2021-02-18T02:50:32.758374205Z","Type":"ClusterctlEvent"}
{"Message":"clusterctl move completed successfully","Operation":"ClusterctlMoveEnd","Timestamp":"2021-02-18T02:50:36.823224336Z","Type":"ClusterctlEvent"}
```

Check that machines have moved

```bash

$ kubectl get machines --kubeconfig ~/.airship/kubeconfig
No resources found in default namespace.

$ kubectl get machines --kubeconfig ~/target-cluster.kubeconfig
NAME                                 PROVIDERID                                                        PHASE
target-cluster-control-plane-pbf4n   gce://airship-gcp/us-east1-b/target-cluster-control-plane-qkgtx   Provisioned
```

## Deploy worker machines in the `target-cluster`

```bash

$ airshipctl phase run workers-target --debug --kubeconfig ~/target-cluster.kubeconfig
[airshipctl] 2021/02/18 02:56:22 opendev.org/airship/airshipctl@/pkg/phase/executors/k8s_applier.go:130: Getting kubeconfig context name from cluster map
[airshipctl] 2021/02/18 02:56:22 opendev.org/airship/airshipctl@/pkg/phase/executors/k8s_applier.go:135: Getting kubeconfig file information from kubeconfig provider
[airshipctl] 2021/02/18 02:56:22 opendev.org/airship/airshipctl@/pkg/phase/executors/k8s_applier.go:140: Filtering out documents that shouldn't be applied to kubernetes from document bundle
[airshipctl] 2021/02/18 02:56:22 opendev.org/airship/airshipctl@/pkg/phase/executors/k8s_applier.go:148: Using kubeconfig at '/home/stack/target-cluster.kubeconfig' and context 'target-cluster'
[airshipctl] 2021/02/18 02:56:22 opendev.org/airship/airshipctl@/pkg/phase/executors/k8s_applier.go:119: WaitTimeout: 33m20s
[airshipctl] 2021/02/18 02:56:22 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:76: Getting infos for bundle, inventory id is workers-target
[airshipctl] 2021/02/18 02:56:22 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:106: Inventory Object config Map not found, auto generating Inventory object
[airshipctl] 2021/02/18 02:56:22 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:113: Injecting Inventory Object: {"apiVersion":"v1","kind":"ConfigMap","metadata":{"creationTimestamp":null,"labels":{"cli-utils.sigs.k8s.io/inventory-id":"workers-target"},"name":"airshipit-workers-target","namespace":"airshipit"}}{nsfx:false,beh:unspecified} into bundle
[airshipctl] 2021/02/18 02:56:22 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:119: Making sure that inventory object namespace airshipit exists
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/target-cluster-md-0 created
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/target-cluster-md-0 created
machinedeployment.cluster.x-k8s.io/target-cluster-md-0 created
3 resource(s) applied. 3 created, 0 unchanged, 0 configured
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/target-cluster-md-0 is NotFound: Resource not found
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/target-cluster-md-0 is NotFound: Resource not found
machinedeployment.cluster.x-k8s.io/target-cluster-md-0 is NotFound: Resource not found
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/target-cluster-md-0 is Current: Resource is current
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/target-cluster-md-0 is Current: Resource is current
machinedeployment.cluster.x-k8s.io/target-cluster-md-0 is Current: Resource is current
all resources has reached the Current status
```

Now, the control plane and worker node are created on google cloud.

Check machine status

$ kubectl get machines --kubeconfig ~/.airship/kubeconfig
NAME                                   PROVIDERID                                                        PHASE
target-cluster-control-plane-pbf4n     gce://airship-gcp/us-east1-b/target-cluster-control-plane-qkgtx   Running
target-cluster-md-0-7bffdbfd9f-dqrf7   gce://airship-gcp/us-east1-b/target-cluster-md-0-7jtz5            Running

![Machines](https://i.imgur.com/XwAOoar.png)

## Tear Down Cluster

```bash
$ airshipctl phase render controlplane-ephemeral -k Cluster | kubectl
--kubeconfig ~/target-cluster.kubeconfig delete -f -

cluster.cluster.x-k8s.io "target-cluster" deleted
```

```bash
$ kind delete clusters --all

Deleted clusters: ["ephemeral-cluster"]
```

## Reference

### Provider Manifests

Provider Configuration is referenced from https://github.com/kubernetes-sigs/cluster-api-provider-gcp/tree/master/config
Cluster API does not support gcp provider out of the box. Therefore, the metadata infromation is added using files in
airshipctl/manifests/function/capg/data

$ tree airshipctl/manifests/function/capg

```
airshipctl/manifests/function/capg
└── v0.3.0
    ├── certmanager
    │   ├── certificate.yaml
    │   ├── kustomization.yaml
    │   └── kustomizeconfig.yaml
    ├── crd
    │   ├── bases
    │   │   ├── infrastructure.cluster.x-k8s.io_gcpclusters.yaml
    │   │   ├── infrastructure.cluster.x-k8s.io_gcpmachines.yaml
    │   │   └── infrastructure.cluster.x-k8s.io_gcpmachinetemplates.yaml
    │   ├── kustomization.yaml
    │   ├── kustomizeconfig.yaml
    │   └── patches
    │       ├── cainjection_in_gcpclusters.yaml
    │       ├── cainjection_in_gcpmachines.yaml
    │       ├── cainjection_in_gcpmachinetemplates.yaml
    │       ├── webhook_in_gcpclusters.yaml
    │       ├── webhook_in_gcpmachines.yaml
    │       └── webhook_in_gcpmachinetemplates.yaml
    ├── data
    │   ├── capg-resources.yaml
    │   ├── kustomization.yaml
    │   └── metadata.yaml
    ├── default
    │   ├── credentials.yaml
    │   ├── kustomization.yaml
    │   ├── manager_credentials_patch.yaml
    │   ├── manager_prometheus_metrics_patch.yaml
    │   ├── manager_role_aggregation_patch.yaml
    │   └── namespace.yaml
    ├── kustomization.yaml
    ├── manager
    │   ├── kustomization.yaml
    │   ├── manager_auth_proxy_patch.yaml
    │   ├── manager_image_patch.yaml
    │   ├── manager_pull_policy.yaml
    │   └── manager.yaml
    ├── patch_crd_webhook_namespace.yaml
    ├── rbac
    │   ├── auth_proxy_role_binding.yaml
    │   ├── auth_proxy_role.yaml
    │   ├── auth_proxy_service.yaml
    │   ├── kustomization.yaml
    │   ├── leader_election_role_binding.yaml
    │   ├── leader_election_role.yaml
    │   ├── role_binding.yaml
    │   └── role.yaml
    └── webhook
        ├── kustomization.yaml
        ├── kustomizeconfig.yaml
        ├── manager_webhook_patch.yaml
        ├── manifests.yaml
        ├── service.yaml
        └── webhookcainjection_patch.yaml
```

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

### Cluster Templates
manifests/function/k8scontrol-capg contains cluster.yaml, controlplane.yaml templates referenced from
[cluster-template](https://github.com/kubernetes-sigs/cluster-api-provider-gcp/blob/master/templates/cluster-template.yaml)

| Template Name     | CRDs |
| ----------------- | ---- |
| cluster.yaml      |   Cluster, GCPCluster   |
| controlplane.yaml | KubeadmControlPlane, GCPMachineTemplate  |

$ tree airshipctl/manifests/function/k8scontrol-capg

```
airshipctl/manifests/function/k8scontrol-capg
├── cluster.yaml
├── controlplane.yaml
└── kustomization.yaml
```

airshipctl/manifests/function/workers-capg contains workers.yaml referenced from
[cluster-template](https://github.com/kubernetes-sigs/cluster-api-provider-gcp/blob/master/templates/cluster-template.yaml)

| Template Name     | CRDs |
| ----------------- | ---- |
| workers.yaml      |  GCPMachineTemplate, MachineDeployment,  KubeadmConfigTemplate |

$ tree airshipctl/manifests/function/workers-capg
```
airshipctl/manifests/function/workers-capg
├── kustomization.yaml
└── workers.yaml
```

### Test Site Manifests

The `gcp-test-site` contains ephemeral and target phase manifests.

```bash
$ tree gcp-test-site/
gcp-test-site/
├── ephemeral
│   └── controlplane
│       ├── kustomization.yaml
│       ├── machine_count.json
│       ├── machine_type.json
│       ├── network_name.json
│       ├── project_name.json
│       └── region_name.json
├── metadata.yaml
├── phases
│   ├── infrastructure-providers.json
│   ├── kustomization.yaml
│   └── plan.yaml
└── target
    ├── initinfra
    │   └── kustomization.yaml
    ├── initinfra-networking
    │   └── kustomization.yaml
    └── workers
        ├── failure_domain.json
        ├── kustomization.yaml
        ├── machine_count.json
        └── machine_type.json

7 directories, 16 files
```

#### gcp-test-site/target

Following phases are available in the gcp test site phase definitions.

|Phase Name | Purpose |
|-----------|---------|
| clusterctl-init-ephemeral | Initializes the ephemeral cluster with capi and capg components
| controlplane-ephemeral     | Patches templates in manifests/function/k8scontrol-capg and deploys the control plane machines in the target cluster|
| initinfra-networking-target | Deploys calico CNI in the target cluster
| clusterctl-init-target | Initializes target cluster with capi and capg components
| clusterctl-move | Moves management CRDs from ephemeral to target cluster
| workers-target          | Patches template in manifests/function/workers-capg and deploys worker nodes in the target cluster|                                                     |

#### Patch Merge Strategy

Json patches  are applied on templates in `manifests/function/k8scontrol-capg`
from `airshipctl/manifests/site/gcp-test-site/target/controlplane` when
`airshipctl phase run controlplane-target` is executed

Json patches are applied on templates in `manifests/function/workers-capg` from
`airshipctl/manifests/site/gcp-test-site/target/workers` when `airshipctl phase
run workers-target` is executed.

| Patch Name                      | Purpose                                                            |
| ------------------------------- | ------------------------------------------------------------------ |
| controlplane/machine_count.json | patches control plane machine count in template function/k8scontrol-capg |
| controlplane/machine_type.json | patches control plane machine type in template function/k8scontrol-capg |
| controlplane/network_name.json | patches control plane network name in template function/k8scontrol-capg |
| controlplane/project_name.json | patches project id template function/k8scontrol-capg |
| controlplane/region_name.json | patches region name in template function/k8scontrol-capg |
| workers/machine_count.json      | patches worker machine count in template function/workers-capg |
| workers/machine_type.json      | patches worker machine type in template function/workers-capg |
| workers/failure_domain.json      | patches failure_domain in template function/workers-capg |

$ tree airshipctl/manifests/site/gcp-test-site/ephemeral/
gcp-test-site/ephemeral/
└── controlplane
    ├── kustomization.yaml
    ├── machine_count.json
    ├── machine_type.json
    ├── network_name.json
    ├── project_name.json
    └── region_name.json


$ tree airshipctl/manifests/site/gcp-test-site/target/

```bash
airshipctl/manifests/site/gcp-test-site/target/
gcp-test-site/target/
├── initinfra
│   └── kustomization.yaml
├── initinfra-networking
│   └── kustomization.yaml
└── workers
    ├── failure_domain.json
    ├── kustomization.yaml
    ├── machine_count.json
    └── machine_type.json

3 directories, 6 files

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