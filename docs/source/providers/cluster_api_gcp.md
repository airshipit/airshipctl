# Airshipctl and Cluster API GCP Integration

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

> airshipctl phase run controlplane-target

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

$ make build-gce-default

$ gcloud compute images list --project ${GCP_PROJECT_ID} --no-standard-images

```
NAME                                         PROJECT                FAMILY                      DEPRECATED  STATUS
cluster-api-ubuntu-1804-v1-16-14-1599066516  virtual-anchor-281401  capi-ubuntu-1804-k8s-v1-16              READY
```

### Create Cloud NAT Router

Kubernetes nodes, to communicate with the control plane, pull container images
from registried (e.g. gcr.io or dockerhub) need to have NAT access or a public
ip. By default, the provider creates Machines without a public IP.

To make sure your cluster can communicate with the outside world, and the load
balancer, you can create a Cloud NAT in the region you'd like your Kubernetes
cluster to live in by following [these
instructions](https://cloud.google.com/nat/docs/using-nat#specify_ip_addresses_for_nat).

For reference, use the below images. You can create 2 cloud NAT routers for
region us-west1 and us-east1

![us-west1](https://i.imgur.com/Q5DRxtV.jpg)

![us-east1](https://i.imgur.com/94qeAch.jpg)

![nat-routers](https://i.imgur.com/wbeBSyF.jpg)


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

$ kind create cluster --name capi-gcp
```
Creating cluster "capi-gcp" ...
WARNING: Overriding docker network due to KIND_EXPERIMENTAL_DOCKER_NETWORK
WARNING: Here be dragons! This is not supported currently.
 ✓ Ensuring node image (kindest/node:v1.18.2) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-capi-gcp"
You can now use your cluster with:

kubectl cluster-info --context kind-capi-gcp
```

$ kubectl get pods -A

```
NAMESPACE            NAME                                             READY   STATUS    RESTARTS   AGE
kube-system          coredns-66bff467f8-kmg7c                         1/1     Running   0          82s
kube-system          coredns-66bff467f8-lg8qc                         1/1     Running   0          82s
kube-system          etcd-capi-gcp-control-plane                      1/1     Running   0          91s
kube-system          kindnet-dzp8v                                    1/1     Running   0          82s
kube-system          kube-apiserver-capi-gcp-control-plane            1/1     Running   0          91s
kube-system          kube-controller-manager-capi-gcp-control-plane   1/1     Running   0          90s
kube-system          kube-proxy-zvdh8                                 1/1     Running   0          82s
kube-system          kube-scheduler-capi-gcp-control-plane            1/1     Running   0          83s
local-path-storage   local-path-provisioner-bd4bb6b75-6drnt           1/1     Running   0          82s
```

## Create airshipctl configuration files

$ mkdir ~/.airship

$ airshipctl config init

Run the below command to configure gcp manifest, and add it to airship config

```
$ airshipctl config set-manifest gcp_manifest --repo primary \
--url https://opendev.org/airship/airshipctl --branch master \
--primary --sub-path manifests/site/gcp-test-site --target-path /tmp/airship
```

$ airshipctl config set-context kind-capi-gcp --manifest gcp_manifest

```
Context "kind-capi-gcp" modified.
```
$ cp ~/.kube/config ~/.airship/kubeconfig

$ airshipctl config get-context

```
Context: kind-capi-gcp
contextKubeconf: kind-capi-gcp_target
manifest: gcp_manifest

LocationOfOrigin: /home/rishabh/.airship/kubeconfig
cluster: kind-capi-gcp_target
user: kind-capi-gcp
```
$ airshipctl config use-context kind-capi-gcp

```
Manifest "gcp_manifest" created.
```

$ airshipctl document pull --debug

```
[airshipctl] 2020/08/12 14:07:13 Reading current context manifest information from /home/rishabh/.airship/config
[airshipctl] 2020/08/12 14:07:13 Downloading primary repository airshipctl from https://review.opendev.org/airship/airshipctl into /tmp/airship
[airshipctl] 2020/08/12 14:07:13 Attempting to download the repository airshipctl
[airshipctl] 2020/08/12 14:07:13 Attempting to clone the repository airshipctl from https://review.opendev.org/airship/airshipctl
[airshipctl] 2020/08/12 14:07:23 Attempting to checkout the repository airshipctl from branch refs/heads/master
```
$ airshipctl config set-manifest gcp_manifest --target-path /tmp/airship/airshipctl

## Configure gcp site variables

`configure project_id`

$ cat /tmp/airship/airshipctl/manifests/site/gcp-test-site/target/controlplane/project_name.json

```
[
  { "op": "replace","path": "/spec/project","value": "<project_id>"}
]
```

Include gcp variables in clusterctl.yaml

The original values for the below variables are as follows:
```
GCP_CONTROL_PLANE_MACHINE_TYPE="n1-standard-4"
GCP_NODE_MACHINE_TYPE="n1-standard-4"
GCP_REGION="us-west1"
GCP_NETWORK_NAME="default"

GCP_PROJECT="<your_project_id>"
GCP_CREDENTIALS="$( cat ~/</path/to/serviceaccount-key.json>)"
```

Edit `airshipctl/manifests/site/gcp-test-site/shared/clusterctl/clusterctl.yaml`
to include gcp variables and their values in base64 encoded format. Use
https://www.base64decode.org/ if required.

To get the GCP_CREDENTIALS in base64 format, use the below command.

$ export GCP_B64ENCODED_CREDENTIALS=$( cat ~/</path/to/serviceaccount-key.json> | base64 | tr -d '\n' )

$ echo $GCP_B64ENCODED_CREDENTIALS

The below shown `clusterctl.yaml`, has encoded the values for all variables except
GCP_PROJECT and GCP_CREDENTIALS. You can use the base64 encoded values for
GCP_PROJECT and GCP_CREDENTIALS based on your project.

The other remaining variables in the `clusterctl.yaml` are base64 encoded.
Their original values is as follows:

```
GCP_CONTROL_PLANE_MACHINE_TYPE="n1-standard-4"
GCP_NODE_MACHINE_TYPE="n1-standard-4"
GCP_REGION="us-west1"
GCP_NETWORK_NAME="default"
```

$ cat /tmp/airship/airshipctl/manifests/site/gcp-test-site/shared/clusterctl/clusterctl.yaml

```
apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  labels:
    airshipit.org/deploy-k8s: "false"
  name: clusterctl-v1
init-options:
  core-provider: "cluster-api:v0.3.3"
  bootstrap-providers:
    - "kubeadm:v0.3.3"
  infrastructure-providers:
    - "gcp:v0.3.0"
  control-plane-providers:
    - "kubeadm:v0.3.3"
providers:
  - name: "gcp"
    type: "InfrastructureProvider"
    variable-substitution: true
    versions:
      v0.3.0: manifests/function/capg/v0.3.0
  - name: "kubeadm"
    type: "BootstrapProvider"
    versions:
      v0.3.3: manifests/function/cabpk/v0.3.3
  - name: "cluster-api"
    type: "CoreProvider"
    versions:
      v0.3.3: manifests/function/capi/v0.3.3
  - name: "kubeadm"
    type: "ControlPlaneProvider"
    versions:
      v0.3.3: manifests/function/cacpk/v0.3.3
additional-vars:
  GCP_CONTROL_PLANE_MACHINE_TYPE: "bjEtc3RhbmRhcmQtNA=="
  GCP_NODE_MACHINE_TYPE: "bjEtc3RhbmRhcmQtNA=="
  GCP_PROJECT: "<B64ENCODED_GCP_PROJECT_ID>"
  GCP_REGION: "dXMtd2VzdDE="
  GCP_NETWORK_NAME: "ZGVmYXVsdA=="
  GCP_B64ENCODED_CREDENTIALS: "<GCP_B64ENCODED_CREDENTIALS>"
```

## Initialize Management Cluster

$ airshipctl phase run clusterctl-init-ephemeral

```
[airshipctl] 2020/09/02 11:14:15 Verifying that variable GCP_REGION is allowed to be appended
[airshipctl] 2020/09/02 11:14:15 Verifying that variable GCP_B64ENCODED_CREDENTIALS is allowed to be appended
[airshipctl] 2020/09/02 11:14:15 Verifying that variable GCP_CONTROL_PLANE_MACHINE_TYPE is allowed to be appended
[airshipctl] 2020/09/02 11:14:15 Verifying that variable GCP_NETWORK_NAME is allowed to be appended
[airshipctl] 2020/09/02 11:14:15 Verifying that variable GCP_NODE_MACHINE_TYPE is allowed to be appended
.
.
.
Patching Secret="capg-manager-bootstrap-credentials" Namespace="capg-system"
Creating Service="capg-controller-manager-metrics-service" Namespace="capg-system"
Creating Deployment="capg-controller-manager" Namespace="capg-system"
Creating inventory entry Provider="infrastructure-gcp" Version="v0.3.0" TargetNamespace="capg-system"
```

$ kubectl get pods -A
```
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
capg-system                         capg-controller-manager-b8655ddb4-swwzk                          2/2     Running   0          54s
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-66c6b6857b-22hg4       2/2     Running   0          73s
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-688f7ccc56-7g676   2/2     Running   0          65s
capi-system                         capi-controller-manager-549c757797-6vscq                         2/2     Running   0          84s
capi-webhook-system                 capg-controller-manager-d5f85c48d-74gj6                          2/2     Running   0          61s
capi-webhook-system                 capi-controller-manager-5f8fc485bb-stflj                         2/2     Running   0          88s
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-6b645d9d4c-2crk7       2/2     Running   0          81s
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-65dbd6f999-cghmx   2/2     Running   0          70s
cert-manager                        cert-manager-77d8f4d85f-cqp7m                                    1/1     Running   0          115s
cert-manager                        cert-manager-cainjector-75f88c9f56-qh9m8                         1/1     Running   0          115s
cert-manager                        cert-manager-webhook-56669d7fcb-6zddl                            1/1     Running   0          115s
kube-system                         coredns-66bff467f8-kmg7c                                         1/1     Running   0          3m55s
kube-system                         coredns-66bff467f8-lg8qc                                         1/1     Running   0          3m55s
kube-system                         etcd-capi-gcp-control-plane                                      1/1     Running   0          4m4s
kube-system                         kindnet-dzp8v                                                    1/1     Running   0          3m55s
kube-system                         kube-apiserver-capi-gcp-control-plane                            1/1     Running   0          4m4s
kube-system                         kube-controller-manager-capi-gcp-control-plane                   1/1     Running   0          4m3s
kube-system                         kube-proxy-zvdh8                                                 1/1     Running   0          3m55s
kube-system                         kube-scheduler-capi-gcp-control-plane                            1/1     Running   0          3m56s
local-path-storage                  local-path-provisioner-bd4bb6b75-6drnt                           1/1     Running   0          3m55s
```

## Create control plane and worker nodes

$ airshipctl phase run controlplane-target --debug
```
[airshipctl] 2020/09/02 11:21:08 building bundle from kustomize path /tmp/airship/airshipctl/manifests/site/gcp-test-site/target/controlplane
[airshipctl] 2020/09/02 11:21:08 Applying bundle, inventory id: kind-capi-gcp-target-controlplane
[airshipctl] 2020/09/02 11:21:08 Inventory Object config Map not found, auto generating Invetory object
[airshipctl] 2020/09/02 11:21:08 Injecting Invetory Object: {"apiVersion":"v1","kind":"ConfigMap","metadata":{"creationTimestamp":null,"labels":{"cli-utils.sigs.k8s.io/inventory-id":"kind-capi-gcp-target-controlplane"},"name":"airshipit-kind-capi-gcp-target-controlplane","namespace":"airshipit"}}{nsfx:false,beh:unspecified} into bundle
[airshipctl] 2020/09/02 11:21:08 Making sure that inventory object namespace airshipit exists
configmap/airshipit-kind-capi-gcp-target-controlplane-5ab3466f created
cluster.cluster.x-k8s.io/gtc created
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/gtc-control-plane created
gcpcluster.infrastructure.cluster.x-k8s.io/gtc created
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/gtc-control-plane created
5 resource(s) applied. 5 created, 0 unchanged, 0 configured
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/gtc-control-plane is NotFound: Resource not found
gcpcluster.infrastructure.cluster.x-k8s.io/gtc is NotFound: Resource not found
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/gtc-control-plane is NotFound: Resource not found
configmap/airshipit-kind-capi-gcp-target-controlplane-5ab3466f is NotFound: Resource not found
cluster.cluster.x-k8s.io/gtc is NotFound: Resource not found
configmap/airshipit-kind-capi-gcp-target-controlplane-5ab3466f is Current: Resource is always ready
cluster.cluster.x-k8s.io/gtc is Current: Resource is current
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/gtc-control-plane is Current: Resource is current
gcpcluster.infrastructure.cluster.x-k8s.io/gtc is Current: Resource is current
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/gtc-control-plane is Current: Resource is current
all resources has reached the Current status
```
$ airshipctl phase run workers-target --debug
```
[airshipctl] 2020/09/02 11:21:20 building bundle from kustomize path /tmp/airship/airshipctl/manifests/site/gcp-test-site/target/workers
[airshipctl] 2020/09/02 11:21:20 Applying bundle, inventory id: kind-capi-gcp-target-workers
[airshipctl] 2020/09/02 11:21:20 Inventory Object config Map not found, auto generating Invetory object
[airshipctl] 2020/09/02 11:21:20 Injecting Invetory Object: {"apiVersion":"v1","kind":"ConfigMap","metadata":{"creationTimestamp":null,"labels":{"cli-utils.sigs.k8s.io/inventory-id":"kind-capi-gcp-target-workers"},"name":"airshipit-kind-capi-gcp-target-workers","namespace":"airshipit"}}{nsfx:false,beh:unspecified} into bundle
[airshipctl] 2020/09/02 11:21:20 Making sure that inventory object namespace airshipit exists
configmap/airshipit-kind-capi-gcp-target-workers-1a36e40a created
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/gtc-md-0 created
machinedeployment.cluster.x-k8s.io/gtc-md-0 created
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/gtc-md-0 created
4 resource(s) applied. 4 created, 0 unchanged, 0 configured
configmap/airshipit-kind-capi-gcp-target-workers-1a36e40a is NotFound: Resource not found
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/gtc-md-0 is NotFound: Resource not found
machinedeployment.cluster.x-k8s.io/gtc-md-0 is NotFound: Resource not found
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/gtc-md-0 is NotFound: Resource not found
configmap/airshipit-kind-capi-gcp-target-workers-1a36e40a is Current: Resource is always ready
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/gtc-md-0 is Current: Resource is current
machinedeployment.cluster.x-k8s.io/gtc-md-0 is Current: Resource is current
gcpmachinetemplate.infrastructure.cluster.x-k8s.io/gtc-md-0 is Current: Resource is current
```

$ kubectl get pods -A
```
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
capg-system                         capg-controller-manager-b8655ddb4-swwzk                          2/2     Running   0          6m9s
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-66c6b6857b-22hg4       2/2     Running   0          6m28s
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-688f7ccc56-7g676   2/2     Running   0          6m20s
capi-system                         capi-controller-manager-549c757797-6vscq                         2/2     Running   0          6m39s
capi-webhook-system                 capg-controller-manager-d5f85c48d-74gj6                          2/2     Running   0          6m16s
capi-webhook-system                 capi-controller-manager-5f8fc485bb-stflj                         2/2     Running   0          6m43s
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-6b645d9d4c-2crk7       2/2     Running   0          6m36s
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-65dbd6f999-cghmx   2/2     Running   0          6m25s
cert-manager                        cert-manager-77d8f4d85f-cqp7m                                    1/1     Running   0          7m10s
cert-manager                        cert-manager-cainjector-75f88c9f56-qh9m8                         1/1     Running   0          7m10s
cert-manager                        cert-manager-webhook-56669d7fcb-6zddl                            1/1     Running   0          7m10s
kube-system                         coredns-66bff467f8-kmg7c                                         1/1     Running   0          9m10s
kube-system                         coredns-66bff467f8-lg8qc                                         1/1     Running   0          9m10s
kube-system                         etcd-capi-gcp-control-plane                                      1/1     Running   0          9m19s
kube-system                         kindnet-dzp8v                                                    1/1     Running   0          9m10s
kube-system                         kube-apiserver-capi-gcp-control-plane                            1/1     Running   0          9m19s
kube-system                         kube-controller-manager-capi-gcp-control-plane                   1/1     Running   0          9m18s
kube-system                         kube-proxy-zvdh8                                                 1/1     Running   0          9m10s
kube-system                         kube-scheduler-capi-gcp-control-plane                            1/1     Running   0          9m11s
local-path-storage                  local-path-provisioner-bd4bb6b75-6drnt                           1/1     Running   0          9m10s
```

To check logs run the below command

$ kubectl logs capg-controller-manager-b8655ddb4-swwzk -n capg-system --all-containers=true -f

```
I0902 18:15:30.884391       1 main.go:213] Generating self signed cert as no cert is provided
I0902 18:15:35.135060       1 main.go:243] Starting TCP socket on 0.0.0.0:8443
I0902 18:15:35.175185       1 main.go:250] Listening securely on 0.0.0.0:8443
I0902 18:15:51.111202       1 listener.go:44] controller-runtime/metrics "msg"="metrics server is starting to listen"  "addr"="127.0.0.1:8080"
I0902 18:15:51.113054       1 main.go:205] setup "msg"="starting manager"
I0902 18:15:51.113917       1 leaderelection.go:242] attempting to acquire leader lease  capg-system/controller-leader-election-capg...
I0902 18:15:51.114691       1 internal.go:356] controller-runtime/manager "msg"="starting metrics server"  "path"="/metrics"
I0902 18:15:51.142032       1 leaderelection.go:252] successfully acquired lease capg-system/controller-leader-election-capg
I0902 18:15:51.145165       1 controller.go:164] controller-runtime/controller "msg"="Starting EventSource"  "c
```

$ kubectl get machines
```
NAME                        PROVIDERID                                                       PHASE
gtc-control-plane-cxcd4     gce://virtual-anchor-281401/us-west1-a/gtc-control-plane-vmplz   Running
gtc-md-0-6cf7474cff-zpbxv   gce://virtual-anchor-281401/us-west1-a/gtc-md-0-7mccx            Running
```

$ kubectl --namespace=default get secret/gtc-kubeconfig -o jsonpath={.data.value} | base64 --decode > ./gtc.kubeconfig

$ kubectl get pods -A --kubeconfig ~/gtc.kubeconfig

```
NAMESPACE     NAME                                              READY   STATUS    RESTARTS   AGE
kube-system   calico-kube-controllers-6d4fbb6df9-8lf4f          1/1     Running   0          5m18s
kube-system   calico-node-6lmqw                                 1/1     Running   0          73s
kube-system   calico-node-qtgzj                                 1/1     Running   1          5m18s
kube-system   coredns-5644d7b6d9-dqd75                          1/1     Running   0          5m18s
kube-system   coredns-5644d7b6d9-ls2q9                          1/1     Running   0          5m18s
kube-system   etcd-gtc-control-plane-vmplz                      1/1     Running   0          4m53s
kube-system   kube-apiserver-gtc-control-plane-vmplz            1/1     Running   0          4m42s
kube-system   kube-controller-manager-gtc-control-plane-vmplz   1/1     Running   0          4m59s
kube-system   kube-proxy-6hk8c                                  1/1     Running   0          5m18s
kube-system   kube-proxy-b8mqw                                  1/1     Running   0          73s
kube-system   kube-scheduler-gtc-control-plane-vmplz            1/1     Running   0          4m47s
```

Now, the control plane and worker node are created on google cloud.

## Tear Down Clusters

If you would like to delete the cluster run the below commands. This will delete
the control plane, workers, machine health check and all other resources
associated with the cluster on gcp.

$ airshipctl phase render controlplane -k Cluster

```
---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: gtc
  namespace: default
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 192.168.0.0/16
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
    kind: KubeadmControlPlane
    name: gtc-control-plane
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
    kind: GCPCluster
    name: gtc
...
```

$ airshipctl phase render controlplane -k Cluster | kubectl delete -f -

```
cluster.cluster.x-k8s.io "gtc" deleted
```

$ kind delete cluster --name capi-gcp
```
Deleting cluster "capi-gcp" ...
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

#### gcp-test-site/shared
airshipctl phase run clusterctl-init-ephemeral uses
airshipctl/manifests/site/gcp-test-site/shared/clusterctl to initialize
management cluster with defined provider components and version.

$ tree airshipctl/manifests/site/gcp-test-site/shared
```
airshipctl/manifests/site/gcp-test-site/shared
└── clusterctl
    ├── clusterctl.yaml
    └── kustomization.yaml
```

#### gcp-test-site/target
There are 3 phases currently available in gcp-test-site/target

|Phase Name | Purpose |
|-----------|---------|
| controlplane     | Patches templates in manifests/function/k8scontrol-capg |
| workers          | Patches template in manifests/function/workers-capg |                                                     |
| initinfra | Simply calls `gcp-test-site/shared/clusterctl` |

Note: `airshipctl phase run clusterctl-init-ephemeral` initializes all the provider components
including the gcp infrastructure provider component.

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

$ tree airshipctl/manifests/site/gcp-test-site/target/

```
airshipctl/manifests/site/gcp-test-site/target/
├── controlplane
│   ├── kustomization.yaml
│   ├── machine_count.json
│   ├── machine_type.json
│   ├── network_name.json
│   ├── project_name.json
│   └── region_name.json
├── initinfra
│   └── kustomization.yaml
└── workers
    ├── failure_domain.json
    ├── kustomization.yaml
    ├── machine_count.json
    └── machine_type.json

3 directories, 11 files

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

```
kind v0.8.1 go1.14.2 linux/amd64
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
