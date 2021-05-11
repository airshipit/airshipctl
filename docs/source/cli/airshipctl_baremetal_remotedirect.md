## airshipctl baremetal remotedirect

Airshipctl command to bootstrap the ephemeral host

### Synopsis

Bootstrap bare metal host. It targets bare metal host from airship inventory based
on the --iso-url, --name, --namespace, --label and --timeout flags provided.


```
airshipctl baremetal remotedirect [flags]
```

### Examples

```

Perform action against hosts with name rdm9r3s3 in all namespaces where the host is found
# airshipctl baremetal remotedirect --name rdm9r3s3

Perform action against hosts with name rdm9r3s3 in namespace metal3
# airshipctl baremetal remotedirect --name rdm9r3s3 --namespace metal3

Perform action against hosts with a label 'foo=bar'
# airshipctl baremetal remotedirect --labels "foo=bar"

```

### Options

```
  -h, --help               help for remotedirect
      --iso-url string     specify iso url for host to boot from
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

