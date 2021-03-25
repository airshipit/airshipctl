## airshipctl phase status

Status of the phase

### Synopsis

Status of the specific life-cycle phase such as ephemeral-control-plane, target-initinfra etc...

```
airshipctl phase status [flags]
```

### Examples

```

#Status of initinfra phase
airshipctl phase status ephemeral-control-plane

```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl phase](airshipctl_phase.md)	 - Manage phases

