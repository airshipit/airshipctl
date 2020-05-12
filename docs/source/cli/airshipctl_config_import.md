## airshipctl config import

Merge information from a kubernetes config file

### Synopsis

Merge the clusters, contexts, and users from an existing kubeConfig file into the airshipctl config file.


```
airshipctl config import <kubeConfig> [flags]
```

### Examples

```

# Import from a kubeConfig file"
airshipctl config import $HOME/.kube/config

```

### Options

```
  -h, --help   help for import
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

