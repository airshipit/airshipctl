## airshipctl baremetal ejectmedia

Airshipctl command to eject virtual media attached to a bare metal host

### Synopsis

Eject virtual media attached to a bare metal host. The command will target bare metal hosts from airship site inventory based on the
--name, --namespace and --labels flags provided. If no flags are provided, airshipctl will select all bare metal hosts in the site
inventory.


```
airshipctl baremetal ejectmedia [flags]
```

### Examples

```

Perform ejectmedia action against hosts with name rdm9r3s3 in all namespaces where the host is found
# airshipctl baremetal ejectmedia --name rdm9r3s3

Perform ejectmedia action against hosts with name rdm9r3s3 in metal3 namespace
# airshipctl baremetal ejectmedia --name rdm9r3s3 --namespace metal3

Perform ejectmedia action against all hosts defined in inventory
# airshipctl baremetal ejectmedia --all

Perform ejectmedia action against hosts with a label 'foo=bar'
# airshipctl baremetal ejectmedia --labels "foo=bar"

```

### Options

```
      --all                specify this to target all hosts in the site inventory
  -h, --help               help for ejectmedia
  -l, --labels string      label(s) to filter desired bare metal host from site manifest documents
      --name string        name to filter desired bare metal host from site manifest document
  -n, --namespace string   airshipctl phase that contains the desired bare metal host from site manifest document(s)
      --timeout duration   timeout on bare metal action (default 10m0s)
```

### Options inherited from parent commands

```
      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl baremetal](airshipctl_baremetal.md)	 - Airshipctl command to manage bare metal host(s)

