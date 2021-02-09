## airshipctl phase list

List phases

### Synopsis

List life-cycle phases which were defined in document model by group.
Phases within a group are executed sequentially. Multiple phase groups
are executed in parallel.


```
airshipctl phase list [flags]
```

### Options

```
  -c, --cluster-name string   filter documents by cluster name
  -h, --help                  help for list
      --plan string           Plan name of a plan
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl phase](airshipctl_phase.md)	 - Manage phases

