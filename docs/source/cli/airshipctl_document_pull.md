## airshipctl document pull

Airshipctl command to pull manifests from remote git repositories

### Synopsis

The remote manifests repositories as well as the target path where
the repositories will be cloned are defined in the airship config file.

By default the airship config file is initialized with the
repository "https://opendev.org/airship/treasuremap" as a source of
manifests and with the manifests target path "$HOME/.airship/default".


```
airshipctl document pull [flags]
```

### Examples

```

Pull manifests from remote repos
# airshipctl document pull
>>>>>>> Updating cmd files for documentation

```

### Options

```
  -h, --help          help for pull
  -n, --no-checkout   no checkout is performed after the clone is complete.
```

### Options inherited from parent commands

```
      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl document](airshipctl_document.md)	 - Airshipctl command to manage site manifest documents

