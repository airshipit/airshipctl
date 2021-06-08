## airshipctl config set-context

Airshipctl command to create/modify context in airshipctl config file

### Synopsis

Creates or modifies context in the airshipctl config file based on the CONTEXT_NAME passed or for the current context
if --current flag is specified. It accepts optional flags which include manifest name and management-config name.


```
airshipctl config set-context CONTEXT_NAME [flags]
```

### Examples

```

To create a new context named "exampleContext"
# airshipctl config set-context exampleContext --manifest=exampleManifest

To update the manifest of the current-context
# airshipctl config set-context --current --manifest=exampleManifest

```

### Options

```
      --current                    update the current context
  -h, --help                       help for set-context
      --management-config string   set the management config for the specified context
      --manifest string            set the manifest for the specified context
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Airshipctl command to manage airshipctl config file

