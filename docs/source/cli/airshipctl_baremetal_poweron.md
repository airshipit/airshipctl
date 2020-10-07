## airshipctl baremetal poweron

Power on a host

### Synopsis

Power on a host

```
airshipctl baremetal poweron [flags]
```

### Options

```
  -h, --help            help for poweron
  -l, --labels string   Label(s) to filter desired baremetal host documents
  -n, --name string     Name to filter desired baremetal host document
      --phase string    airshipctl phase that contains the desired baremetal host document(s) (default "bootstrap")
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl baremetal](airshipctl_baremetal.md)	 - Perform actions on baremetal hosts

