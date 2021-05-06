## airshipctl cluster list

Airshipctl command to get and list defined clusters

### Synopsis


Retrieve and list the defined clusters in the table form or display just the name as specified.


```
airshipctl cluster list [flags]
```

### Examples

```

Retrieve list of clusters
# airshipctl cluster list --airshipconf /tmp/airconfig
# airshipctl cluster list -o table
# airshipctl cluster list -o name

```

### Options

```
  -h, --help            help for list
  -o, --output string   output formats. Supported options are 'table' and 'name' (default "name")
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl cluster](airshipctl_cluster.md)	 - Airshipctl command to manage kubernetes clusters

