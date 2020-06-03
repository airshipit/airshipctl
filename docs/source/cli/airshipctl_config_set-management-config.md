## airshipctl config set-management-config

Modify an out-of-band management configuration

### Synopsis

Modify an out-of-band management configuration

```
airshipctl config set-management-config NAME [flags]
```

### Options

```
  -h, --help                     help for set-management-config
      --insecure                 Ignore SSL certificate verification on out-of-band management requests
      --management-type string   Set the out-of-band management type (default "redfish")
      --use-proxy                Use the proxy configuration specified in the local environment (default true)
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

