## airshipctl config get-context

Get context information from the airshipctl config

### Synopsis

Display information about contexts such as associated manifests, users, and clusters.


```
airshipctl config get-context [NAME] [flags]
```

### Examples

```

# List all contexts
airshipctl config get-contexts

# Display the current context
airshipctl config get-context --current

# Display a specific context
airshipctl config get-context exampleContext

```

### Options

```
      --current   get the current context
  -h, --help      help for get-context
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

