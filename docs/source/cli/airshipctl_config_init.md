## airshipctl config init

Airshipctl command to generate initial configuration file for airshipctl

### Synopsis

Generates airshipctl config file. This file by default will be written to the $HOME/.airship directory,
and will contain default configuration. In case if flag --airshipconf provided - the default configuration
will be written to the file in the specified location instead. If a configuration file already exists
at the specified path, an error will be thrown; to overwrite it, specify the --overwrite flag.


```
airshipctl config init [flags]
```

### Examples

```

To create new airshipctl config file at the default location
# airshipctl config init

To create new airshipctl config file at the custom location
# airshipctl config init --airshipconf path/to/config

To create new airshipctl config file at the custom location and overwrite it
# airshipctl config init --overwrite --airshipconf path/to/config

```

### Options

```
  -h, --help        help for init
      --overwrite   overwrite config file
```

### Options inherited from parent commands

```
      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Airshipctl command to manage airshipctl config file

