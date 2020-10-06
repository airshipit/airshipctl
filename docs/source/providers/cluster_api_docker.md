# Airshipctl and Cluster API Docker Integration

## Overview
Airshipctl and cluster api docker integration facilitates usage of `airshipctl`
to create cluster api management and workload clusters using `docker as
infrastructure provider`.

## Workflow
A simple workflow that can be tested involves the following operations:

**Initialize the management cluster with cluster api and docker provider
components**

> airshipctl cluster init --debug

**Create a workload cluster, with control plane and worker nodes**

> airshipctl phase run controlplane

> airshipctl phase run workers

Note: `airshipctl phase run initinfra-ephemeral` is not used because all the provider
components are initialized  using `airshipctl cluster init`

The phase `initinfra` is included just to get `validate docs` to
pass.

## Common Pre-requisites

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
initialized with cluster API and Cluster API docker provider components.

Run the following command to create a kind configuration file, that  would mount
the docker.sock file from the host operating system into the kind cluster. This
is required by the management cluster to create machines on host as docker
containers.

$ vim kind-cluster-with-extramounts.yaml

```
kind: Cluster
apiVersion: kind.sigs.k8s.io/v1alpha3
nodes:
  - role: control-plane
    extraMounts:
      - hostPath: /var/run/docker.sock
        containerPath: /var/run/docker.sock

```

Save and exit.

$ export KIND_EXPERIMENTAL_DOCKER_NETWORK=bridge

$ kind create cluster --name capi-docker --config ~/kind-cluster-with-extramounts.yaml

```
Creating cluster "capi-docker" ...
WARNING: Overriding docker network due to KIND_EXPERIMENTAL_DOCKER_NETWORK
WARNING: Here be dragons! This is not supported currently.
 ✓ Ensuring node image (kindest/node:v1.18.2) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-capi-docker"
You can now use your cluster with:

kubectl cluster-info --context kind-capi-docker
Check if all the pods are up.
```

$ kubectl get pods -A

```
NAMESPACE            NAME                                                READY   STATUS    RESTARTS   AGE
kube-system          coredns-6955765f44-fvg8p                            1/1     Running   0          72s
kube-system          coredns-6955765f44-gm96d                            1/1     Running   0          72s
kube-system          etcd-capi-docker-control-plane                      1/1     Running   0          83s
kube-system          kindnet-7fqv6                                       1/1     Running   0          72s
kube-system          kube-apiserver-capi-docker-control-plane            1/1     Running   0          83s
kube-system          kube-controller-manager-capi-docker-control-plane   1/1     Running   0          83s
kube-system          kube-proxy-gqlnm                                    1/1     Running   0          72s
kube-system          kube-scheduler-capi-docker-control-plane            1/1     Running   0          83s
local-path-storage   local-path-provisioner-7745554f7f-2qcv7             1/1     Running   0          72s
```

## Create airshipctl configuration files

$ mkdir ~/.airship

$ airshipctl config init

Run the below command to configure docker manifest, and add it to airship config

```
$ airshipctl config set-manifest docker_manifest --repo primary \
--url https://opendev.org/airship/airshipctl --branch master \
--phase --sub-path manifests/site/docker-test-site --target-path /tmp/airship
```

$ airshipctl config set-context kind-capi-docker --manifest docker_manifest

```
Context "kind-capi-docker" modified.
```
$ cp ~/.kube/config ~/.airship/kubeconfig

$ airshipctl config get-context

```
Context: kind-capi-docker
contextKubeconf: kind-capi-docker_target
manifest: docker_manifest

LocationOfOrigin: /home/rishabh/.airship/kubeconfig
cluster: kind-capi-docker_target
user: kind-capi-docker
```
$ airshipctl config use-context kind-capi-docker

```
Manifest "docker_manifest" created.
```

$ airshipctl document pull --debug

```
[airshipctl] 2020/08/12 14:07:13 Reading current context manifest information from /home/rishabh/.airship/config
[airshipctl] 2020/08/12 14:07:13 Downloading primary repository airshipctl from https://review.opendev.org/airship/airshipctl into /tmp/airship
[airshipctl] 2020/08/12 14:07:13 Attempting to download the repository airshipctl
[airshipctl] 2020/08/12 14:07:13 Attempting to clone the repository airshipctl from https://review.opendev.org/airship/airshipctl
[airshipctl] 2020/08/12 14:07:23 Attempting to checkout the repository airshipctl from branch refs/heads/master
```
$ airshipctl config set-manifest docker_manifest --target-path /tmp/airship/airshipctl

## Initialize the management cluster

$ airshipctl cluster init --debug

```
[airshipctl] 2020/08/12 14:08:23 Starting cluster-api initiation
Installing the clusterctl inventory CRD
Creating CustomResourceDefinition="providers.clusterctl.cluster.x-k8s.io"
Fetching providers
[airshipctl] 2020/08/12 14:08:23 Creating arishipctl repository implementation interface for provider cluster-api of type CoreProvider
[airshipctl] 2020/08/12 14:08:23 Setting up airshipctl provider Components client
Provider type: CoreProvider, name: cluster-api
[airshipctl] 2020/08/12 14:08:23 Getting airshipctl provider components, skipping variable substitution: true.
Provider type: CoreProvider, name: cluster-api
Fetching File="components.yaml" Provider="cluster-api" Version="v0.3.3"
[airshipctl] 2020/08/12 14:08:23 Building cluster-api provider component documents from kustomize path at /tmp/airship/airshipctl/manifests/function/capi/v0.3.3
[airshipctl] 2020/08/12 14:08:24 Creating arishipctl repository implementation interface for provider kubeadm of type BootstrapProvider
[airshipctl] 2020/08/12 14:08:24 Setting up airshipctl provider Components client
Provider type: BootstrapProvider, name: kubeadm
[airshipctl] 2020/08/12 14:08:24 Getting airshipctl provider components, skipping variable substitution: true.
Provider type: BootstrapProvider, name: kubeadm
Fetching File="components.yaml" Provider="bootstrap-kubeadm" Version="v0.3.3"
[airshipctl] 2020/08/12 14:08:24 Building cluster-api provider component documents from kustomize path at /tmp/airship/airshipctl/manifests/function/cabpk/v0.3.3
[airshipctl] 2020/08/12 14:08:24 Creating arishipctl repository implementation interface for provider kubeadm of type ControlPlaneProvider
[airshipctl] 2020/08/12 14:08:24 Setting up airshipctl provider Components client
Provider type: ControlPlaneProvider, name: kubeadm
[airshipctl] 2020/08/12 14:08:24 Getting airshipctl provider components, skipping variable substitution: true.
Provider type: ControlPlaneProvider, name: kubeadm
Fetching File="components.yaml" Provider="control-plane-kubeadm" Version="v0.3.3"
[airshipctl] 2020/08/12 14:08:24 Building cluster-api provider component documents from kustomize path at /tmp/airship/airshipctl/manifests/function/cacpk/v0.3.3
[airshipctl] 2020/08/12 14:08:24 Creating arishipctl repository implementation interface for provider docker of type InfrastructureProvider
[airshipctl] 2020/08/12 14:08:24 Setting up airshipctl provider Components client
Provider type: InfrastructureProvider, name: docker
[airshipctl] 2020/08/12 14:08:24 Getting airshipctl provider components, skipping variable substitution: true.
Provider type: InfrastructureProvider, name: docker
Fetching File="components.yaml" Provider="infrastructure-docker" Version="v0.3.7"
[airshipctl] 2020/08/12 14:08:24 Building cluster-api provider component documents from kustomize path at /tmp/airship/airshipctl/manifests/function/capd/v0.3.7
[airshipctl] 2020/08/12 14:08:24 Creating arishipctl repository implementation interface for provider cluster-api of type CoreProvider
Fetching File="metadata.yaml" Provider="cluster-api" Version="v0.3.3"
[airshipctl] 2020/08/12 14:08:24 Building cluster-api provider component documents from kustomize path at /tmp/airship/airshipctl/manifests/function/capi/v0.3.3
[airshipctl] 2020/08/12 14:08:25 Creating arishipctl repository implementation interface for provider kubeadm of type BootstrapProvider
Fetching File="metadata.yaml" Provider="bootstrap-kubeadm" Version="v0.3.3"
[airshipctl] 2020/08/12 14:08:25 Building cluster-api provider component documents from kustomize path at /tmp/airship/airshipctl/manifests/function/cabpk/v0.3.3
[airshipctl] 2020/08/12 14:08:25 Creating arishipctl repository implementation interface for provider kubeadm of type ControlPlaneProvider
Fetching File="metadata.yaml" Provider="control-plane-kubeadm" Version="v0.3.3"
[airshipctl] 2020/08/12 14:08:25 Building cluster-api provider component documents from kustomize path at /tmp/airship/airshipctl/manifests/function/cacpk/v0.3.3
[airshipctl] 2020/08/12 14:08:25 Creating arishipctl repository implementation interface for provider docker of type InfrastructureProvider
Fetching File="metadata.yaml" Provider="infrastructure-docker" Version="v0.3.7"
[airshipctl] 2020/08/12 14:08:25 Building cluster-api provider component documents from kustomize path at /tmp/airship/airshipctl/manifests/function/capd/v0.3.7
Installing cert-manager
Creating Namespace="cert-manager"
Creating CustomResourceDefinition="challenges.acme.cert-manager.io"
Creating CustomResourceDefinition="orders.acme.cert-manager.io"
Creating CustomResourceDefinition="certificaterequests.cert-manager.io"
Creating CustomResourceDefinition="certificates.cert-manager.io"
Creating CustomResourceDefinition="clusterissuers.cert-manager.io"
Creating CustomResourceDefinition="issuers.cert-manager.io"
Creating ServiceAccount="cert-manager-cainjector" Namespace="cert-manager"
Creating ServiceAccount="cert-manager" Namespace="cert-manager"
Creating ServiceAccount="cert-manager-webhook" Namespace="cert-manager"
Creating ClusterRole="cert-manager-cainjector"
Creating ClusterRoleBinding="cert-manager-cainjector"
Creating Role="cert-manager-cainjector:leaderelection" Namespace="kube-system"
Creating RoleBinding="cert-manager-cainjector:leaderelection" Namespace="kube-system"
Creating ClusterRoleBinding="cert-manager-webhook:auth-delegator"
Creating RoleBinding="cert-manager-webhook:webhook-authentication-reader" Namespace="kube-system"
Creating ClusterRole="cert-manager-webhook:webhook-requester"
Creating Role="cert-manager:leaderelection" Namespace="kube-system"
Creating RoleBinding="cert-manager:leaderelection" Namespace="kube-system"
Creating ClusterRole="cert-manager-controller-issuers"
Creating ClusterRole="cert-manager-controller-clusterissuers"
Creating ClusterRole="cert-manager-controller-certificates"
Creating ClusterRole="cert-manager-controller-orders"
Creating ClusterRole="cert-manager-controller-challenges"
Creating ClusterRole="cert-manager-controller-ingress-shim"
Creating ClusterRoleBinding="cert-manager-controller-issuers"
Creating ClusterRoleBinding="cert-manager-controller-clusterissuers"
Creating ClusterRoleBinding="cert-manager-controller-certificates"
Creating ClusterRoleBinding="cert-manager-controller-orders"
Creating ClusterRoleBinding="cert-manager-controller-challenges"
Creating ClusterRoleBinding="cert-manager-controller-ingress-shim"
Creating ClusterRole="cert-manager-view"
Creating ClusterRole="cert-manager-edit"
Creating Service="cert-manager" Namespace="cert-manager"
Creating Service="cert-manager-webhook" Namespace="cert-manager"
Creating Deployment="cert-manager-cainjector" Namespace="cert-manager"
Creating Deployment="cert-manager" Namespace="cert-manager"
Creating Deployment="cert-manager-webhook" Namespace="cert-manager"
Creating APIService="v1beta1.webhook.cert-manager.io"
Creating MutatingWebhookConfiguration="cert-manager-webhook"
Creating ValidatingWebhookConfiguration="cert-manager-webhook"
Waiting for cert-manager to be available...
Installing Provider="cluster-api" Version="v0.3.3" TargetNamespace="capi-system"
Creating shared objects Provider="cluster-api" Version="v0.3.3"
Creating Namespace="capi-webhook-system"
Creating CustomResourceDefinition="clusters.cluster.x-k8s.io"
Creating CustomResourceDefinition="machinedeployments.cluster.x-k8s.io"
Creating CustomResourceDefinition="machinehealthchecks.cluster.x-k8s.io"
Creating CustomResourceDefinition="machinepools.exp.cluster.x-k8s.io"
Creating CustomResourceDefinition="machines.cluster.x-k8s.io"
Creating CustomResourceDefinition="machinesets.cluster.x-k8s.io"
Creating MutatingWebhookConfiguration="capi-mutating-webhook-configuration"
Creating Service="capi-webhook-service" Namespace="capi-webhook-system"
Creating Deployment="capi-controller-manager" Namespace="capi-webhook-system"
Creating Certificate="capi-serving-cert" Namespace="capi-webhook-system"
Creating Issuer="capi-selfsigned-issuer" Namespace="capi-webhook-system"
Creating ValidatingWebhookConfiguration="capi-validating-webhook-configuration"
Creating instance objects Provider="cluster-api" Version="v0.3.3" TargetNamespace="capi-system"
Creating Namespace="capi-system"
Creating Role="capi-leader-election-role" Namespace="capi-system"
Creating ClusterRole="capi-system-capi-aggregated-manager-role"
Creating ClusterRole="capi-system-capi-manager-role"
Creating ClusterRole="capi-system-capi-proxy-role"
Creating RoleBinding="capi-leader-election-rolebinding" Namespace="capi-system"
Creating ClusterRoleBinding="capi-system-capi-manager-rolebinding"
Creating ClusterRoleBinding="capi-system-capi-proxy-rolebinding"
Creating Service="capi-controller-manager-metrics-service" Namespace="capi-system"
Creating Deployment="capi-controller-manager" Namespace="capi-system"
Creating inventory entry Provider="cluster-api" Version="v0.3.3" TargetNamespace="capi-system"
Installing Provider="bootstrap-kubeadm" Version="v0.3.3" TargetNamespace="capi-kubeadm-bootstrap-system"
Creating shared objects Provider="bootstrap-kubeadm" Version="v0.3.3"
Creating CustomResourceDefinition="kubeadmconfigs.bootstrap.cluster.x-k8s.io"
Creating CustomResourceDefinition="kubeadmconfigtemplates.bootstrap.cluster.x-k8s.io"
Creating Service="capi-kubeadm-bootstrap-webhook-service" Namespace="capi-webhook-system"
Creating Deployment="capi-kubeadm-bootstrap-controller-manager" Namespace="capi-webhook-system"
Creating Certificate="capi-kubeadm-bootstrap-serving-cert" Namespace="capi-webhook-system"
Creating Issuer="capi-kubeadm-bootstrap-selfsigned-issuer" Namespace="capi-webhook-system"
Creating instance objects Provider="bootstrap-kubeadm" Version="v0.3.3" TargetNamespace="capi-kubeadm-bootstrap-system"
Creating Namespace="capi-kubeadm-bootstrap-system"
Creating Role="capi-kubeadm-bootstrap-leader-election-role" Namespace="capi-kubeadm-bootstrap-system"
Creating ClusterRole="capi-kubeadm-bootstrap-system-capi-kubeadm-bootstrap-manager-role"
Creating ClusterRole="capi-kubeadm-bootstrap-system-capi-kubeadm-bootstrap-proxy-role"
Creating RoleBinding="capi-kubeadm-bootstrap-leader-election-rolebinding" Namespace="capi-kubeadm-bootstrap-system"
Creating ClusterRoleBinding="capi-kubeadm-bootstrap-system-capi-kubeadm-bootstrap-manager-rolebinding"
Creating ClusterRoleBinding="capi-kubeadm-bootstrap-system-capi-kubeadm-bootstrap-proxy-rolebinding"
Creating Service="capi-kubeadm-bootstrap-controller-manager-metrics-service" Namespace="capi-kubeadm-bootstrap-system"
Creating Deployment="capi-kubeadm-bootstrap-controller-manager" Namespace="capi-kubeadm-bootstrap-system"
Creating inventory entry Provider="bootstrap-kubeadm" Version="v0.3.3" TargetNamespace="capi-kubeadm-bootstrap-system"
Installing Provider="control-plane-kubeadm" Version="v0.3.3" TargetNamespace="capi-kubeadm-control-plane-system"
Creating shared objects Provider="control-plane-kubeadm" Version="v0.3.3"
Creating CustomResourceDefinition="kubeadmcontrolplanes.controlplane.cluster.x-k8s.io"
Creating MutatingWebhookConfiguration="capi-kubeadm-control-plane-mutating-webhook-configuration"
Creating Service="capi-kubeadm-control-plane-webhook-service" Namespace="capi-webhook-system"
Creating Deployment="capi-kubeadm-control-plane-controller-manager" Namespace="capi-webhook-system"
Creating Certificate="capi-kubeadm-control-plane-serving-cert" Namespace="capi-webhook-system"
Creating Issuer="capi-kubeadm-control-plane-selfsigned-issuer" Namespace="capi-webhook-system"
Creating ValidatingWebhookConfiguration="capi-kubeadm-control-plane-validating-webhook-configuration"
Creating instance objects Provider="control-plane-kubeadm" Version="v0.3.3" TargetNamespace="capi-kubeadm-control-plane-system"
Creating Namespace="capi-kubeadm-control-plane-system"
Creating Role="capi-kubeadm-control-plane-leader-election-role" Namespace="capi-kubeadm-control-plane-system"
Creating Role="capi-kubeadm-control-plane-manager-role" Namespace="capi-kubeadm-control-plane-system"
Creating ClusterRole="capi-kubeadm-control-plane-system-capi-kubeadm-control-plane-manager-role"
Creating ClusterRole="capi-kubeadm-control-plane-system-capi-kubeadm-control-plane-proxy-role"
Creating RoleBinding="capi-kubeadm-control-plane-leader-election-rolebinding" Namespace="capi-kubeadm-control-plane-system"
Creating ClusterRoleBinding="capi-kubeadm-control-plane-system-capi-kubeadm-control-plane-manager-rolebinding"
Creating ClusterRoleBinding="capi-kubeadm-control-plane-system-capi-kubeadm-control-plane-proxy-rolebinding"
Creating Service="capi-kubeadm-control-plane-controller-manager-metrics-service" Namespace="capi-kubeadm-control-plane-system"
Creating Deployment="capi-kubeadm-control-plane-controller-manager" Namespace="capi-kubeadm-control-plane-system"
Creating inventory entry Provider="control-plane-kubeadm" Version="v0.3.3" TargetNamespace="capi-kubeadm-control-plane-system"
Installing Provider="infrastructure-docker" Version="v0.3.7" TargetNamespace="capd-system"
Creating shared objects Provider="infrastructure-docker" Version="v0.3.7"
Creating CustomResourceDefinition="dockerclusters.infrastructure.cluster.x-k8s.io"
Creating CustomResourceDefinition="dockermachines.infrastructure.cluster.x-k8s.io"
Creating CustomResourceDefinition="dockermachinetemplates.infrastructure.cluster.x-k8s.io"
Creating ValidatingWebhookConfiguration="capd-validating-webhook-configuration"
Creating instance objects Provider="infrastructure-docker" Version="v0.3.7" TargetNamespace="capd-system"
Creating Namespace="capd-system"
Creating Role="capd-leader-election-role" Namespace="capd-system"
Creating ClusterRole="capd-system-capd-manager-role"
Creating ClusterRole="capd-system-capd-proxy-role"
Creating RoleBinding="capd-leader-election-rolebinding" Namespace="capd-system"
Creating ClusterRoleBinding="capd-system-capd-manager-rolebinding"
Creating ClusterRoleBinding="capd-system-capd-proxy-rolebinding"
Creating Service="capd-controller-manager-metrics-service" Namespace="capd-system"
Creating Service="capd-webhook-service" Namespace="capd-system"
Creating Deployment="capd-controller-manager" Namespace="capd-system"
Creating Certificate="capd-serving-cert" Namespace="capd-system"
Creating Issuer="capd-selfsigned-issuer" Namespace="capd-system"
Creating inventory entry Provider="infrastructure-docker" Version="v0.3.7" TargetNamespace="capd-system"
```

Wait for all the pods to be up.

$ kubectl get pods -A
```
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
capd-system                         capd-controller-manager-75f5d546d7-frrm5                         2/2     Running   0          77s
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-5bb9bfdc46-mhbqz       2/2     Running   0          85s
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-77466c7666-t69m5   2/2     Running   0          81s
capi-system                         capi-controller-manager-5798474d9f-tp2c2                         2/2     Running   0          89s
capi-webhook-system                 capi-controller-manager-5d64dd9dfb-r6mb2                         2/2     Running   1          91s
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-7c78fff45-dmnlc        2/2     Running   0          88s
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-58465bb88f-c6j5q   2/2     Running   0          84s
cert-manager                        cert-manager-69b4f77ffc-8vchm                                    1/1     Running   0          117s
cert-manager                        cert-manager-cainjector-576978ffc8-frsxg                         1/1     Running   0          117s
cert-manager                        cert-manager-webhook-c67fbc858-qxrcj                             1/1     Running   1          117s
kube-system                         coredns-6955765f44-f28p7                                         1/1     Running   0          3m12s
kube-system                         coredns-6955765f44-nq5qk                                         1/1     Running   0          3m12s
kube-system                         etcd-capi-docker-control-plane                                   1/1     Running   0          3m25s
kube-system                         kindnet-nxm6k                                                    1/1     Running   0          3m12s
kube-system                         kube-apiserver-capi-docker-control-plane                         1/1     Running   0          3m25s
kube-system                         kube-controller-manager-capi-docker-control-plane                1/1     Running   0          3m25s
kube-system                         kube-proxy-5jmc5                                                 1/1     Running   0          3m12s
kube-system                         kube-scheduler-capi-docker-control-plane                         1/1     Running   0          3m25s
local-path-storage                  local-path-provisioner-7745554f7f-ms989                          1/1     Running   0          3m12s
```
Now, the management cluster is initialized with cluster api and cluster api
docker provider components.

$ kubectl get providers -A

```
NAMESPACE                           NAME                    TYPE   PROVIDER                 VERSION   WATCH NAMESPACE
capd-system                         infrastructure-docker          InfrastructureProvider   v0.3.7
capi-kubeadm-bootstrap-system       bootstrap-kubeadm              BootstrapProvider        v0.3.3
capi-kubeadm-control-plane-system   control-plane-kubeadm          ControlPlaneProvider     v0.3.3
capi-system                         cluster-api                    CoreProvider             v0.3.3
```

$ docker ps

```
CONTAINER ID        IMAGE                          COMMAND                  CREATED             STATUS              PORTS                                  NAMES
b9690cecdcf2        kindest/node:v1.18.2           "/usr/local/bin/entr…"   14 minutes ago      Up 14 minutes       127.0.0.1:32773->6443/tcp              capi-docker-control-plane
```


## Create your first workload cluster

$ airshipctl phase run controlplane --debug

```
[airshipctl] 2020/08/12 14:10:12 building bundle from kustomize path /tmp/airship/airshipctl/manifests/site/docker-test-site/target/controlplane
[airshipctl] 2020/08/12 14:10:12 Applying bundle, inventory id: kind-capi-docker-target-controlplane
[airshipctl] 2020/08/12 14:10:12 Inventory Object config Map not found, auto generating Invetory object
[airshipctl] 2020/08/12 14:10:12 Injecting Invetory Object: {"apiVersion":"v1","kind":"ConfigMap","metadata":{"creationTimestamp":null,"labels":{"cli-utils.sigs.k8s.io/inventory-id":"kind-capi-docker-target-controlplane"},"name":"airshipit-kind-capi-docker-target-controlplane","namespace":"airshipit"}}{nsfx:false,beh:unspecified} into bundle
[airshipctl] 2020/08/12 14:10:12 Making sure that inventory object namespace airshipit exists
configmap/airshipit-kind-capi-docker-target-controlplane-87efb53a created
cluster.cluster.x-k8s.io/dtc created
machinehealthcheck.cluster.x-k8s.io/dtc-mhc-0 created
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/dtc-control-plane created
dockercluster.infrastructure.cluster.x-k8s.io/dtc created
dockermachinetemplate.infrastructure.cluster.x-k8s.io/dtc-control-plane created
6 resource(s) applied. 6 created, 0 unchanged, 0 configured
machinehealthcheck.cluster.x-k8s.io/dtc-mhc-0 is NotFound: Resource not found
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/dtc-control-plane is NotFound: Resource not found
dockercluster.infrastructure.cluster.x-k8s.io/dtc is NotFound: Resource not found
dockermachinetemplate.infrastructure.cluster.x-k8s.io/dtc-control-plane is NotFound: Resource not found
configmap/airshipit-kind-capi-docker-target-controlplane-87efb53a is NotFound: Resource not found
cluster.cluster.x-k8s.io/dtc is NotFound: Resource not found
configmap/airshipit-kind-capi-docker-target-controlplane-87efb53a is Current: Resource is always ready
cluster.cluster.x-k8s.io/dtc is Current: Resource is current
machinehealthcheck.cluster.x-k8s.io/dtc-mhc-0 is Current: Resource is current
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/dtc-control-plane is Current: Resource is current
dockercluster.infrastructure.cluster.x-k8s.io/dtc is Current: Resource is current
dockermachinetemplate.infrastructure.cluster.x-k8s.io/dtc-control-plane is Current: Resource is current
all resources has reached the Current status
```

$ kubectl get pods -A
```
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
capd-system                         capd-controller-manager-75f5d546d7-frrm5                         2/2     Running   0          77s
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-5bb9bfdc46-mhbqz       2/2     Running   0          85s
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-77466c7666-t69m5   2/2     Running   0          81s
capi-system                         capi-controller-manager-5798474d9f-tp2c2                         2/2     Running   0          89s
capi-webhook-system                 capi-controller-manager-5d64dd9dfb-r6mb2                         2/2     Running   1          91s
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-7c78fff45-dmnlc        2/2     Running   0          88s
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-58465bb88f-c6j5q   2/2     Running   0          84s
cert-manager                        cert-manager-69b4f77ffc-8vchm                                    1/1     Running   0          117s
cert-manager                        cert-manager-cainjector-576978ffc8-frsxg                         1/1     Running   0          117s
cert-manager                        cert-manager-webhook-c67fbc858-qxrcj                             1/1     Running   1          117s
kube-system                         coredns-6955765f44-f28p7                                         1/1     Running   0          3m12s
kube-system                         coredns-6955765f44-nq5qk                                         1/1     Running   0          3m12s
kube-system                         etcd-capi-docker-control-plane                                   1/1     Running   0          3m25s
kube-system                         kindnet-nxm6k                                                    1/1     Running   0          3m12s
kube-system                         kube-apiserver-capi-docker-control-plane                         1/1     Running   0          3m25s
kube-system                         kube-controller-manager-capi-docker-control-plane                1/1     Running   0          3m25s
kube-system                         kube-proxy-5jmc5                                                 1/1     Running   0          3m12s
kube-system                         kube-scheduler-capi-docker-control-plane                         1/1     Running   0          3m25s
local-path-storage                  local-path-provisioner-7745554f7f-ms989                          1/1     Running   0          3m12s
```

$ kubectl logs capd-controller-manager-75f5d546d7-frrm5 -n capd-system --all-containers=true -f

```
I0812 21:11:24.761608       1 controller.go:272] controller-runtime/controller "msg"="Successfully Reconciled" "controller"="dockermachine" "name"="dtc-control-plane-zc5bw" "namespace"="default"
I0812 21:11:25.189401       1 controller.go:272] controller-runtime/controller "msg"="Successfully Reconciled" "controller"="dockermachine" "name"="dtc-control-plane-zc5bw" "namespace"="default"
I0812 21:11:26.219320       1 generic_predicates.go:38] controllers/DockerMachine "msg"="One of the provided predicates returned false, blocking further processing" "predicate"="ClusterUnpausedAndInfrastructureReady" "predicateAggregation"="All"
I0812 21:11:26.219774       1 cluster_predicates.go:143] controllers/DockerMachine "msg"="Cluster was not unpaused, blocking further processing" "cluster"="dtc" "eventType"="update" "namespace"="default" "predicate"="ClusterUpdateUnpaused"
I0812 21:11:26.222004       1 cluster_predicates.go:111] controllers/DockerMachine "msg"="Cluster infrastructure did not become ready, blocking further processing" "cluster"="dtc" "eventType"="update" "namespace"="default" "predicate"="ClusterUpdateInfraReady"
I0812 21:11:26.223003       1 generic_predicates.go:89] controllers/DockerMachine "msg"="All of the provided predicates returned false, blocking further processing" "predicate"="ClusterUnpausedAndInfrastructureReady" "predicateAggregation"="Any"
I0812 21:11:26.223239       1 generic_predicates.go:89] controllers/DockerMachine "msg"="All of the provided predicates returned false, blocking further processing" "predicate"="ClusterUnpausedAndInfrastructureReady" "predicateAggregation"="Any"
I0812 21:11:26.219658       1 cluster_predicates.go:143] controllers/DockerCluster "msg"="Cluster was not unpaused, blocking further processing" "cluster"="dtc" "eventType"="update" "namespace"="default" "predicate"="ClusterUpdateUnpaused"
I0812 21:11:26.229665       1 generic_predicates.go:89] controllers/DockerCluster "msg"="All of the provided predicates returned false, blocking further processing" "predicate"="ClusterUnpaused" "predicateAggregation"="Any"
```

$ kubectl get machines
```
NAME                      PROVIDERID                               PHASE
dtc-control-plane-p4fsx   docker:////dtc-dtc-control-plane-p4fsx   Running
```

$ airshipctl phase run workers --debug

```
[airshipctl] 2020/08/12 14:11:55 building bundle from kustomize path /tmp/airship/airshipctl/manifests/site/docker-test-site/target/worker
[airshipctl] 2020/08/12 14:11:55 Applying bundle, inventory id: kind-capi-docker-target-worker
[airshipctl] 2020/08/12 14:11:55 Inventory Object config Map not found, auto generating Invetory object
[airshipctl] 2020/08/12 14:11:55 Injecting Invetory Object: {"apiVersion":"v1","kind":"ConfigMap","metadata":{"creationTimestamp":null,"labels":{"cli-utils.sigs.k8s.io/inventory-id":"kind-capi-docker-target-worker"},"name":"airshipit-kind-capi-docker-target-worker","namespace":"airshipit"}}{nsfx:false,beh:unspecified} into bundle
[airshipctl] 2020/08/12 14:11:55 Making sure that inventory object namespace airshipit exists
configmap/airshipit-kind-capi-docker-target-worker-b56f83 created
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/dtc-md-0 created
machinedeployment.cluster.x-k8s.io/dtc-md-0 created
dockermachinetemplate.infrastructure.cluster.x-k8s.io/dtc-md-0 created
4 resource(s) applied. 4 created, 0 unchanged, 0 configured
dockermachinetemplate.infrastructure.cluster.x-k8s.io/dtc-md-0 is NotFound: Resource not found
configmap/airshipit-kind-capi-docker-target-worker-b56f83 is NotFound: Resource not found
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/dtc-md-0 is NotFound: Resource not found
machinedeployment.cluster.x-k8s.io/dtc-md-0 is NotFound: Resource not found
configmap/airshipit-kind-capi-docker-target-worker-b56f83 is Current: Resource is always ready
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/dtc-md-0 is Current: Resource is current
machinedeployment.cluster.x-k8s.io/dtc-md-0 is Current: Resource is current
dockermachinetemplate.infrastructure.cluster.x-k8s.io/dtc-md-0 is Current: Resource is current
```

$ kubectl get machines
```
NAME                       PROVIDERID                               PHASE
dtc-control-plane-p4fsx    docker:////dtc-dtc-control-plane-p4fsx   Running
dtc-md-0-94c79cf9c-8ct2g                                            Provisioning
```

$ kubectl logs capd-controller-manager-75f5d546d7-frrm5 -n capd-system --all-containers=true -f

```
I0812 21:10:14.071166       1 cluster_predicates.go:111] controllers/DockerMachine "msg"="Cluster infrastructure did not become ready, blocking further processing" "cluster"="dtc" "eventType"="update" "namespace"="default" "predicate"="ClusterUpdateInfraReady"
I0812 21:10:14.071204       1 generic_predicates.go:89] controllers/DockerMachine "msg"="All of the provided predicates returned false, blocking further processing" "predicate"="ClusterUnpausedAndInfrastructureReady" "predicateAggregation"="Any"
I0812 21:10:14.071325       1 generic_predicates.go:89] controllers/DockerMachine "msg"="All of the provided predicates returned false, blocking further processing" "predicate"="ClusterUnpausedAndInfrastructureReady" "predicateAggregation"="Any"
I0812 21:10:14.082937       1 generic_predicates.go:38] controllers/DockerMachine "msg"="One of the provided predicates returned false, blocking further processing" "predicate"="ClusterUnpausedAndInfrastructureReady" "predicateAggregation"="All"
I0812 21:10:14.082981       1 cluster_predicates.go:143] controllers/DockerMachine "msg"="Cluster was not unpaused, blocking further processing" "cluster"="dtc" "eventType"="update" "namespace"="default" "predicate"="ClusterUpdateUnpaused"
I0812 21:10:14.082994       1 cluster_predicates.go:143] controllers/DockerCluster "msg"="Cluster was not unpaused, blocking further processing" "cluster"="dtc" "eventType"="update" "namespace"="default" "predicate"="ClusterUpdateUnpaused"
I0812 21:10:14.083012       1 cluster_predicates.go:111] controllers/DockerMachine "msg"="Cluster infrastructure did not become ready, blocking further processing" "cluster"="dtc" "eventType"="update" "namespace"="default" "predicate"="ClusterUpdateInfraReady"
I0812 21:10:14.083024       1 generic_predicates.go:89] controllers/DockerCluster "msg"="All of the provided predicates returned false, blocking further processing" "predicate"="ClusterUnpaused" "predicateAggregation"="Any"
I0812 21:10:14.083036       1 generic_predicates.go:89] controllers/DockerMachine "msg"="All of the provided predicates returned false, blocking further processing"
```
$ kubectl get machines
```
NAME                       PROVIDERID                                PHASE
dtc-control-plane-p4fsx    docker:////dtc-dtc-control-plane-p4fsx    Running
dtc-md-0-94c79cf9c-8ct2g   docker:////dtc-dtc-md-0-94c79cf9c-8ct2g   Running
```

$ kubectl --namespace=default get secret/dtc-kubeconfig -o jsonpath={.data.value} | base64 --decode > ./dtc.kubeconfig

$ kubectl get nodes --kubeconfig dtc.kubeconfig

```
NAME                           STATUS   ROLES    AGE     VERSION
dtc-dtc-control-plane-p4fsx    Ready    master   5m45s   v1.18.6
dtc-dtc-md-0-94c79cf9c-8ct2g   Ready    <none>   4m45s   v1.18.6
```

$ kubectl get pods -A --kubeconfig dtc.kubeconfig
```
NAMESPACE     NAME                                                  READY   STATUS    RESTARTS   AGE
kube-system   calico-kube-controllers-59b699859f-xp8dv              1/1     Running   0          5m40s
kube-system   calico-node-5drwf                                     1/1     Running   0          5m39s
kube-system   calico-node-bqw5j                                     1/1     Running   0          4m53s
kube-system   coredns-6955765f44-8kg27                              1/1     Running   0          5m40s
kube-system   coredns-6955765f44-lqzzq                              1/1     Running   0          5m40s
kube-system   etcd-dtc-dtc-control-plane-p4fsx                      1/1     Running   0          5m49s
kube-system   kube-apiserver-dtc-dtc-control-plane-p4fsx            1/1     Running   0          5m49s
kube-system   kube-controller-manager-dtc-dtc-control-plane-p4fsx   1/1     Running   0          5m49s
kube-system   kube-proxy-cjcls                                      1/1     Running   0          5m39s
kube-system   kube-proxy-fkvpc                                      1/1     Running   0          4m53s
kube-system   kube-scheduler-dtc-dtc-control-plane-p4fsx            1/1     Running   0          5m49s
```

## Tear Down Clusters

The command shown below can be used to delete the control plane, worker nodes and all
associated cluster resources

$ airshipctl phase render controlplane -k Cluster | kubectl delete -f -

```
cluster.cluster.x-k8s.io "dtc" deleted
```

To delete the kind cluster, use the following command:

$ kind delete cluster --name capi-docker

```
Deleting cluster "capi-docker" ...
```

## Reference

### Provider Manifests

Provider Configuration is referenced from
[config](https://github.com/kubernetes-sigs/cluster-api/tree/master/test/infrastructure/docker/config)
Cluster API does not support docker out of the box. Therefore, the metadata
infromation is added using files in `airshipctl/manifests/function/capd/data`

$ tree airshipctl/manifests/function/capd

```
airshipctl/manifests/function/capd
└── v0.3.7
    ├── certmanager
    │   ├── certificate.yaml
    │   ├── kustomization.yaml
    │   └── kustomizeconfig.yaml
    ├── crd
    │   ├── bases
    │   │   ├── infrastructure.cluster.x-k8s.io_dockerclusters.yaml
    │   │   ├── infrastructure.cluster.x-k8s.io_dockermachines.yaml
    │   │   └── infrastructure.cluster.x-k8s.io_dockermachinetemplates.yaml
    │   ├── kustomization.yaml
    │   ├── kustomizeconfig.yaml
    │   └── patches
    │       ├── cainjection_in_dockerclusters.yaml
    │       ├── cainjection_in_dockermachines.yaml
    │       ├── webhook_in_dockerclusters.yaml
    │       └── webhook_in_dockermachines.yaml
    ├── data
    │   ├── kustomization.yaml
    │   └── metadata.yaml
    ├── default
    │   ├── kustomization.yaml
    │   └── namespace.yaml
    ├── kustomization.yaml
    ├── manager
    │   ├── kustomization.yaml
    │   ├── manager_auth_proxy_patch.yaml
    │   ├── manager_image_patch.yaml
    │   ├── manager_prometheus_metrics_patch.yaml
    │   ├── manager_pull_policy.yaml
    │   └── manager.yaml
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

10 directories, 37 files
```

### Cluster Templates

`manifests/function/k8scontrol-capd` contains cluster.yaml, controlplane.yaml
templates referenced from
[cluster-template](https://github.com/kubernetes-sigs/cluster-api/blob/master/test/e2e/data/infrastructure-docker/cluster-template.yaml)


| Template Name     | CRDs |
| ----------------- | ---- |
| cluster.yaml      |   Cluster, DockerCluster   |
| controlplane.yaml | KubeadmControlPlane, DockerMachineTemplate, MachineHealthCheck  |

$ tree airshipctl/manifests/function/k8scontrol-capd

```
airshipctl/manifests/function/k8scontrol-capd
├── cluster.yaml
├── controlplane.yaml
└── kustomization.yaml
```

`airshipctl/manifests/function/workers-capd` contains workers.yaml referenced
from
[cluster-template](https://github.com/kubernetes-sigs/cluster-api/blob/master/test/e2e/data/infrastructure-docker/cluster-template.yaml)

| Template Name     | CRDs |
| ----------------- | ---- |
| workers.yaml | KubeadmConfigTemplate, MachineDeployment, DockerMachineTemplate |

$ tree airshipctl/manifests/function/workers-capd

```
airshipctl/manifests/function/workers-capd
├── kustomization.yaml
└── workers.yaml
```

### Test Site Manifests

#### docker-test-site/shared

`airshipctl cluster init` uses `airshipctl/manifests/site/docker-test-site/shared/clusterctl`
to initialize management cluster with defined provider components and version.

`$ tree airshipctl/manifests/site/docker-test-site/shared`

```
/tmp/airship/airshipctl/manifests/site/docker-test-site/shared
└── clusterctl
    ├── clusterctl.yaml
    └── kustomization.yaml

1 directory, 2 files
```

#### docker-test-site/target

There are 3 phases currently available in `docker-test-site/target`.

| Phase Name  | Purpose |
| ----------- | --------- |
| controlplane | Patches templates in manifests/function/k8scontrol-capd |
| workers   | Patches template in manifests/function/workers-capd |
| initinfra | Simply calls `docker-test-site/shared/clusterctl` |

Note: `airshipctl cluster init` initializes all the provider components
including the docker infrastructure provider component. As a result, `airshipctl
phase run initinfra-ephemeral` is not used.

At the moment, `phase initinfra` is only present for two reasons:
- `airshipctl` complains if the phase is not found
- `validate site docs to pass`

#### Patch Merge Strategy

Json patch to patch `control plane machine count` is applied on templates in `manifests/function/k8scontrol-capd`
from `airshipctl/manifests/site/docker-test-site/target/controlplane` when `airshipctl phase run
controlplane` is executed

Json patch to patch `workers machine count` is applied on templates in `manifests/function/workers-capd`
from `airshipctl/manifests/site/docker-test-site/target/workers` when `airshipctl phase run
workers` is executed.

| Patch Name | Purpose  |
| ---------- | -------- |
| controlplane/machine_count.json | patches control plane machine count in template function/k8scontrol-capd |
| workers/machine_count.json      | patches workers machine count in template function/workers-capd |

$ tree airshipctl/manifests/site/docker-test-site/target/

```
airshipctl/manifests/site/docker-test-site/target/
├── controlplane
│   ├── kustomization.yaml
│   └── machine_count.json
├── initinfra
│   └── kustomization.yaml
└── workers
    ├── kustomization.yaml
    └── machine_count.json

3 directories, 7 files
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

## Special Instructions

Swap was disabled on the VM using `sudo swapoff -a`
