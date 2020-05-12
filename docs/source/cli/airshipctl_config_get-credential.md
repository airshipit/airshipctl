## airshipctl config get-credential

Get user credentials from the airshipctl config

### Synopsis

Display a specific user's credentials, or all defined user
credentials if no name is provided.


```
airshipctl config get-credential [NAME] [flags]
```

### Examples

```

# List all user credentials
airshipctl config get-credentials

# Display a specific user's credentials
airshipctl config get-credential exampleUser

```

### Options

```
  -h, --help   help for get-credential
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

