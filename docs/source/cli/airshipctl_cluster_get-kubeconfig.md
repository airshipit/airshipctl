## airshipctl cluster get-kubeconfig

Airshipctl command to retrieve kubeconfig for a desired cluster

### Synopsis

Retrieves kubeconfig of the cluster and prints it to stdout.

If you specify CLUSTER_NAME, kubeconfig will have a CurrentContext set to CLUSTER_NAME and
will have its context defined.

If you don't specify CLUSTER_NAME, kubeconfig will have multiple contexts for every cluster
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
      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl cluster](airshipctl_cluster.md)	 - Airshipctl command to manage kubernetes clusters

