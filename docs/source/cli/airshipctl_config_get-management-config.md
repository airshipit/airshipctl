## airshipctl config get-management-config

Airshipctl command to view management config(s) defined in the airshipctl config

### Synopsis


Displays a specific management config information, or all defined management configs if no name is provided.
The information relates to reboot-delays and retry in seconds along with management-type that has to be used.


```
airshipctl config get-management-config MGMT_CONFIG_NAME [flags]
```

### Examples

```

View all management configurations
# airshipctl config get-management-configs

View a specific management configuration named "default"
# airshipctl config get-management-config default

```

### Options

```
  -h, --help   help for get-management-config
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Airshipctl command to manage airshipctl config file

