## airshipctl config use-context

Switch to a different context

### Synopsis

Switch to a different context defined in the airshipctl config file.
This command doesn't change a context for the kubeconfig file.


```
airshipctl config use-context NAME [flags]
```

### Examples

```

# Switch to a context named "exampleContext" in airshipctl config file
airshipctl config use-context exampleContext

```

### Options

```
  -h, --help   help for use-context
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file

