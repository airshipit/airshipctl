## airshipctl baremetal poweroff

Airshipctl command to shutdown bare metal host(s)

### Synopsis

Power off bare metal host(s). The command will target bare metal hosts from airship inventory based on the
--name, --namespace and --labels flags provided. If no flags are provided, airshipctl will select all bare metal hosts in the
inventory.


```
airshipctl baremetal poweroff [flags]
```

### Examples

```

Perform action against hosts with name rdm9r3s3 in all namespaces where the host is found
# airshipctl baremetal poweroff --name rdm9r3s3

Perform action against hosts with name rdm9r3s3 in namespace metal3
# airshipctl baremetal poweroff --name rdm9r3s3 --namespace metal3

Perform action against all hosts defined in inventory
# airshipctl baremetal poweroff --all

Perform action against hosts with a label 'foo=bar'
# airshipctl baremetal poweroff --labels "foo=bar"

```

### Options

```
      --all                specify this to target all hosts in the inventory
  -h, --help               help for poweroff
  -l, --labels string      label(s) to filter desired bare metal host documents
      --name string        name to filter desired bare metal host document
  -n, --namespace string   airshipctl phase that contains the desired bare metal host document(s)
      --timeout duration   timeout on bare metal action (default 10m0s)
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl baremetal](airshipctl_baremetal.md)	 - Airshipctl command to manage bare metal host(s)

