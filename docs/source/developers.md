# Developer's Guide

This guide explains how to set up your environment for developing on
airshipctl.

## Environment expectations

- Git
- Go 1.13
- Docker

### Installing Git

Instructions to install Git are [here][11].

### Installing Go 1.13

Instructions to install Golang are [here][12].

The `make test` verification step requires the GNU Compiler Collection (gcc) to be installed.

To install the GNU Compiler Collection (gcc):

```sh
sudo apt-get install gcc
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

### DNS Configuration

If you cannot reach the Google DNS servers from your local environment, add your
DNS servers to
`manifests/type/airship-core/shared/catalogues/common-networking.yaml` in place
of the Google ones.

## Clone airshipctl code

Run the following command to download the latest airshipctl code:

```sh
git clone https://opendev.org/airship/airshipctl.git
```

NOTE: The airshipctl application is a Go module. This means that there is no
need to clone the repository into the $GOPATH directory in order to build it.
You should be able to build it from any directory as long as $GOPATH is
defined correctly.

### Installing Docker & Other Tools

Prior to building the airshipctl binary, ensure you have Docker,
Ansible & other tools installed in your environment.

There is a script in the airshipctl directory named `00_setup.sh` which can be
run to download all the required binaries and packages. This script code can be
viewed [here][1].

Standalone instructions to install Docker are [here][13]. This is not necessary
if you run `00_setup.sh`.

## Building airshipctl

Run the following command to build the airshipctl binary:

```sh
make build
```

This will compile airshipctl and place the resulting binary into the bin
directory.

To test the build, including linting and coverage reports, run:

```sh
make test
```

To run all tests in a containerized environment, run:

```sh
make docker-image-test-suite
```

## Docker Images

To build an `airshipctl` Docker image, run:

```sh
make docker-image
```

Pre-built images are already available at [quay.io][2]. Moreover, in the
directory `airshipctl/tools/gate/`, different scripts are present which will
run and download all the required images. The script [10_build_gate.sh][3]
will download all the required images.

## Contribution Guidelines

We welcome contributions. This project has set up some guidelines in order to
ensure that

- code quality remains high
- the project remains consistent, and
- contributions follow the open source legal requirements.

Our intent is not to burden contributors, but to build elegant and
high-quality open source code so that our users will benefit.
Make sure you have read and understood the main airshipctl
[Contributing Guide][4].

## Structure of the Code

The code for the airshipctl project is organized as follows:

- The client-facing code is located in `cmd/`. Code inside of `cmd/` is not
designed for library reuse.
- Shared libraries are stored in `pkg/`.
- Both commands and shared libraries may require test data fixtures. These
should be placed in a `testdata/` subdirectory within the command or library.
- The `testutil/` directory contains functions that are helpful for unit
tests.
- The `zuul.d/` directory contains Zuul YAML definitions for CI/CD jobs to
run.
- The `playbooks/` directory contains playbooks that the Zuul CI/CD jobs will
run.
- The `tools/` directory contains scripts used by the Makefile and CI/CD
pipeline.
- The `tools/gate` directory consists of different scripts. These scripts
will setup the environment as per requirements and install all the required
packages and binaries. This will also download all the required docker images.
- The `docs/` folder is used for documentation and examples.
- Go dependencies are managed by `go mod` and stored in `go.mod` and `go.sum`

## Git Conventions

We use Git for our version control system. The `master` branch is the home of
the current development candidate. Releases are tagged.
We accept changes to the code via Gerrit pull requests. One workflow for doing
this is as follows:

1. `git clone` the `airshipctl` repository. For this run the command:

    ```sh
    git clone https://opendev.org/airship/airshipctl.git
    ```

2. Use [OpenDev documentation][5] to setup Gerrit with the repo.

3. When set, use [this guide][6] to learn the OpenDev development workflow,
in a sandbox environment. You can then apply the learnings to start developing
airshipctl.

## Go Conventions

We follow the Go coding style standards very closely. Typically, running
`goimports -w -local opendev.org/airship/airshipctl ./` will make your code
beautiful for you.

We also typically follow the conventions of `golangci-lint`.
Read more:

- Effective Go [introduces formatting][7].
- The Go Wiki has a great article on [formatting][8].

## Testing

In order to ensure that all package unit tests follow the same standard and
use the same frameworks, airshipctl has a document outlining
[specific test guidelines][9] maintained separately.
Moreover, there are few scripts in directory `tools/gate` which run different
tests. The script [20_run_gate_runner.sh][10] will generate airshipctl config
file, deploy ephemeral cluster with infra and cluster API, deploy target cluster
and verify all control pods.

## Steps to build a Local All-in-one VM Environment

Pre-requisites:
Make sure the following conditions are met:
1. Nested Virtualization enabled on the Host
2. A Virtual Machine with 20 GB RAM, 4 vCPU and 60GB Disk and Ubuntu 18.04 Installed.
3. Clone the following repo -
    - git clone https://opendev.org/airship/airshipctl.git
4. Download test security key and add it to environment variable.
   - curl -fsSL -o /tmp/key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
   - export SOPS_IMPORT_PGP="$(cat /tmp/key.asc)"
   - export SOPS_PGP_FP="FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4"
5. Execute the following scripts one by one
    1. ./tools/gate/00_setup.sh
    2. ./tools/gate/10_build_gate.sh
    3. sudo -E ./tools/deployment/01_install_kubectl.sh
    4. sudo -E ./tools/deployment/02_install_clusterctl.sh
    5. sudo -E ./tools/deployment/22_test_configs.sh
    6. sudo -E ./tools/deployment/23_pull_documents.sh
    7. sudo -E ./tools/deployment/24_build_ephemeral_iso.sh
    8. sudo -E ./tools/deployment/25_deploy_ephemeral_node.sh
    9. sudo -E ./tools/deployment/26_deploy_metal3_capi_ephemeral_node.sh
    10. sudo -E ./tools/deployment/30_deploy_controlplane.sh
    11. sudo -E ./tools/deployment/31_deploy_initinfra_target_node.sh
    12. sudo -E ./tools/deployment/32_cluster_init_target_node.sh
    13. sudo -E ./tools/deployment/33_cluster_move_target_node.sh
    14. sudo -E ./tools/deployment/34_deploy_worker_node.sh
    15. sudo -E ./tools/deployment/35_deploy_workload.sh

6. How to verify the ephemeral cluster and target cluster is deployed successfully
    Validate Ephemeral Cluster is Operational:
    ```Markdown
    kubectl --kubeconfig /home/user/.airship/kubeconfig --context ephemeral-cluster get pods --all-namespaces
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

      Validate Target Cluster is Operational:

    ```Markdown
    kubectl --kubeconfig /home/user/.airship/kubeconfig --context target-cluster get pods --all-namespaces
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

7. How to deploy Workloads
    Once the Target is Operational, Workloads can be deployed on the Target Cluster.
    A small demo workload can be deployed using ./tools/deployment/35_deploy_workload.sh.This demo includes ingress as a workload.
    To verify execute kubectl command as below:
    ```Markdown
    $ kubectl --kubeconfig /home/user/.airship/kubeconfig --context target-cluster get pods -n ingress

    NAME                                                    READY   STATUS    RESTARTS   AGE
    ingress-ingress-nginx-controller-7d5d89f47d-p8hms       1/1     Running   1          6d19h
    ingress-ingress-nginx-defaultbackend-6c49f4ff7f-nzsjw   1/1     Running   1          6d19h
    ```
    Additional Workloads can be defined under  ~/airshipctl/manifests/site/test-site/target/workload/kustomization.yaml which specifies the resources as below
    ```Markdown
    $ pwd
    /home/user/airshipctl/manifests/site/test-site/target/workload
    $ cat kustomization.yaml
    resources:
    - ../../../../function/airshipctl-base-catalogues
    - ../../../../type/gating/target/workload
    transformers:
    - ../../../../type/gating/target/workload/ingress/replacements
    $ pwd
    /home/user/airshipctl/manifests/type/gating/target/workload
    $ ll
    total 16
    drwxrwxr-x 3 user user 4096 Nov 16 17:02 ./
    drwxrwxr-x 3 user user 4096 Nov 16 17:02 ../
    drwxrwxr-x 3 user user 4096 Nov 16 17:02 ingress/
    -rw-rw-r-- 1 user user   23 Nov 16 17:02 kustomization.yaml
    ```
8. In case the All-in-One-VM is restarted and the nested VMs do not get restarted automatically simply execute the below steps to make the Target Cluster up again.
    ```Markdown
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

9. In case the deployment needs to be cleaned and rerun again, run the below script.
    - sudo ./tools/deployment/clean.sh


[1]: https://github.com/airshipit/airshipctl/blob/master/tools/gate/00_setup.sh
[2]: https://quay.io/airshipit/airshipctl
[3]: https://github.com/airshipit/airshipctl/blob/master/tools/gate/10_build_gate.sh
[4]: https://github.com/airshipit/airshipctl/blob/master/CONTRIBUTING.md
[5]: https://docs.openstack.org/contributors/common/setup-gerrit.html
[6]: https://docs.opendev.org/opendev/infra-manual/latest/sandbox.html
[7]: https://golang.org/doc/effective_go.html#formatting
[8]: https://github.com/golang/go/wiki/CodeReviewComments
[9]: https://github.com/airshipit/airshipctl/blob/master/docs/source/testing-guidelines.md
[10]: https://github.com/airshipit/airshipctl/blob/master/tools/gate/20_run_gate_runner.sh
[11]: https://git-scm.com/book/en/v2/Getting-Started-Installing-Git
[12]: https://golang.org/doc/install
[13]: https://docs.docker.com/get-docker/
