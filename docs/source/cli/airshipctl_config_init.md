## airshipctl config init

Generate initial configuration file for airshipctl

### Synopsis

Generate an airshipctl config file. This file by default will be written to the $HOME/.airship directory,
and will contain default configuration. In case if flag --airshipconf provided - the file will be
written to the specified location instead. If a configuration file already exists at the specified path,
an error will be thrown; to overwrite it, specify the --overwrite flag.


```
airshipctl config init [flags]
```

### Examples

```

# Create new airshipctl config file at the default location
airshipctl config init

# Create new airshipctl config file at the custom location
airshipctl config init --airshipconf path/to/config

# Create new airshipctl config file at custom location and overwrite it
airshipctl config init --overwrite --airshipconf path/to/config

```

### Options

```
  -h, --help        help for init
      --overwrite   overwrite config file
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

