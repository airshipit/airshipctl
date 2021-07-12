# Deploy a Virtualized Environment

This guide demonstrates how to deploy the virtualized airshipctl and treasuremap
gating environments. While both environments provide an under-the-hood
demonstration of how Airship works, they are not required for development. We
recommend that developers testing changes consider if using Airship-in-a-Pod
(AIAP) or using Kubectl to apply rendered manifests to external Kubernetes
clusters better suits their needs before proceeding.

## Pre-requisites

The following are pre-requisites for deploying virtualized Airship environments:

  - Minimum 20 GB RAM
  - Minimum 8 vCPUs
  - Minimum 100 GB storage
  - Ubuntu 18.04
  - Nested virtualization (if your host is a virtual machine)

## Select an environment

This guide supports the airshipctl `test-site` and treasuremap `test-site`.

## Clone repositories

1. Clone airshipctl:

```sh
git clone https://opendev.org/airship/airshipctl.git
```

2.  If you are deploying a Treasuremap site, clone Treasuremap to the same
parent directory as airshipctl:

```sh
git clone https://opendev.org/airship/treasuremap.git
```

### Proxy Setup

If your organization requires development behind a proxy server, you will need
to define the following environment variables with your organization's
information:

```sh
HTTP_PROXY=http://username:password@host:port
HTTPS_PROXY=http://username:password@host:port
NO_PROXY="localhost,127.0.0.1,10.23.0.0/16,10.96.0.0/12"
PROXY=http://username:password@host:port
USE_PROXY=true
```

`10.23.0.0/16` encapsulates the range of addresses used by airshipctl
development environment virtual machines, and `10.96.0.0/12` is the Kubernetes
service CIDR.

### Configure DNS servers

If you cannot reach the Google DNS servers from your local environment, you will
need to replace the Google DNS servers with your DNS servers in your site's
`NetworkCatalogue`.

For the airshipctl test-site, update
`airshipctl/manifests/type/gating/shared/catalogues/networking.yaml`. For
the treasuremap test-site, update
`treasuremap/manifests/site/test-site/target/catalogues/networking.yaml`.

### Configure test encryption key

Execute the following to download and export the test encryption key and fingerprint.

 ```sh
curl -fsSL -o /tmp/key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
export SOPS_IMPORT_PGP="$(cat /tmp/key.asc)"
export SOPS_PGP_FP="FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4"
```

### Run the setup scripts

#### Install required packages and configure Ansible

From the root of the airshipctl repository, run:

```sh
./tools/gate/00_setup.sh
```

#### Create virsh VMs

From the root of the airshipctl repository, run:

```sh
./tools/gate/10_build_gate.sh
```

#### Generate an airshipctl configuration file

For the airshipctl test-site, execute the following from the root of the
airshipctl repository:

```sh
./tools/deployment/22_test_configs.sh
```

For the treasuremap test-site, execute the following from the root of the treasuremap repository:

```sh
./tools/deployment/airship-core/22_test_configs.sh
```

#### Download deployment manifests (documents)

For the airshipctl test-site, execute the following from the root of the
airshipctl repository:

```sh
./tools/deployment/23_pull_documents.sh
```

For the treasuremap test-site, execute the following from the root of the treasuremap repository:

```sh
./tools/deployment/airship-core/23_pull_documents.sh
```

#### Generate site secrets

For the airshipctl test-site, execute the following from the root of the
airshipctl repository:

```sh
./tools/deployment/23_generate_secrets.sh
```

For the treasuremap test-site, execute the following from the root of the treasuremap repository:

```sh
./tools/deployment/airship-core/23_generate_secrets.sh
```

#### Build ephemeral node ISO and target cluster control plane and data plane images

For the airshipctl test-site, execute the following from the root of the
airshipctl repository:

```sh
./tools/deployment/24_build_images.sh
```

For the treasuremap test-site, execute the following from the root of the treasuremap repository:

```sh
./tools/deployment/airship-core/24_build_images.sh
```

#### Deploy the ephemeral and target clusters

For the airshipctl test-site, execute the following from the root of the
airshipctl repository:

```sh
./tools/deployment/25_deploy_gating.sh
```

For the treasuremap test-site, execute the following from the root of the treasuremap repository:

```sh
./tools/deployment/airship-core/25_deploy_gating.sh
```

### Troubleshooting

### Validate Ephemeral Cluster is Operational

If the `25_deploy_gating.sh` script fails with:

```
19: Retrying to reach the apiserver
+ sleep 60
+ '[' 19 -ge 30 ]
+ timeout 20 kubectl --context ephemeral-cluster get node -o name
+ wc -l
The connection to the server 10.23.25.101:6443 was refused - did you specify the right host or port?
```

or a similar error, validate that the ephemeral cluster is reachable:

```sh
kubectl --kubeconfig ~/.airship/kubeconfig --context ephemeral-cluster get pods --all-namespaces
```

The command should yield output that looks like this:

```
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-556678c94-hngzj        2/2     Running   0          50s
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-556d47dffd-qljht   2/2     Running   0          47s
capi-system                         capi-controller-manager-67859f6b78-2tgcx                         2/2     Running   0          54s
capi-webhook-system                 capi-controller-manager-5c785c685c-fds47                         2/2     Running   0          55s
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-77658d7745-5bb7z       2/2     Running   0          52s
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-74dcf8b9c-ds4l7    2/2     Running   0          49s
capi-webhook-system                 capm3-controller-manager-568747bbbb-zld5v                        2/2     Running   0          45s
capm3-system                        capm3-controller-manager-698c6d6df9-n72cf                        2/2     Running   0          42s
cert-manager                        cert-manager-578cd6d964-lznfq                                    1/1     Running   0          76s
cert-manager                        cert-manager-cainjector-5ffff9dd7c-h9v6l                         1/1     Running   0          76s
cert-manager                        cert-manager-webhook-556b9d7dfd-hvvfs                            1/1     Running   0          75s
hardware-classification             hardware-classification-controller-manager-776b5f66f8-6z9xl      2/2     Running   0          10m
kube-system                         calico-kube-controllers-94b8f9766-6cl6l                          1/1     Running   0          10m
kube-system                         calico-node-dw6c8                                                1/1     Running   0          10m
kube-system                         coredns-66bff467f8-57wpm                                         1/1     Running   0          13m
kube-system                         coredns-66bff467f8-lbfw2                                         1/1     Running   0          13m
kube-system                         etcd-ephemeral                                                   1/1     Running   0          13m
kube-system                         kube-apiserver-ephemeral                                         1/1     Running   0          13m
kube-system                         kube-controller-manager-ephemeral                                1/1     Running   0          13m
kube-system                         kube-proxy-whdhw                                                 1/1     Running   0          13m
kube-system                         kube-scheduler-ephemeral                                         1/1     Running   0          13m
metal3                              ironic-5d95b49d6c-lr6b2                                          4/4     Running   0          10m
metal3                              metal3-baremetal-operator-84f9df77fb-zq4qv                       3/3     Running   0          10m
```

One of the most common reasons for a failed ephemeral cluster deployment is
because a user is behind a corporate firewall and has not configured the proxy
and DNS settings required for the virtual machines to reach the internet. If the
ephemeral cluster is not reachable, we recommend validating that you have
configured your environment's proxy and DNS settings above.

#### Validate Target Cluster is Operational

Similarly, you can validate that your target cluster is operational using the context `target-cluster`:

```sh
kubectl --kubeconfig ~/.airship/kubeconfig --context target-cluster get pods --all-namespaces
```

```
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-556678c94-svqmn        2/2     Running   0          56s
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-556d47dffd-z28lq   2/2     Running   0          46s
capi-system                         capi-controller-manager-67859f6b78-x4k25                         2/2     Running   0          64s
capi-webhook-system                 capi-controller-manager-5c785c685c-9t58p                         2/2     Running   0          69s
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-77658d7745-wv8bt       2/2     Running   0          62s
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-74dcf8b9c-rskqk    2/2     Running   0          51s
capi-webhook-system                 capm3-controller-manager-568747bbbb-gpvqc                        2/2     Running   0          35s
capm3-system                        capm3-controller-manager-698c6d6df9-n6pfm                        2/2     Running   0          27s
cert-manager                        cert-manager-578cd6d964-nkgj7                                    1/1     Running   0          99s
cert-manager                        cert-manager-cainjector-5ffff9dd7c-ps62z                         1/1     Running   0          99s
cert-manager                        cert-manager-webhook-556b9d7dfd-2spgg                            1/1     Running   0          99s
flux-system                         helm-controller-cbb96fc8d-7vh96                                  1/1     Running   0          11m
flux-system                         source-controller-64f4b85496-zfj6w                               1/1     Running   0          11m
hardware-classification             hardware-classification-controller-manager-776b5f66f8-zd5rt      2/2     Running   0          11m
kube-system                         calico-kube-controllers-94b8f9766-9r2cn                          1/1     Running   0          11m
kube-system                         calico-node-6gfpc                                                1/1     Running   0          11m
kube-system                         coredns-66bff467f8-4gggz                                         1/1     Running   0          16m
kube-system                         coredns-66bff467f8-qgbhj                                         1/1     Running   0          16m
kube-system                         etcd-node01                                                      1/1     Running   0          16m
kube-system                         kube-apiserver-node01                                            1/1     Running   0          16m
kube-system                         kube-controller-manager-node01                                   1/1     Running   0          16m
kube-system                         kube-proxy-ch6z9                                                 1/1     Running   0          16m
kube-system                         kube-scheduler-node01                                            1/1     Running   0          16m
metal3                              ironic-5d95b49d6c-8xwcx                                          4/4     Running   0          11m
metal3                              metal3-baremetal-operator-84f9df77fb-25h4w                       3/3     Running   0          11m
```

#### Restart VMs

In case a restart of your host causes the Airship VMs to not restart, execute
the commands below to restart your VMs.

```
$ sudo virsh list --all
Id    Name                           State
----------------------------------------------------
-     air-ephemeral                  shut off
-     air-target-1                   shut off
-     air-worker-1                   shut off
$ virsh net-start air_nat
Network air_nat started
$ virsh net-start air_prov
Network air_prov started
$ virsh start air-target-1
Domain air-target-1 started
$ virsh start air-worker-1
Domain air-worker-1 started
$ sudo virsh list --all
Id    Name                           State
----------------------------------------------------
3     air-target-1                   running
4     air-worker-1                   running
```

### Re-deploying

In case you need to re-run the deployment from a clean state, we recommend
running the script below from the root of the airshipctl repository beforehand.

```sh
sudo ./tools/deployment/clean.sh
```
