## airshipctl phase run

Run phase

### Synopsis

Run specific life-cycle phase such as ephemeral-control-plane, target-initinfra etc...

```
airshipctl phase run PHASE_NAME [flags]
```

### Examples

```

# Run initinfra phase
airshipctl phase run ephemeral-control-plane

```

### Options

```
      --dry-run   simulate phase execution
  -h, --help      help for run
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl phase](airshipctl_phase.md)	 - Manage phases

