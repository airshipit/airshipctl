## airshipctl config get-management-config

View a management config or all management configs defined in the airshipctl config

### Synopsis

View a management config or all management configs defined in the airshipctl config

```
airshipctl config get-management-config [NAME] [flags]
```

### Examples

```

# View all defined management configurations
airshipctl config get-management-configs

# View a specific management configuration named "default"
airshipctl config get-management-config default

```

### Options

```
  -h, --help   help for get-management-config
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

