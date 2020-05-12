## airshipctl phase apply

Apply phase to a cluster

### Synopsis

Apply specific phase to kubernetes cluster such as control-plane, workloads, initinfra


```
airshipctl phase apply PHASE_NAME [flags]
```

### Examples

```

# Apply initinfra phase to a cluster
airshipctl phase apply initinfra

```

### Options

```
      --dry-run   don't deliver documents to the cluster, simulate the changes instead
  -h, --help      help for apply
      --prune     if set to true, command will delete all kubernetes resources that are not defined in airship documents and have airshipit.org/deployed=apply label
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl phase](airshipctl_phase.md)	 - Manage phases

