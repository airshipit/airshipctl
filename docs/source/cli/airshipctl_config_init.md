## airshipctl config init

Generate initial configuration files for airshipctl

### Synopsis

Generate an airshipctl config file and its associated kubeConfig file.
These files will be written to the $HOME/.airship directory, and will contain
default configurations.

NOTE: This will overwrite any existing config files in $HOME/.airship


```
airshipctl config init [flags]
```

### Options

```
  -h, --help   help for init
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

