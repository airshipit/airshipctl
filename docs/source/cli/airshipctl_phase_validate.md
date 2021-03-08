## airshipctl phase validate

Assert that a phase is valid

### Synopsis

Command which would validate that the phase contains the required documents to run the phase.


```
airshipctl phase validate PHASE_NAME [flags]
```

### Examples

```

# validate initinfra phase
airshipctl phase validate initinfra

```

### Options

```
  -h, --help   help for validate
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl phase](airshipctl_phase.md)	 - Manage phases

