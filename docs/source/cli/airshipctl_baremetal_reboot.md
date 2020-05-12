## airshipctl baremetal reboot

Reboot a host

### Synopsis

Reboot a host

```
airshipctl baremetal reboot [flags]
```

### Options

```
  -h, --help            help for reboot
  -l, --labels string   Label(s) to filter desired baremetal host documents
  -n, --name string     Name to filter desired baremetal host document
      --phase string    airshipctl phase that contains the desired baremetal host document(s) (default "bootstrap")
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl baremetal](airshipctl_baremetal.md)	 - Perform actions on baremetal hosts

