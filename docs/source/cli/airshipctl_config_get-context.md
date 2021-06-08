## airshipctl config get-context

Airshipctl command to get context(s) information from the airshipctl config

### Synopsis

Displays information about contexts such as associated manifests, users, and clusters. It would display a specific
context information, or all defined context information if no name is provided.


```
airshipctl config get-context CONTEXT_NAME [flags]
```

### Examples

```

List all contexts
# airshipctl config get-contexts

Display the current context
# airshipctl config get-context --current

Display a specific context
# airshipctl config get-context exampleContext

```

### Options

```
      --current       get the current context
      --format yaml   supported output format yaml or `table`, default is `yaml` (default "yaml")
  -h, --help          help for get-context
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Airshipctl command to manage airshipctl config file

