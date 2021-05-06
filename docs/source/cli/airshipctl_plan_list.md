## airshipctl plan list

Airshipctl command to list plans

### Synopsis

List plans defined in site manifest.


```
airshipctl plan list [flags]
```

### Examples

```

List plan
# airshipctl plan list

List plan(yaml output format)
# airshipctl plan list -o yaml

List plan(table output format)
# airshipctl plan list -o table
```

### Options

```
  -h, --help            help for list
  -o, --output string   output format. Supported formats are 'table' and 'yaml' (default "table")
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl plan](airshipctl_plan.md)	 - Airshipctl command to manage plans

