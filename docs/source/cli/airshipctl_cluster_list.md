## airshipctl cluster list

Retrieve the list of defined clusters

### Synopsis

Retrieve the list of defined clusters

```
airshipctl cluster list [flags]
```

### Examples

```
# Retrieve cluster list
airshipctl cluster list --airshipconf /tmp/airconfig
airshipctl cluster list -o table
airshipctl cluster list -o name

```

### Options

```
  -h, --help            help for list
  -o, --output string   'table' and 'name' are available output formats (default "name")
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl cluster](airshipctl_cluster.md)	 - Manage Kubernetes clusters

