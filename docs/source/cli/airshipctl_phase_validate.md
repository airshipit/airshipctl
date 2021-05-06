## airshipctl phase validate

Airshipctl command to validate phase and its documents

### Synopsis


Validates phase and its documents. To list the phases associated with a site, run 'airshipctl phase list'.


```
airshipctl phase validate PHASE_NAME [flags]
```

### Examples

```

To validate initinfra phase
# airshipctl phase validate initinfra

```

### Options

```
  -h, --help   help for validate
```

### Options inherited from parent commands

```
      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl phase](airshipctl_phase.md)	 - Airshipctl command to manage phases

