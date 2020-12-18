## airshipctl cluster get-kubeconfig

Retrieve kubeconfig for a desired cluster

### Synopsis

Retrieve cluster kubeconfig and save it to file or stdout.


```
airshipctl cluster get-kubeconfig [cluster_name] [flags]
```

### Examples

```
# Retrieve target-cluster kubeconfig and print it to stdout
airshipctl cluster get-kubeconfig target-cluster

```

### Options

```
  -h, --help   help for get-kubeconfig
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl cluster](airshipctl_cluster.md)	 - Manage Kubernetes clusters

