## airshipctl phase plan

List phases

### Synopsis

List life-cycle phases which were defined in document model by group.
Phases within a group are executed sequentially. Multiple phase groups
are executed in parallel.


```
airshipctl phase plan [flags]
```

### Options

```
  -h, --help   help for plan
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl phase](airshipctl_phase.md)	 - Manage phases

