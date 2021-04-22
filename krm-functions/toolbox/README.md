# Toolbox

This is KRM function written in `go` and uses the `kyaml` library for executing binaries inside container. It helps to run scripts in container as a `airshipctl` phase.
The toolbox image has pre-installed `sh` shell,`kubectl` and `calicoctl`.

## How to run your script as `airshipctl` phase

`NOTE`: All file paths in the following steps depend on the site you are working with and differ depending on the environment.

1. Create a phase document (kind: Phase)

        apiVersion: airshipit.org/v1alpha1
        kind: Phase
        metadata:
          name: kubectl-wait-node-ephemeral
          clusterName: ephemeral-cluster
        config:
          executorRef:
            apiVersion: airshipit.org/v1alpha1
            kind: GenericContainer
            name: kubectl-get-node

2. Create executor document (kind: GenericContainer). The [executor](https://github.com/airshipit/airshipctl/blob/master/manifests/phases/executors.yaml) use `configRef` to reference `ConfigMap` that will be generated using `configMapGenerator`. `configRef` must reference a Kubernetes ConfigMap with data key `script` with the script you want to execute. You can use kustomize [`configMapGenerator`](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/configmapgenerator/#configmap-from-file) to create ConfigMaps ([see example](https://github.com/airshipit/airshipctl/blob/master/manifests/function/phase-helpers/wait_node/kustomization.yaml)).

        apiVersion: airshipit.org/v1alpha1
        kind: GenericContainer
        metadata:
          name: kubectl-get-node
          labels:
            airshipit.org/deploy-k8s: "false"
        spec:
          type: krm
          image: quay.io/airshipit/toolbox:latest
          hostNetwork: true
          envVars
            MY_ENV # airshipctl will populate this value from your current env, you can pass credentials like this
            MY_ENV_TWO="my-value"
        configRef:
          kind: ConfigMap
          name: kubectl-get-node
          apiVersion: v1

3. Add your script as a ConfigMap. Scripts inside container have access to site kubeconfig in `${KUBECONFIG}` and to context of the cluster in `${KCTL_CONTEXT}` environment variables.

        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: kubectl-get-node
        data:
          script: |
            #!/bin/sh
            calicoctl apply --context ${KTCL_CONTEXT} -f $RENDERED_BUNDLE_PATH
            kubectl apply --context ${KTCL_CONTEXT} -f $RENDERED_BUNDLE_PATH

    1. add kustomize resources
    2. include them into PhaseConfigBundle

4. Make sure it is added to the bundle:
    1. `airshipctl phase render --source config -k ConfigMap` find your configmap in the output
    2. `airshipctl phase render --source config -k Phase` find your phase in output
    3. `airshipctl phase render --source config -k GenericContainer` find your executor in output

5) Run your phase:
`airshipctl phase run kubectl-wait-node-ephemeral`

## Input bundle usage

The KRM function writes to filesystem input bundle specified in `documentEntryPoint` in phase declaration and imports the path to this bundle in `RENDERED_BUNDLE_PATH` environment variable. For example it can be used with `calicoctl` as `calicoctl apply -f $RENDERED_BUNDLE_PATH`
Documents can be filtered by group, version and kind. You need to set `RESOURCE_GROUP_FILTER`, `RESOURCE_VERSION_FILTER` and/or`RESOURCE_KIND_FILTER` in executor definition to enable filtering.

## Important notes
1. The script must write to STDOUT valid yaml or redirect output to STDERR otherwise phase will fail with `mapping values are not allowed in this context`
2. All shell scripts must begin with `set -xe`. This allows errors to be passed from the container to the airshipctl itself. Without this flags the container will never fail.
