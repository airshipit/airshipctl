## airshipctl baremetal powerstatus

Airshipctl command to retrieve the power status of a bare metal host

### Synopsis

Retrieve the power status of a bare metal host. It targets a bare metal host from airship inventory
based on the --name, --namespace, --label and --timeout flags provided.


```
airshipctl baremetal powerstatus [flags]
```

### Examples

```

Perform action against host with name rdm9r3s3 in all namespaces where the host is found
# airshipctl baremetal powerstatus --name rdm9r3s3

Perform action against host with name rdm9r3s3 in namespace metal3
# airshipctl baremetal powerstatus --name rdm9r3s3 --namespace metal3

Perform action against host with a label 'foo=bar'
# airshipctl baremetal powerstatus --labels "foo=bar"

```

### Options

```
  -h, --help               help for powerstatus
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

