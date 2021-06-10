## airshipctl baremetal reboot

Airshipctl command to reboot host(s)

### Synopsis

Reboot bare metal host(s). The command will target bare metal hosts from airship site inventory based on the
--name, --namespace and --labels flags provided. If no flags are provided, airshipctl will select all bare metal hosts in the site
inventory.


```
airshipctl baremetal reboot [flags]
```

### Examples

```

Perform reboot action against hosts with name rdm9r3s3 in all namespaces where the host is found
# airshipctl baremetal reboot --name rdm9r3s3

Perform reboot action against hosts with name rdm9r3s3 in metal3 namespace
# airshipctl baremetal reboot --name rdm9r3s3 --namespace metal3

Perform reboot action against all hosts defined in inventory
# airshipctl baremetal reboot --all

Perform reboot action against hosts with a label 'foo=bar'
# airshipctl baremetal reboot --labels "foo=bar"

```

### Options

```
      --all                specify this to target all hosts in the site inventory
  -h, --help               help for reboot
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

