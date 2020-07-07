# Airship in a Pod

Airship in a pod is a Kubernetes pod definition which describes all of the
components required to deploy a fully functioning Airship 2 deployment. The pod
consists of the following "Task" containers:

* `airshipctl-builder`: This container builds the airshipctl binary and makes it
  available to the other containers
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

In order to deploy Airship in a Pod for development, you must first have a
working Kubernetes cluster. This guide assumes that a developer will deploy
using [minikube](https://minikube.sigs.k8s.io/docs/start/):

```
sudo -E minikube start --driver=none
```

## Usage

Since Airship in a Pod is just a pod definition, deploying and using it is as
simple as deploying and using any Kubernetes pod.

#### Deploy the Pod

```
kubectl apply -f airship-in-a-pod.yaml
```

#### View Pod Logs

```
kubectl logs airship-in-a-pod -c $CONTAINER
```

#### Interact with the Pod

```
kubectl exec -it airship-in-a-pod -c $CONTAINER -- bash
```

where `$CONTAINER` is one of the containers listed above.


### Output

Airship-in-a-pod produces the following outputs:

* The airshipctl repo and associated binary used with the deployment
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
