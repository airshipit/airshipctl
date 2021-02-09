## airshipctl baremetal reboot

Reboot a hosts

### Synopsis

Reboot baremetal hosts
The command will target baremetal hosts from airship inventory kustomize root
based on the --name, --namespace and --labels flags provided. If no flags are
provided airshipctl will try to select all baremetal hosts in the inventory


```
airshipctl baremetal reboot [flags]
```

### Examples

```
Perform action against hosts with name rdm9r3s3 in all namespaces where the host is found
# airshipctl baremetal reboot --name rdm9r3s3

Perform action against hosts with name rdm9r3s3 in namespace metal3
# airshipctl baremetal reboot --name rdm9r3s3 --namespace metal3

Perform action against all hosts defined in inventory
# airshipctl baremetal reboot --all

Perform action against hosts with a label 'foo=bar'
# airshipctl baremetal reboot --labels "foo=bar"

```

### Options

```
      --all                specify this to target all hosts in the inventory
  -h, --help               help for reboot
  -l, --labels string      Label(s) to filter desired baremetal host documents
      --name string        Name to filter desired baremetal host document
  -n, --namespace string   airshipctl phase that contains the desired baremetal host document(s)
      --timeout duration   timeout on baremetal action (default 10m0s)
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl baremetal](airshipctl_baremetal.md)	 - Perform actions on baremetal hosts

