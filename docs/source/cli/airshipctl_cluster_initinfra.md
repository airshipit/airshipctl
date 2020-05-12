## airshipctl cluster initinfra

Deploy initinfra components to cluster

### Synopsis

Deploy initial infrastructure to kubernetes cluster such as
metal3.io, argo, tiller and other manifest documents with appropriate labels


```
airshipctl cluster initinfra [flags]
```

### Examples

```

# Deploy infrastructure to a cluster
airshipctl cluster initinfra

```

### Options

```
      --cluster-type string   select cluster type to deploy initial infrastructure to; currently only ephemeral is supported (default "ephemeral")
      --dry-run               don't deliver documents to the cluster, simulate the changes instead
  -h, --help                  help for initinfra
      --prune                 if set to true, command will delete all kubernetes resources that are not defined in airship documents and have airshipit.org/deployed=initinfra label
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl cluster](airshipctl_cluster.md)	 - Manage Kubernetes clusters

