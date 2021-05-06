## airshipctl cluster get-kubeconfig

Airshipctl command to retrieve kubeconfig for a desired cluster

### Synopsis

Retrieves kubeconfig of the cluster and prints it to stdout.

If you specify clusterName, kubeconfig will have a CurrentContext set to clusterName and
will have its context defined.

If you don't specify clusterName, kubeconfig will have multiple contexts for every cluster
in the airship site. Context names will correspond to cluster names. CurrentContext will be empty.


```
airshipctl cluster get-kubeconfig CLUSTER_NAME [flags]
```

### Examples

```

Retrieve target-cluster kubeconfig
# airshipctl cluster get-kubeconfig target-cluster

Retrieve kubeconfig for the entire site; the kubeconfig will have context for every cluster
# airshipctl cluster get-kubeconfig

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

* [airshipctl cluster](airshipctl_cluster.md)	 - Airshipctl command to manage kubernetes clusters

