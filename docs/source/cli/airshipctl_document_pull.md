## airshipctl document pull

Pulls documents from remote git repository

### Synopsis

The remote manifests repositories as well as the target path where
the repositories will be cloned are defined in the airship config file.

By default the airship config file is initialized with the
repository "https://opendev.org/airship/treasuremap" as a source of
manifests and with the manifests target path "$HOME/.airship/default".

```
airshipctl document pull [flags]
```

### Options

```
  -h, --help          help for pull
  -n, --no-checkout   No checkout is performed after the clone is complete.
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl document](airshipctl_document.md)	 - Manage deployment documents

