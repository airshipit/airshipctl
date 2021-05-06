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

To get power status of host with name rdm9r3s3 in all namespaces where the host is found
# airshipctl baremetal powerstatus --name rdm9r3s3

To get power status of host with name rdm9r3s3 in metal3 namespace
# airshipctl baremetal powerstatus --name rdm9r3s3 --namespace metal3

To get power status of host with a label 'foo=bar'
# airshipctl baremetal powerstatus --labels "foo=bar"

```

### Options

```
  -h, --help               help for powerstatus
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

