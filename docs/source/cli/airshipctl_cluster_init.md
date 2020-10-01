## airshipctl cluster init

Deploy cluster-api provider components

### Synopsis


Initialize cluster-api providers based on airshipctl document set.
document set must contain document of Kind: Clusterctl in phase initinfra.
Path to initinfra phase is defined in the initinfra phase document located
in the manifest repository.
Clusterctl document example:
---
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
    - "metal3:v0.3.1"
  control-plane-providers:
    - "kubeadm:v0.3.3"
providers:
  - name: "metal3"
    type: "InfrastructureProvider"
    versions:
      v0.3.1: manifests/function/capm3/v0.3.1
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


```
airshipctl cluster init [flags]
```

### Examples

```

# Initialize clusterctl providers and components
airshipctl cluster init

```

### Options

```
  -h, --help                help for init
      --kubeconfig string   Path to kubeconfig associated with cluster being managed
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl cluster](airshipctl_cluster.md)	 - Manage Kubernetes clusters

