## airshipctl config get-cluster

Get cluster information from the airshipctl config

### Synopsis

Display a specific cluster or all defined clusters if no name is provided.

Note that if a specific cluster's name is provided, the --cluster-type flag
must also be provided.
Valid values for the --cluster-type flag are [ephemeral|target].


```
airshipctl config get-cluster [NAME] [flags]
```

### Examples

```

# List all clusters
airshipctl config get-clusters

# Display a specific cluster
airshipctl config get-cluster --cluster-type=ephemeral exampleCluster

```

### Options

```
      --cluster-type string   type of the desired cluster
  -h, --help                  help for get-cluster
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

