# Airship in a Pod

Airship in a pod is a Kubernetes pod definition which describes all of the
components required to deploy a fully functioning Airship 2 deployment. The pod
consists of the following "Task" containers:

* `artifact-setup`: This container collects the airshipctl binary repo, builds
  the airshipctl binary (and associated kustomize plugins), and makes them
  available to the other containers
* `infra-builder`: This container creates the various virtual networks and
  machines required for an Airship deployment
* `runner`: The runner container is the "meat" of the pod, and executes the
  deployment. It sets up a customized airshipctl config file, then uses
  airshipctl to pull the specified manifests and execute the deployment

The pod also contains the following "Support" containers:

* `libvirt`: This provides virtualisation
* `sushy-tools`: This is used for its BMC emulator
* `docker-in-docker`: This is used for nesting containers
* `nginx`: This is used for image hosting

## Azure Kubernetes Service (AKS) Quick Start

Airship-in-a-Pod can be easily run within AKS by running the script:

```
tools/airship-in-a-pod/scripts/aiap-in-aks.sh
```

Environment variables can be supplied to override default, such as:

* `AIAP_POD`: the kustomization to use for the AIAP Pod definition
* `CLEANUP_GROUP`: whether to delete the resource group created for
   AIAP.  Defaults to `false`.

Please consult the script for the full list of overrideable variables.

Note that authentication (e.g. `az login`) must be done prior to invoking
the script.

## Prerequisites

### Nested Virtualisation

If deployment is done on a VM, ensure that nested virtualisation is enabled.

### Environment variable setup

If you are within a proxy environment, ensure that the following environment
variables are defined, and NO_PROXY has the IP address which minikube uses.
Check the [minikube documentation](https://minikube.sigs.k8s.io/docs/commands/ip/)
for retrieving the minikube ip.

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
Refer to the [minikube documentation](https://minikube.sigs.k8s.io/docs/start/) for more details.

## Usage

Since Airship in a Pod is just a pod definition, deploying and using it is as
simple as deploying and using any other Kubernetes pod with `kubectl apply -f`.
The base pod definition can be found
[here](https://github.com/airshipit/airshipctl/tree/master/tools/airship-in-a-pod/examples/base)
and deploys using the current master `airshipctl` binary and current master
[test site](https://github.com/airshipit/airshipctl/tree/master/manifests/site/test-site).

### Pod configuration

Further configuration can be applied to the pod definition via
[`kustomize`](https://kustomize.io/). Options that can be configured can be
found in the [airshipctl example](https://github.com/airshipit/airshipctl/blob/master/tools/airship-in-a-pod/examples/airshipctl/replacements.yaml).
You may choose to either modify one of the examples or create your own.

Once you've created the desired configuration, the kustomized pod can be deployed with the following:

```
kustomize build ${PATH_TO_KUSTOMIZATION} | kubectl apply -f -

```

### Interacting with the Pod

For a quick rundown of what a particular container is doing, simply check the logs for that container.

```
# $CONTAINER is one of [runner infra-builder artifact-setup libvirt sushy dind nginx]
kubectl logs airship-in-a-pod -c $CONTAINER
```

For a deeper dive, consider `exec`ing into one of the containers.

```
# $CONTAINER is one of [runner infra-builder artifact-setup libvirt sushy dind nginx]
kubectl exec -it airship-in-a-pod -c $CONTAINER -- bash
```

#### Interacting with the Nodes

If you would like to interact with the nodes used in the deployment, you should
first prevent the runner container from exiting (check the examples/airshipctl
replacements for the option to do this). While the runner container is alive,
`exec` into it using the above command. The `kubectl` tool can then be used to
interact with a cluster. Choose a context from `kubectl config get-contexts`
and switch to it via `kubectl config use-context ${MY_CONTEXT}`.

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
  `USE_CACHED_AIRSHIPCTL` to `true`.
* If using a cached ephemeral iso, the iso must first be contained in a tarball named `iso.tar.gz`, must be stored in the
  `$CACHE_DIR/` directory, and the developer must have set
  `USE_CACHED_ISO` to `true`.
