## airshipctl cluster get-kubeconfig

Retrieve kubeconfig for a desired cluster

### Synopsis

Retrieve cluster kubeconfig and print it to stdout


```
airshipctl cluster get-kubeconfig [cluster_name] [flags]
```

### Examples

```
# Retrieve target-cluster kubeconfig
airshipctl cluster get-kubeconfig target-cluster --kubeconfig /tmp/kubeconfig

```

### Options

```
      --context string      specify context within the kubeconfig file
  -h, --help                help for get-kubeconfig
      --kubeconfig string   path to kubeconfig associated with parental cluster
  -n, --namespace string    namespace where cluster is located, if not specified default one will be used (default "default")
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl cluster](airshipctl_cluster.md)	 - Manage Kubernetes clusters

