## airshipctl config set-management-config

Airshipctl command to create/modify out-of-band management configuration in airshipctl config file

### Synopsis


Creates or modifies management config information based on the MGMT_CONFIG_NAME passed. The allowed set
of optional flags are management-type, system-action-retries and system-reboot-delay. Use --use-proxy
and --insecure to enable proxy and insecure options respectively.


```
airshipctl config set-management-config MGMT_CONFIG_NAME [flags]
```

### Examples

```

Create management configuration
# airshipctl config set-management-config default

Create or update management configuration named "default" with retry and to enable insecure options
# airshipctl config set-management-config default --insecure --system-action-retries 40

Enable proxy for "test" management configuration
# airshipctl config set-management-config test --use-proxy

```

### Options

```
  -h, --help                        help for set-management-config
      --insecure                    ignore SSL certificate verification on out-of-band management requests
      --management-type string      set the out-of-band management type (default "redfish")
      --system-action-retries int   set the number of attempts to poll a host for a status (default 30)
      --system-reboot-delay int     set the number of seconds to wait between power actions (e.g. shutdown, startup) (default 30)
      --use-proxy                   use the proxy configuration specified in the local environment (default true)
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Airshipctl command to manage airshipctl config file

