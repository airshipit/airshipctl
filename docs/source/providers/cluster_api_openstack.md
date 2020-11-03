# Airshipctl integration with Cluster API Openstack

## Overview

This document provides instructions on the usage of airshipctl, to perform the
following operations with openstack as infrastructure provider:

- Initialize the management cluster with cluster api, and cluster api openstack
  provider components
- Create a target workload cluster with controlplane and worker machines on an openstack
  cloud environment

## Workflow

A simple workflow that can be tested involves the following operations:

Initialize a management cluster with cluster api and openstack provider
components:

*`$ airshipctl phase run clusterctl-init-ephemeral`*  or *`$ airshipctl phase run clusterctl-init-ephemeral --debug`*

Create a target workload cluster with control plane and worker nodes:

*`$ airshipctl phase run controlplane-target`*

*`$ airshipctl phase run workers-target`*

## Common Prerequisite

- Install [Docker](https://www.docker.com/)
- Install and setup [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- Install [Kind](https://kind.sigs.k8s.io/)
- Install [Kustomize](https://kubernetes-sigs.github.io/kustomize/installation/binaries/)
- Install [Airshipctl](https://docs.airshipit.org/airshipctl/developers.html)

## Openstack Prerequisites

### Credentials

In order to comunicate with openstack cloud environment, following set of credentials are
needed to be generated.

The [env.rc](<https://github.com/kubernetes-sigs/cluster-api-provider-openstack/blob/master/docs/env.rc>) script
sets the required environment variables related to credentials.

`source env.rc <path/to/clouds.yaml> <cloud>`

The following variables are set.

```bash
OPENSTACK_CLOUD: The cloud name which is used as second argument in the command above
OPENSTACK_CLOUD_YAML_B64: The secret used by capo to access OpenStack cloud
OPENSTACK_CLOUD_PROVIDER_CONF_B64: The content of cloud.conf used by OpenStack cloud
OPENSTACK_CLOUD_CACERT_B64: (Optional) The content of custom CA file which can be specified in the clouds.yaml
```

### SSH key pair

An ssh key-pair must be specified by setting the `OPENSTACK_SSH_KEY_NAME` environment variable.

A key-pair can be created by executing the following command

`openstack keypair create [--public-key <file> | --private-key <file>] <name>`

### Availability zone

The availability zone must be set as an environment variable `OPENSTACK_FAILURE_DOMAIN`.

### DNS server

The DNS servers must be set as an environment variable `OPENSTACK_DNS_NAMESERVERS`.

### External network

The openstack environment should have an external network already present.
The external network id can be specified by setting the `spec.externalNetworkId` of `OpenStackCluster` CRD of the cluster template.
The public network id can be obtained by using command

```bash
openstack network list --external
```

### Floating IP

A floating IP is automatically created and associated with the load balancer or controller node, however floating IP can also be
specified explicitly by setting the `spec.apiServerLoadBalancerFlotingIP` of `OpenStackCluster` CRD.

Floating ip can be created using `openstack floating ip create <public network>` command.

Note: Only user with admin role can create a floating IP with specific IP address.

### Operating system image

A cluster api compatible image is required for creating workload kubernetes clusters. The kubeadm bootstrap provider that capo uses
depends on some pre-installed software like a container runtime, kubelet, kubeadm and also on an up-to-date version of cloud-init.

The image can be referenced by setting an environment variable `OPENSTACK_IMAGE_NAME`.

#### Install Packer

$ mkdir packer

$ cd packer

$ wget <https://releases.hashicorp.com/packer/1.6.0/packer_1.6.0_linux_amd64.zip>

$ unzip packer_1.6.0_linux_amd64.zip

$ sudo mv packer /usr/local/bin/

#### Install Ansible

$ sudo apt update

$ sudo apt upgrade

$ sudo apt install software-properties-common

$ sudo apt-add-repository ppa:ansible/ansible

$ sudo apt update

$ sudo apt install ansible

#### Build Cluster API Compliant VM Image

```bash
$ sudo -i
# apt install qemu-kvm libvirt-bin qemu-utils
$ sudo usermod -a -G kvm `<yourusername>`
$ sudo chown root:kvm /dev/kvm

```

Exit and log back in to make the change take place.

```bash
git clone <https://github.com/kubernetes-sigs/image-builder.git> image-builder
cd image-builder/images/capi/
vim packer/qemu/qemu-ubuntu-1804.json
```

Update the iso_url to `http://cdimage.ubuntu.com/releases/18.04/release/ubuntu-18.04.5-server-amd64.iso`

Make sure to use the correct checksum value from
[ubuntu-releases](http://cdimage.ubuntu.com/releases/18.04.5/release/SHA256SUMS)

$ make build-qemu-ubuntu-1804

#### Upload Images to Openstack

$ openstack image create --container-format bare --disk-format qcow2 --file  ubuntu-1804-kube-v1.16.14 ubuntu-1804-kube-v1.16.4

``` bash
$ openstack image list
+--------------------------------------+--------------------------+--------+
| ID                                   | Name                     | Status |
+--------------------------------------+--------------------------+--------+
| 10e31af1-5414-4bae-9500-922db677e695 | amphora-x64-haproxy      | active |
| 61bf8071-5e00-4806-83e0-612f8da03bf8 | cirros-0.5.1-x86_64-disk | active |
| 4fd894c7-9964-461b-bc9f-2e90fdade505 | ubuntu-1804-kube-v1.16.4 | active |
+--------------------------------------+--------------------------+--------+
```

## Getting Started

Kind is used to setup a kubernetes cluster, that will later be transformed
into a management cluster using airshipctl. The kind kubernetes cluster will be
initialized with cluster API and Cluster API openstack(CAPO) provider components.

$ export KIND_EXPERIMENTAL_DOCKER_NETWORK=bridge

$ kind create cluster --name capi-openstack --config ~/kind-cluster-config.yaml

```bash
Creating cluster "capi-openstack" ...
WARNING: Overriding docker network due to KIND_EXPERIMENTAL_DOCKER_NETWORK
WARNING: Here be dragons! This is not supported currently.
 ✓ Ensuring node image (kindest/node:v1.18.2) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-capi-openstack"
You can now use your cluster with:

kubectl cluster-info --context kind-capi-openstack
```

Check if all the pods are up.
$ kubectl get pods -A

```bash
NAMESPACE            NAME                                                   READY   STATUS    RESTARTS   AGE
kube-system          coredns-66bff467f8-2thc2                               1/1     Running   0          2m43s
kube-system          coredns-66bff467f8-4qbvk                               1/1     Running   0          2m43s
kube-system          etcd-capi-openstack-control-plane                      1/1     Running   0          2m58s
kube-system          kindnet-xwp2x                                          1/1     Running   0          2m43s
kube-system          kube-apiserver-capi-openstack-control-plane            1/1     Running   0          2m58s
kube-system          kube-controller-manager-capi-openstack-control-plane   1/1     Running   0          2m58s
kube-system          kube-proxy-khhvd                                       1/1     Running   0          2m43s
kube-system          kube-scheduler-capi-openstack-control-plane            1/1     Running   0          2m58s
local-path-storage   local-path-provisioner-bd4bb6b75-qnbjk                 1/1     Running   0          2m43s
```

## Create airshipctl configuration

$ mkdir ~/.airship

$ airshipctl config init

Run the below command to configure openstack manifest, and add it to airship config

$ airshipctl config set-manifest openstack_manifest --repo primary --url \
<https://opendev.org/airship/airshipctl> --branch master --primary \
--sub-path manifests/site/openstack-test-site --target-path /tmp/airship/

$ airshipctl config set-context kind-capi-openstack --manifest openstack_manifest

```bash
Context "kind-capi-openstack" created.
```

$ cp ~/.kube/config ~/.airship/kubeconfig

$ airshipctl config get-context

```bash
Context: kind-capi-openstack
contextKubeconf: kind-capi-openstack_target
manifest: openstack_manifest

LocationOfOrigin: /home/stack/.airship/kubeconfig
cluster: kind-capi-openstack_target
user: kind-capi-openstack
```

$ airshipctl config use-context kind-capi-openstack

$ airshipctl document pull --debug

```bash
[airshipctl] 2020/09/10 23:19:32 Reading current context manifest information from /home/stack/.airship/config
[airshipctl] 2020/09/10 23:19:32 Downloading primary repository airshipctl from https://opendev.org/airship/airshipctl into /tmp/airship/
[airshipctl] 2020/09/10 23:19:32 Attempting to download the repository airshipctl
[airshipctl] 2020/09/10 23:19:32 Attempting to clone the repository airshipctl from https://opendev.org/airship/airshipctl
[airshipctl] 2020/09/10 23:19:32 Attempting to open repository airshipctl
[airshipctl] 2020/09/10 23:19:32 Attempting to checkout the repository airshipctl from branch refs/heads/master
```

$ airshipctl config set-manifest openstack_manifest --target-path /tmp/airship/airshipctl

## Initialize Management cluster

Execute the following command to initialize the Management cluster with CAPI and CAPO components.

$ airshipctl phase run clusterctl-init-ephemeral --debug

```bash
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:104: Verifying that variable CONTAINER_CAPD_AUTH_PROXY is allowed to be appended
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:104: Verifying that variable CONTAINER_CAPM3_MANAGER is allowed to be appended
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:104: Verifying that variable CONTAINER_CABPK_AUTH_PROXY is allowed to be appended
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:104: Verifying that variable CONTAINER_CABPK_MANAGER is allowed to be appended
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:104: Verifying that variable CONTAINER_CACPK_AUTH_PROXY is allowed to be appended
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:104: Verifying that variable CONTAINER_CAPI_MANAGER is allowed to be appended
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:104: Verifying that variable CONTAINER_CAPM3_AUTH_PROXY is allowed to be appended
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:104: Verifying that variable CONTAINER_CACPK_MANAGER is allowed to be appended
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:104: Verifying that variable CONTAINER_CAPD_MANAGER is allowed to be appended
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/implementations/reader.go:104: Verifying that variable CONTAINER_CAPI_AUTH_PROXY is allowed to be appended
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/events/processor.go:61: Received event: {4 {InitType {[]} {<nil>} {ApplyEventResourceUpdate ServersideApplied <nil>} {ResourceUpdateEvent <nil> <nil>} {PruneEventResourceUpdate Pruned <nil>} {DeleteEventResourceUpdate Deleted <nil>}} {<nil>} {ResourceUpdateEvent <nil> <nil>} {0 starting clusterctl init executor} {0 }}
[airshipctl] 2020/10/11 06:03:40 opendev.org/airship/airshipctl@/pkg/clusterctl/client/client.go:67: Starting cluster-api initiation
Installing the clusterctl inventory CRD
...
```

Wait for all the pods to be up.

$ kubectl get pods -A

```bash
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-dfdf9877b-g44hd        2/2     Running   0          59s
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-76c847457b-z2jtr   2/2     Running   0          58s
capi-system                         capi-controller-manager-7c7978f565-rk7qk                         2/2     Running   0          59s
capi-webhook-system                 capi-controller-manager-748c57d64d-wjbnj                         2/2     Running   0          60s
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-65f979767f-bv6dr       2/2     Running   0          59s
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-7f5d88dcf9-k6kpf   2/2     Running   1          58s
capi-webhook-system                 capo-controller-manager-7d76dc9ddc-b9xhw                         2/2     Running   0          57s
capo-system                         capo-controller-manager-79445d5984-k9fmc                         2/2     Running   0          57s
cert-manager                        cert-manager-77d8f4d85f-nkg58                                    1/1     Running   0          71s
cert-manager                        cert-manager-cainjector-75f88c9f56-fcrc6                         1/1     Running   0          72s
cert-manager                        cert-manager-webhook-56669d7fcb-cbzfn                            1/1     Running   1          71s
kube-system                         coredns-66bff467f8-2thc2                                         1/1     Running   0          29m
kube-system                         coredns-66bff467f8-4qbvk                                         1/1     Running   0          29m
kube-system                         etcd-capi-openstack-control-plane                                1/1     Running   0          29m
kube-system                         kindnet-xwp2x                                                    1/1     Running   0          29m
kube-system                         kube-apiserver-capi-openstack-control-plane                      1/1     Running   0          29m
kube-system                         kube-controller-manager-capi-openstack-control-plane             1/1     Running   0          29m
kube-system                         kube-proxy-khhvd                                                 1/1     Running   0          29m
kube-system                         kube-scheduler-capi-openstack-control-plane                      1/1     Running   0          29m
local-path-storage                  local-path-provisioner-bd4bb6b75-qnbjk                           1/1     Running   0          29m
```

At this point, the management cluster is initialized with cluster api and cluster api openstack provider components.

## Create control plane and worker nodes

$ airshipctl phase run controlplane-target --debug

```bash
[airshipctl] 2020/10/11 06:05:31 opendev.org/airship/airshipctl@/pkg/k8s/applier/executor.go:126: Getting kubeconfig file information from kubeconfig provider
[airshipctl] 2020/10/11 06:05:31 opendev.org/airship/airshipctl@/pkg/k8s/applier/executor.go:131: Filtering out documents that shouldnt be applied to kubernetes from document bundle
[airshipctl] 2020/10/11 06:05:31 opendev.org/airship/airshipctl@/pkg/k8s/applier/executor.go:115: WaitTimeout: 33m20s
[airshipctl] 2020/10/11 06:05:31 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:77: Getting infos for bundle, inventory id is controlplane-target
[airshipctl] 2020/10/11 06:05:31 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:107: Inventory Object config Map not found, auto generating Inventory object
[airshipctl] 2020/10/11 06:05:31 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:114: Injecting Inventory Object: {"apiVersion":"v1","kind":"ConfigMap","metadata":{"creationTimestamp":null,"labels":{"cli-utils.sigs.k8s.io/inventory-id":"controlplane-target"},"name":"airshipit-controlplane-target","namespace":"airshipit"}}{nsfx:false,beh:unspecified} into bundle
[airshipctl] 2020/10/11 06:05:31 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:120: Making sure that inventory object namespace airshipit exists
secret/ostgt-cloud-config created
cluster.cluster.x-k8s.io/ostgt created
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/ostgt-control-plane created
openstackcluster.infrastructure.cluster.x-k8s.io/ostgt created
openstackmachinetemplate.infrastructure.cluster.x-k8s.io/ostgt-control-plane created
5 resource(s) applied. 5 created, 0 unchanged, 0 configured
cluster.cluster.x-k8s.io/ostgt is NotFound: Resource not found
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/ostgt-control-plane is NotFound: Resource not found
openstackcluster.infrastructure.cluster.x-k8s.io/ostgt is NotFound: Resource not found
openstackmachinetemplate.infrastructure.cluster.x-k8s.io/ostgt-control-plane is NotFound: Resource not found
secret/ostgt-cloud-config is NotFound: Resource not found
secret/ostgt-cloud-config is Current: Resource is always ready
cluster.cluster.x-k8s.io/ostgt is InProgress:
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/ostgt-control-plane is Current: Resource is current
openstackcluster.infrastructure.cluster.x-k8s.io/ostgt is Current: Resource is current
openstackmachinetemplate.infrastructure.cluster.x-k8s.io/ostgt-control-plane is Current: Resource is current
cluster.cluster.x-k8s.io/ostgt is InProgress:
openstackcluster.infrastructure.cluster.x-k8s.io/ostgt is Current: Resource is current
cluster.cluster.x-k8s.io/ostgt is InProgress: Scaling up to 1 replicas (actual 0)
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/ostgt-control-plane is InProgress: Scaling up to 1 replicas (actual 0)
cluster.cluster.x-k8s.io/ostgt is InProgress:
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/ostgt-control-plane is InProgress:
cluster.cluster.x-k8s.io/ostgt is Current: Resource is Ready
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/ostgt-control-plane is Current: Resource is Ready
all resources has reached the Current status
```

$ airshipctl phase run workers-target --debug

```bash
[airshipctl] 2020/10/11 06:05:48 opendev.org/airship/airshipctl@/pkg/k8s/applier/executor.go:126: Getting kubeconfig file information from kubeconfig provider
[airshipctl] 2020/10/11 06:05:48 opendev.org/airship/airshipctl@/pkg/k8s/applier/executor.go:131: Filtering out documents that shouldnt be applied to kubernetes from document bundle
[airshipctl] 2020/10/11 06:05:48 opendev.org/airship/airshipctl@/pkg/k8s/applier/executor.go:115: WaitTimeout: 33m20s
[airshipctl] 2020/10/11 06:05:48 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:77: Getting infos for bundle, inventory id is workers-target
[airshipctl] 2020/10/11 06:05:48 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:107: Inventory Object config Map not found, auto generating Inventory object
[airshipctl] 2020/10/11 06:05:48 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:114: Injecting Inventory Object: {"apiVersion":"v1","kind":"ConfigMap","metadata":{"creationTimestamp":null,"labels":{"cli-utils.sigs.k8s.io/inventory-id":"workers-target"},"name":"airshipit-workers-target","namespace":"airshipit"}}{nsfx:false,beh:unspecified} into bundle
[airshipctl] 2020/10/11 06:05:48 opendev.org/airship/airshipctl@/pkg/k8s/applier/applier.go:120: Making sure that inventory object namespace airshipit exists
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/ostgt-md-0 created
machinedeployment.cluster.x-k8s.io/ostgt-md-0 created
openstackmachinetemplate.infrastructure.cluster.x-k8s.io/ostgt-md-0 created
3 resource(s) applied. 3 created, 0 unchanged, 0 configured
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/ostgt-md-0 is NotFound: Resource not found
machinedeployment.cluster.x-k8s.io/ostgt-md-0 is NotFound: Resource not found
openstackmachinetemplate.infrastructure.cluster.x-k8s.io/ostgt-md-0 is NotFound: Resource not found
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/ostgt-md-0 is Current: Resource is current
machinedeployment.cluster.x-k8s.io/ostgt-md-0 is Current: Resource is current
openstackmachinetemplate.infrastructure.cluster.x-k8s.io/ostgt-md-0 is Current: Resource is current
all resources has reached the Current status
```

$ kubectl get po -A

```bash
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-dfdf9877b-g44hd        2/2     Running   0          36m
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-76c847457b-z2jtr   2/2     Running   0          36m
capi-system                         capi-controller-manager-7c7978f565-rk7qk                         2/2     Running   0          36m
capi-webhook-system                 capi-controller-manager-748c57d64d-wjbnj                         2/2     Running   0          36m
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-65f979767f-bv6dr       2/2     Running   0          36m
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-7f5d88dcf9-k6kpf   2/2     Running   1          36m
capi-webhook-system                 capo-controller-manager-7d76dc9ddc-b9xhw                         2/2     Running   0          36m
capo-system                         capo-controller-manager-79445d5984-k9fmc                         2/2     Running   0          36m
cert-manager                        cert-manager-77d8f4d85f-nkg58                                    1/1     Running   0          36m
cert-manager                        cert-manager-cainjector-75f88c9f56-fcrc6                         1/1     Running   0          36m
cert-manager                        cert-manager-webhook-56669d7fcb-cbzfn                            1/1     Running   1          36m
kube-system                         coredns-66bff467f8-2thc2                                         1/1     Running   0          64m
kube-system                         coredns-66bff467f8-4qbvk                                         1/1     Running   0          64m
kube-system                         etcd-capi-openstack-control-plane                                1/1     Running   0          64m
kube-system                         kindnet-xwp2x                                                    1/1     Running   0          64m
kube-system                         kube-apiserver-capi-openstack-control-plane                      1/1     Running   0          64m
kube-system                         kube-controller-manager-capi-openstack-control-plane             1/1     Running   0          64m
kube-system                         kube-proxy-khhvd                                                 1/1     Running   0          64m
kube-system                         kube-scheduler-capi-openstack-control-plane                      1/1     Running   0          64m
local-path-storage                  local-path-provisioner-bd4bb6b75-qnbjk                           1/1     Running   0          64m
```

To check logs run the below command

$ kubectl logs capo-controller-manager-79445d5984-k9fmc -n capo-system --all-containers=true -f

```bash
I0910 23:36:54.768316       1 listener.go:44] controller-runtime/metrics "msg"="metrics server is starting to listen"  "addr"="127.0.0.1:8080"
I0910 23:36:54.768890       1 main.go:235] setup "msg"="starting manager"
I0910 23:36:54.769149       1 leaderelection.go:242] attempting to acquire leader lease  capo-system/controller-leader-election-capo...
I0910 23:36:54.769199       1 internal.go:356] controller-runtime/manager "msg"="starting metrics server"  "path"="/metrics"
I0910 23:36:54.853723       1 leaderelection.go:252] successfully acquired lease capo-system/controller-leader-election-capo
I0910 23:36:54.854706       1 controller.go:164] controller-runtime/controller "msg"="Starting EventSource"  "controller"="openstackcluster" "source"={"                Type":{"metadata":{"creationTimestamp":null},"spec":{"cloudsSecret":null,"cloudName":"","network":{},"subnet":{},"managedAPIServerLoadBalancer":false,"m                anagedSecurityGroups":false,"caKeyPair":{},"etcdCAKeyPair":{},"frontProxyCAKeyPair":{},"saKeyPair":{},"controlPlaneEndpoint":{"host":"","port":0}},"stat                us":{"ready":false}}}
I0910 23:36:54.854962       1 controller.go:164] controller-runtime/controller "msg"="Starting EventSource"  "controller"="openstackmachine" "source"={"                Type":{"metadata":{"creationTimestamp":null},"spec":{"cloudsSecret":null,"cloudName":"","flavor":"","image":""},"status":{"ready":false}}}
```

$ kubectl get machines

```bash
NAME                          PROVIDERID                                         PHASE
ostgt-control-plane-cggt7     openstack://a6da4363-9419-4e14-b67a-3ce86da198c4   Running
ostgt-md-0-6b564d74b8-8h8d8   openstack://23fd5b75-e3f4-4e89-b900-7a6873a146c2   Running
ostgt-md-0-6b564d74b8-pj4lm   openstack://9b8323a2-757f-4905-8006-4514862fde75   Running
ostgt-md-0-6b564d74b8-wnw8l   openstack://1a8f10da-5d12-4c50-a60d-f2e24a387611   Running
```

$ kubectl get secrets

```bash
NAME                        TYPE                                  DATA   AGE
default-token-vfcm7         kubernetes.io/service-account-token   3      114m
ostgt-ca                    Opaque                                2      47m
ostgt-cloud-config          Opaque                                2      51m
ostgt-control-plane-gd2gq   cluster.x-k8s.io/secret               1      47m
ostgt-etcd                  Opaque                                2      47m
ostgt-kubeconfig            Opaque                                1      47m
ostgt-md-0-j76jg            cluster.x-k8s.io/secret               1      44m
ostgt-md-0-kdjsv            cluster.x-k8s.io/secret               1      44m
ostgt-md-0-q4vmn            cluster.x-k8s.io/secret               1      44m
ostgt-proxy                 Opaque                                2      47m
ostgt-sa                    Opaque                                2      47m
```

$ kubectl --namespace=default get secret/ostgt-kubeconfig -o jsonpath={.data.value} | base64 --decode > ./ostgt.kubeconfig

$ kubectl get pods -A --kubeconfig ~/ostgt.kubeconfig

```bash
NAMESPACE     NAME                                                READY   STATUS    RESTARTS   AGE
kube-system   calico-kube-controllers-7865ff46b6-8pbnq            1/1     Running   0          47m
kube-system   calico-node-7kpjb                                   1/1     Running   0          44m
kube-system   calico-node-d8dcc                                   1/1     Running   0          45m
kube-system   calico-node-mdwnt                                   1/1     Running   0          47m
kube-system   calico-node-n2qr8                                   1/1     Running   0          45m
kube-system   coredns-6955765f44-dkvwq                            1/1     Running   0          47m
kube-system   coredns-6955765f44-p4mbh                            1/1     Running   0          47m
kube-system   etcd-ostgt-control-plane-vpmqg                      1/1     Running   0          47m
kube-system   kube-apiserver-ostgt-control-plane-vpmqg            1/1     Running   0          47m
kube-system   kube-controller-manager-ostgt-control-plane-vpmqg   1/1     Running   0          47m
kube-system   kube-proxy-j6msn                                    1/1     Running   0          44m
kube-system   kube-proxy-kgxvq                                    1/1     Running   0          45m
kube-system   kube-proxy-lfmlf                                    1/1     Running   0          45m
kube-system   kube-proxy-zq26j                                    1/1     Running   0          47m
kube-system   kube-scheduler-ostgt-control-plane-vpmqg            1/1     Running   0          47m
```

$ kubectl get nodes --kubeconfig ~/ostgt.kubeconfig

```bash
NAME                        STATUS   ROLES    AGE   VERSION
ostgt-control-plane-vpmqg   Ready    master   49m   v1.17.3
ostgt-md-0-6p2f9            Ready    <none>   46m   v1.17.3
ostgt-md-0-h8hn9            Ready    <none>   47m   v1.17.3
ostgt-md-0-k9k66            Ready    <none>   46m   v1.17.3
```

$ kubectl get cs --kubeconfig ~/ostgt.kubeconfig

```bash
NAME                 STATUS    MESSAGE             ERROR
controller-manager   Healthy   ok
scheduler            Healthy   ok
etcd-0               Healthy   {"health":"true"}
```

Now, the control plane and worker node are created on openstack.

![Machines](../img/openstack-machines.png)

## Tear Down Clusters

In order to delete the cluster run the below command. This will delete
the control plane, workers and all other resources
associated with the cluster on openstack.

```bash
$ kubectl delete cluster ostgt
cluster.cluster.x-k8s.io "ostgt" deleted
```

$ kind delete cluster --name capi-openstack

## Reference

### Installation Using Devstack

- Install [Devstack](https://docs.openstack.org/devstack/latest/guides/devstack-with-lbaas-v2.html)

- Create `ubuntu-1910-kube-v1.17.3.qcow2` image in the devstack.

Download a capi compatible image for ubuntu OS.

```bash
wget https://github.com/sbueringer/image-builder/releases/download/v1.17.3-04/ubuntu-1910-kube-v1.17.3.qcow2

openstack image create --container-format bare --disk-format qcow2 --file  ubuntu-1910-kube-v1.17.3.qcow2 ubuntu-1910-kube-v1.17.3
```

Check if the image status is `active`

```bash
stack@stackdev-ev:/opt/stack/devstack$ openstack image list
+--------------------------------------+--------------------------+--------+
| ID                                   | Name                     | Status |
+--------------------------------------+--------------------------+--------+
| 83002c1d-436d-4007-bea1-3ffc94fa193b | amphora-x64-haproxy      | active |
| a801c914-a0b9-485a-ba5f-246e912cb656 | cirros-0.5.1-x86_64-disk | active |
| 8e8fc7a8-cfe0-4251-bdde-8600838f2ed8 | ubuntu-1910-kube-v1.17.3 | active |
```

- Generate credentials

In devstack environment, normally the `clouds.yaml` file is found at `etc/openstack/` location.

Execute the following command to generate the cloud credentials for devstack

```bash
wget https://raw.githubusercontent.com/kubernetes-sigs/cluster-api-provider-openstack/master/templates/env.rc -O /tmp/env.rc
source /tmp/env.rc /etc/openstack/clouds.yaml devstack
```

A snippet of sample clouds.yaml file can be seen below for cloud `devstack`.

```bash
clouds:
  devstack:
    auth:
      auth_url: http://10.0.4.4/identity
      project_id: f3deb6e94bee4addaed3ba42d6ffaeba
      user_domain_name: Default
      username: demo
      password: pass
    region_name: RegionOne
```

The list of project_id-s can be retrieved by `openstack project list` in the devstack environment.

- Ensure that `demo` user has `admin` rights so that floating ip-s can be created at the time of
workload cluster deployment.

```bash
cd /opt/stack/devstack
export OS_USERNAME=admin
$ . ./openrc
$ openstack role add --project demo --user demo admin
```

- Create Floating IP

To create floating ip, following command can be used

`openstack floating ip create public --floating-ip-address $FLOATING_IP_ADDRESS`

where `FLOATING_IP_ADDRESS` is the specified ip address and `public` is the name of
the external network in devstack.

`openstack floating ip list` command shows the list of all floating ip-s.

- Allow ssh access to controlplane and worker nodes

Cluster api creates following security groups if `spec.managedSecurityGroups` of
`OpenStackCluster` CRD is set to true.

- k8s-cluster-default-`<CLUSTER-NAME>`-secgroup-controlplane (for control plane)
- k8s-cluster-default-`<CLUSTER-NAME>`-secgroup-worker (for worker nodes)

These security group rules include the kubeadm's
[Check required ports](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#check-required-ports)
so that each node can not be logged in through ssh by default.

If ssh access to the nodes is required then follow the below steps -

Create a security group allowing ssh access

```bash
openstack security group create --project demo --project-domain Default allow-ssh
openstack security group rule create allow-ssh --protocol tcp --dst-port 22:22 --remote-ip 0.0.0.0/0
```

Add the security group to `OpenStackMachineTemplate` CRD as below

```bash
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: OpenStackMachineTemplate
metadata:
  name: ${CLUSTER_NAME}-control-plane
spec:
  template:
    spec:
      securityGroups:
      - name: allow-ssh
```

### Provider Manifests

Provider Configuration for Capo is referenced from
[Config](https://github.com/kubernetes-sigs/cluster-api-provider-openstack/tree/master/config)

$ tree airshipctl/manifests/function/capo

```bash
└── v0.3.1
    ├── certmanager
    │   ├── certificate.yaml
    │   ├── kustomization.yaml
    │   └── kustomizeconfig.yaml
    ├── crd
    │   ├── bases
    │   │   ├── infrastructure.cluster.x-k8s.io_openstackclusters.yaml
    │   │   ├── infrastructure.cluster.x-k8s.io_openstackmachines.yaml
    │   │   └── infrastructure.cluster.x-k8s.io_openstackmachinetemplates.yaml
    │   ├── kustomization.yaml
    │   ├── kustomizeconfig.yaml
    │   └── patches
    │       ├── cainjection_in_openstackclusters.yaml
    │       ├── cainjection_in_openstackmachines.yaml
    │       ├── cainjection_in_openstackmachinetemplates.yaml
    │       ├── webhook_in_openstackclusters.yaml
    │       ├── webhook_in_openstackmachines.yaml
    │       └── webhook_in_openstackmachinetemplates.yaml
    ├── default
    │   ├── kustomization.yaml
    │   ├── manager_role_aggregation_patch.yaml
    │   └── namespace.yaml
    ├── kustomization.yaml
    ├── manager
    │   ├── kustomization.yaml
    │   ├── manager.yaml
    │   ├── manager_auth_proxy_patch.yaml
    │   ├── manager_image_patch.yaml
    │   └── manager_pull_policy.yaml
    ├── patch_crd_webhook_namespace.yaml
    ├── rbac
    │   ├── auth_proxy_role.yaml
    │   ├── auth_proxy_role_binding.yaml
    │   ├── auth_proxy_service.yaml
    │   ├── kustomization.yaml
    │   ├── leader_election_role.yaml
    │   ├── leader_election_role_binding.yaml
    │   ├── role.yaml
    │   └── role_binding.yaml
    └── webhook
        ├── kustomization.yaml
        ├── kustomizeconfig.yaml
        ├── manager_webhook_patch.yaml
        ├── manifests.yaml
        ├── service.yaml
        └── webhookcainjection_patch.yaml
```

### Kind Configuration

```bash
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
     apiServerAddress: "127.0.0.1"
     apiServerPort: 37533
nodes:
  - role: control-plane
    extraMounts:
      - hostPath: /var/run/docker.sock
        containerPath: /var/run/docker.sock
      - hostPath: /tmp/airship/airshipctl/tools/deployment/certificates
        containerPath: /etc/kubernetes/pki
    kubeadmConfigPatches:
    - |
      kind: ClusterConfiguration
      certificatesDir: /etc/kubernetes/pki
```

### Capo Phases

```bash
/airshipctl/manifests/capo-phases$ tree
.
├── cluster-map.yaml
├── executors.yaml
├── kubeconfig.yaml
├── kustomization.yaml
├── phases.yaml
└── plan.yaml
```

$ cat phases.yaml

```bash
---
apiVersion: airshipit.org/v1alpha1
kind: Phase
metadata:
  name: clusterctl-init-ephemeral
  clusterName: kind-capi-openstack
config:
  executorRef:
    apiVersion: airshipit.org/v1alpha1
    kind: Clusterctl
    name: clusterctl_init
---
apiVersion: airshipit.org/v1alpha1
kind: Phase
metadata:
  name: controlplane-target
  clusterName: kind-capi-openstack
config:
  executorRef:
    apiVersion: airshipit.org/v1alpha1
    kind: KubernetesApply
    name: kubernetes-apply
  documentEntryPoint: manifests/site/openstack-test-site/target/controlplane
---
apiVersion: airshipit.org/v1alpha1
kind: Phase
metadata:
  name: workers-target
  clusterName: kind-capi-openstack
config:
  cluster: kind-capi-openstack
  executorRef:
    apiVersion: airshipit.org/v1alpha1
    kind: KubernetesApply
    name: kubernetes-apply
  documentEntryPoint: manifests/site/openstack-test-site/target/workers
```

### Cluster Templates

airshipctl/manifests/function/k8scontrol-capo contains cluster.yaml, controlplane.yaml templates.

```bash
cluster.yaml: Contains CRDs Cluster, OpenstackCluster, Secret
controlplane.yaml: Contains CRDs KubeadmControlPlane, OpenstackMachineTemplate
```

$ tree airshipctl/manifests/function/k8scontrol-capo

```bash
airshipctl/manifests/function/k8scontrol-capo
├── cluster.yaml
├── controlplane.yaml
└── kustomization.yaml
```

airshipctl/manifests/function/workers-capo contains workers.yaml

```bash
workers.yaml: Contains CRDs Cluster, OpenstackCluster, Secret
```

$ tree airshipctl/manifests/function/workers-capo

```bash
airshipctl/manifests/function/workers-capo
.
├── kustomization.yaml
└── workers.yaml
```

### Test Site Manifests

#### openstack-test-site/target

Following phase entrypoints reside in the openstack-test-site.

```bash
controlplane - Patches templates in manifests/function/k8scontrol-capo
workers - Patches template in manifests/function/workers-capo
```

Note: `airshipctl phase run clusterctl-init-ephemeral` initializes all the provider components including the openstack infrastructure provider component.

#### Patch Merge Strategy

Json and strategic merge patches are applied on templates in `manifests/function/k8scontrol-capo`
from `airshipctl/manifests/site/openstack-test-site/target/controlplane` when
`airshipctl phase run controlplane-target` command is executed

Json and strategic merge patches are applied on templates in `manifests/function/workers-capo`
from `airshipctl/manifests/site/openstack-test-site/target/workers` when
`airshipctl phase run workers-target` command is executed

```bash
controlplane/control_plane_ip.json: patches control plane ip in template function/k8scontrol-capo/cluster.yaml
controlplane/dns_servers.json: patches dns servers in template function/k8scontrol-capo/cluster.yaml
controlplane/external_network_id.json: patches external network id in template function/k8scontrol-capo/cluster.yaml
cluster_clouds_yaml_patch.yaml: patches clouds.yaml configuration in template function/k8scontrol-capo/cluster.yaml
controlplane/control_plane_ip_patch.yaml: patches controlplane ip in template function/k8scontrol-capo/controlplane.yaml
controlplane/control_plane_config_patch.yaml: patches cloud configuration in template function/k8scontrol-capo/controlplane.yaml
controlplane/ssh_key_patch.yaml: patches ssh key in template function/k8scontrol-capo/controlplane.yaml
controlplane/control_plane_machine_count_patch.yaml: patches controlplane replica count in template function/k8scontrol-capo/controlplane.yaml
controlplane/control_plane_machine_flavor_patch.yaml: patches controlplane machine flavor in template function/k8scontrol-capo/controlplane.yaml
workers/workers_cloud_conf_patch.yaml: patches cloud configuration in template function/workers-capo/workers.yaml
workers/workers_machine_count_patch.yaml: patches worker replica count in template function/workers-capo/workers.yaml
workers/workers_machine_flavor_patch.yaml: patches worker machine flavor in template function/workers-capo/workers.yaml
workers/workers_ssh_key_patch.yaml: patches ssh key in template function/workers-capo/workers.yaml

```

## Software Version Information

All the instructions provided in the document have been tested using the software and
version, provided in this section.

### Docker

```bash
$ docker version
Client: Docker Engine - Community
 Version:           19.03.12
 API version:       1.40
 Go version:        go1.13.10
 Git commit:        48a66213fe
 Built:             Mon Jun 22 15:45:36 2020
 OS/Arch:           linux/amd64
 Experimental:      false

Server: Docker Engine - Community
 Engine:
  Version:          19.03.12
  API version:      1.40 (minimum version 1.12)
  Go version:       go1.13.10
  Git commit:       48a66213fe
  Built:            Mon Jun 22 15:44:07 2020
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

### Kind

```bash
$ kind version
kind v0.8.1 go1.14.2 linux/amd64
```

### Kubectl

```bash
$ kubectl version --short
Client Version: v1.17.4
Server Version: v1.18.2
```

### Go

```bash
$ go version
go version go1.10.4 linux/amd64
```

### Kustomize

```bash
$ kustomize version
{Version:kustomize/v3.8.0 GitCommit:6a50372dd5686df22750b0c729adaf369fbf193c BuildDate:2020-07-05T14:08:42Z GoOs:linux GoArch:amd64}
```

### OS

```bash
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

## Virtual Machine Specification

All the instructions in the document were perfomed on VM(with nested virtualization enabled)
with 16 vCPUs, 64 GB RAM.
