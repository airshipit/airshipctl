## airshipctl config get-encryption-config

Get an encryption config information from the airshipctl config

### Synopsis

Display a specific encryption config information, or all defined encryption configs if no name is provided.


```
airshipctl config get-encryption-config NAME [flags]
```

### Examples

```

# List all the encryption configs airshipctl knows about
airshipctl config get-encryption-configs

# Display a specific encryption config
airshipctl config get-encryption-config exampleConfig

```

### Options

```
  -h, --help   help for get-encryption-config
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

