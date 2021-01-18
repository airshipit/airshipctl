## airshipctl plan list

List plans

### Synopsis

List life-cycle plans which were defined in document model.


```
airshipctl plan list [flags]
```

### Examples

```

#list plan
airshipctl plan list

#list plan(yaml output format)
airshipctl plan list -o yaml

#list plan(table output format)
airshipctl plan list -o table
```

### Options

```
  -h, --help            help for list
  -o, --output string   'table' and 'yaml' are available output formats (default "table")
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl plan](airshipctl_plan.md)	 - Manage plans

