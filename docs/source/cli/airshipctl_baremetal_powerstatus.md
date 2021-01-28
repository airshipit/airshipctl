## airshipctl baremetal powerstatus

Retrieve the power status of a baremetal host

### Synopsis

Retrieve the power status of a baremetal host

```
airshipctl baremetal powerstatus [flags]
```

### Options

```
  -h, --help               help for powerstatus
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

