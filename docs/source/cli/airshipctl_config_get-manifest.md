## airshipctl config get-manifest

Get a manifest information from the airshipctl config

### Synopsis

Display a specific manifest information, or all defined manifests if no name is provided.


```
airshipctl config get-manifest NAME [flags]
```

### Examples

```

# List all the manifests airshipctl knows about
airshipctl config get-manifests

# Display a specific manifest
airshipctl config get-manifest e2e

```

### Options

```
  -h, --help   help for get-manifest
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

