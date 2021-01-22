## airshipctl baremetal remotedirect

Bootstrap the ephemeral host

### Synopsis

Bootstrap the ephemeral host

```
airshipctl baremetal remotedirect [flags]
```

### Options

```
  -h, --help               help for remotedirect
      --iso-url string     specify iso url for host to boot from
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

