## airshipctl cluster move

Move Cluster API objects, provider specific objects and all dependencies to the target cluster

### Synopsis

Move Cluster API objects, provider specific objects and all dependencies to the target cluster.

Note: The destination cluster MUST have the required provider components installed.


```
airshipctl cluster move [flags]
```

### Examples

```

Move Cluster API objects, provider specific objects and all dependencies to the target cluster.

  airshipctl cluster move --target-context <context name>

```

### Options

```
  -h, --help                    help for move
      --kubeconfig string       Path to kubeconfig associated with cluster being managed
      --target-context string   Context to be used within the kubeconfig file for the target cluster. If empty, current context will be used.
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl cluster](airshipctl_cluster.md)	 - Manage Kubernetes clusters

