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
* `status-checker`: This container is used to track the completion status of
  the task containers.

## Deployment Options

1. [Deploy on Azure (Quick Start)](#azure-kubernetes-service-aks-quick-start)
2. [Deploy using Minikube on Linux](#minikube-installation)

### Azure Kubernetes Service (AKS) Quick Start

Note: This section provides a means of very quickly getting up and running with
AIAP, but requires access to Azure. If you would like to deploy to a native
Linux environment, please refer to the [Minikube Installation](#minikube-installation).

Upon logging into and authenticating an Azure account (via `az login`),
Airship-in-a-Pod can be easily run within AKS by running the script:

```
tools/airship-in-a-pod/scripts/aiap-in-aks.sh
```

Environment variables can be supplied to override default, such as:

* `AIAP_POD`: the kustomization to use for the AIAP Pod definition
* `CLEANUP_GROUP`: whether to delete the resource group created for
   AIAP.  Defaults to `false`.

Please consult the script for the full list of overrideable variables.

### Minikube Installation

This sections provides instructions for deploying AIAP to a single-node minikube.

#### Prerequisites

* Nested Virtualisation: If deploying on a VM, ensure that nested virtualisation is enabled.
* Environment variable setup: If you are within a proxy environment, ensure that the following environment
variables are defined, and `NO_PROXY` has the IP address which minikube uses.
Check the [minikube documentation](https://minikube.sigs.k8s.io/docs/commands/ip/)
for retrieving the minikube ip.

```
export USE_PROXY=true
export HTTP_PROXY=http://username:password@host:port
export HTTPS_PROXY=http://username:password@host:port
export NO_PROXY="localhost,127.0.0.1,10.23.0.0/16,10.96.0.0/12,10.1.1.44"
export http_proxy="$HTTP_PROXY"
export https_proxy="$HTTPS_PROXY"
export no_proxy="$NO_PROXY"
```

#### To start minikube

Within the environment, with appropriate env variables set, run the following command.

```
sudo -E minikube start --driver=none
```

Refer to the [minikube documentation](https://minikube.sigs.k8s.io/docs/start/) for more details.

#### Deploy the Pod

Since Airship in a Pod is just a pod definition, deploying and using it is as
simple as deploying and using any other Kubernetes pod with `kubectl apply -f`.
The base pod definition can be found
[here](https://github.com/airshipit/airshipctl/tree/master/tools/airship-in-a-pod/examples/base)
and deploys using the current master `airshipctl` binary and current master
[test site](https://github.com/airshipit/airshipctl/tree/master/manifests/site/test-site).

#### Pod configuration

Further configuration can be applied to the pod definition via
[`kustomize`](https://kustomize.io/). Options that can be configured can be
found in the [airshipctl example](https://github.com/airshipit/airshipctl/blob/master/tools/airship-in-a-pod/examples/airshipctl/replacements.yaml).
You may choose to either modify one of the examples or create your own.

Once you've created the desired configuration, the kustomized pod can be deployed with the following:

```
kustomize build ${PATH_TO_KUSTOMIZATION} | kubectl apply -f -
```

## Finishing a Deployment

A deployment of Airship-in-a-pod is denoted by one of two states:

1. The runner container reaches the end of its execution successfully
2. An error occurs in any of the containers

The statuses for the task containers is aggregated in the `status-checker`
container, which provides a status report every 5 seconds. The status report
has the following structure:

```
artifact-setup: <$STATUS> infra-builder: <$STATUS> runner: <$STATUS>
```

In the above, `$STATUS` can be any of `RUNNING`, `SUCCESS`, `FAILED`, or
`UNKNOWN`. The last line of the `status-checker`'s logs will always contain the
most recent status report. This status report can be used to determine the
overall health of the deployment, as in the following:

```
# Check if AIAP has finished successfully
test $(kubectl logs airship-in-a-pod -c status-checker --tail 1 | grep -o "SUCCESS" | wc -l) = 3

# Check if AIAP has failed
kubectl logs airship-in-a-pod -c status-checker --tail 1 | grep -q "FAILED"
```

## Interacting with the Pod

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

## Interacting with the Nodes

If you would like to interact with the nodes used in the deployment, you should
first prevent the runner container from exiting (check the examples/airshipctl
replacements for the option to do this). While the runner container is alive,
`exec` into it using the above command. The `kubectl` tool can then be used to
interact with a cluster. Choose a context from `kubectl config get-contexts`
and switch to it via `kubectl config use-context ${MY_CONTEXT}`.

## Output

Airship-in-a-pod produces the following outputs:

* The airshipctl repo, manifest repo, and airshipctl binary used with the deployment.
* A tarball containing the generated ephemeral ISO, as well as the
  configuration used during generation.

These artifacts are placed at `ARTIFACTS_DIR` (defaults to `/opt/aiap-files/artifacts`).

## Caching
#TODO: Need to review this.

As it can be cumbersome and time-consuming to build and rebuild binaries and
images, some options are made available for caching. A developer may re-use
artifacts from previous runs (or provide their own) by placing them in
`CACHE_DIR` (defaults to `/opt/aiap-files/cache`). Special care is needed for the
caching:

* If using a cached `airshipctl`, the `airshipctl` binary must be stored in the
  `$CACHE_DIR/airshipctl/bin/` directory, and the developer must have set
  `USE_CACHED_AIRSHIPCTL` to `true`.
* If using a cached ephemeral iso, the iso must first be contained in a tarball named `iso.tar.gz`, must be stored in the
  `$CACHE_DIR/` directory, and the developer must have set
  `USE_CACHED_ISO` to `true`.
