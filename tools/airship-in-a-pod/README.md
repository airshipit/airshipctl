# Airship in a Pod

Airship in a pod is a Kubernetes pod definition which describes all of the
components required to deploy a fully functioning Airship 2 deployment. The pod
consists of the following "Task" containers:

* `artifact-setup`: This container builds the airshipctl binary and makes it
  available to the other containers. Also, based on the configuration provided
  in the airship-in-a-pod manifest, airshipctl/treasuremap(based on the usecase) git repositories
  will be downloaded and the required tag or commitId will be checked out.
* `infra-builder`: This container creates the various virtual networks and
  machines required for an Airship deployment
* `runner`: The runner container is the "meat" of the pod, and executes the
  deployment

The pod also contains the following "Support" containers:

* `libvirt`: This provides virtualisation
* `sushy-tools`: This is used for its BMC emulator
* `docker-in-docker`: This is used for nesting containers*
* `nginx`: This is used for image hosting


## Prerequisites

### Nested Virtualisation

If deployment is done on a VM, ensure that nested virtualization is enabled.

### Setup shared directory

Create the following directory with appropriate r+w permissions.

```
sudo mkdir /opt/.airship
```

### Environment variable setup

If you are within a proxy environment, ensure that the following environment
variables are defined, and NO_PROXY has the IP address which minikube uses.
For retrieving minikube ip refer: [minikube-ip](https://minikube.sigs.k8s.io/docs/commands/ip/)

```
export HTTP_PROXY=http://username:password@host:port
export HTTPS_PROXY=http://username:password@host:port
export NO_PROXY="localhost,127.0.0.1,10.23.0.0/16,10.96.0.0/12,10.1.1.44"
export PROXY=http://username:password@host:port
export USE_PROXY=true
export http_proxy=http://username:password@host:port
export https_proxy=http://username:password@host:port
export no_proxy="localhost,127.0.0.1,10.23.0.0/16,10.96.0.0/12,10.1.1.44"
export proxy=http://username:password@host:port
```

### To start minikube

Within the environment, with appropriate env variables set, run the following command.

```
sudo -E minikube start --driver=none

```
Refer [minikube](https://minikube.sigs.k8s.io/docs/start/)for more details.

## Usage

Since Airship in a Pod is just a pod definition, deploying and using it is as
simple as deploying and using any Kubernetes pod with kustomize tool.

###  Pod configuration

The below section provides steps to configure site with [airshipctl](https://github.com/airshipit/airshipctl)/[treasuremap](https://github.com/airshipit/treasuremap) manifests.

#### For airshipctl

Within the examples/airshipctl directory, update the existing patchset.yaml
file to reflect the airshipctl branch reference as required.

filepath : airshipctl/tools/airship-in-a-pod/examples/airshipctl/patchset.yaml


```
- op: replace
  path: "/spec/containers/4/env/4/value"
  value: <branch reference>

```

#### For treasuremap

For treasuremap related manifests, use the patchset.yaml from
examples/treasuremap and  update the following to reflect
the treasuremap branch reference and the pinned airshipctl reference
as required. The pinned airshipctl reference is the tag/commitId with
which treasuremap is tested and found working satisfactorily. This
could be found listed as 'AIRSHIPCTL_REF' attribute under the zuul.d
directory of treasuremap repository.

filepath : airshipctl/tools/airship-in-a-pod/examples/treasuremap/patchset.yaml

```
- op: replace
  path: "/spec/containers/4/env/4/value"
  value: <branch reference>

- op: replace
  path: "/spec/containers/4/env/6/value"
  value: <airshipctl_ref>

```

For more details, please consult the examples directory.

### Deploy the Pod

Once patchset.yaml for either airshipctl/treasuremap is ready, run the following
command against the running minikube cluster as shown below.

For example to run AIAP with treasuremap manifests, run the following commands.

```
cd tools/airship-in-a-pod/examples/{either airshipctl or treasuremap}
kustomize build . | kubectl apply -f -

```

### View Pod Logs

```
kubectl logs airship-in-a-pod -c $CONTAINER
```

### Interact with the Pod

```
kubectl exec -it airship-in-a-pod -c $CONTAINER -- bash
```

where `$CONTAINER` is one of the containers listed above.

### Inspect Cluster

Once AIAP is fully installed with a target cluster (air-target-1 and air-worker-1 nodes)
installed and running, the cluster could be monitored using the following steps.

#### Log into the runner container

```
kubectl exec -it airship-in-a-pod -c runner -- bash
```

Run the .profile file using the following command to run kubectl/airshipctl commands
as below.

```
source ~/.profile

```

To run kubectl commands on Target cluster, use --kubeconfig and --context params
within kubectl as below.

```
kubectl --kubeconfig /root/.airship/kubeconfig --context target-cluster get pods -A'
```


### Output

Airship-in-a-pod produces the following outputs:

* The airshipctl repo, manifest repo, and airshipctl binary used with the deployment.
* A tarball containing the generated ephemeral ISO, as well as the
  configuration used during generation.

These artifacts are placed at `ARTIFACTS_DIR` (defaults to /opt/aiap-artifacts`).


### Caching

As it can be cumbersome and time-consuming to build and rebuild binaries and
images, some options are made available for caching. A developer may re-use
artifacts from previous runs (or provide their own) by placing them in
`CACHE_DIR` (defaults to `/opt/aiap-cache`). Special care is needed for the
caching:

* If using a cached `airshipctl`, the `airshipctl` binary must be stored in the
  `$CACHE_DIR/airshipctl/bin/` directory, and the developer must have set
  `USE_CACHED_ARTIFACTS` to `true`.
* If using a cached ephemeral iso, the iso must first be contained in a tarball named `iso.tar.gz`, must be stored in the
  `$CACHE_DIR/` directory, and the developer must have set
  `USE_CACHED_ISO` to `true`.
